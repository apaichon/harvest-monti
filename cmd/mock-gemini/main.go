// Command mock-gemini is the executable spec for the Gemini Live frame
// schema described in DES-0009 §2. It serves a WebSocket endpoint at
// /gemini that replays deterministic frame sequences from a fixtures
// directory.
//
// Wire shape (subset matched by internal/voice/providers/gemini/adapter.go):
//
//	upstream → server: {type: "open", session_id: "...", lang: "en-US"}
//	upstream → server: {type: "audio", pcm: "<base64>"}
//	upstream → server: {type: "function_result", call_id: "...", payload: {...}}
//	server → upstream: {type: "transcript", role: "user|assistant", text: "...", final: bool}
//	server → upstream: {type: "function_call", name: "...", call_id: "...", arguments: {...}}
//	server → upstream: {type: "audio", pcm: "<base64>", final: bool}
//	server → upstream: {type: "error", code: "...", message: "..."}
//	server → upstream: control close with code 1011 (synthesized for TC-8).
//
// Fixture format (e2e/fixtures/voice/<name>.cassette.json):
//
//	{
//	  "name": "happy-path",
//	  "frames": [
//	    {"after_ms": 50, "type": "transcript", "role": "user",
//	     "text": "add truffle pasta", "final": true},
//	    {"after_ms": 80, "type": "function_call",
//	     "name": "menu_search", "call_id": "fc1",
//	     "arguments": {"query": "truffle", "allergens": ["peanut"]}},
//	    ...
//	  ]
//	}
//
// Admin control endpoints:
//
//	POST /admin/inject/5xx        — next /gemini connect returns HTTP 503
//	POST /admin/inject/ws_drop    — next session closes with code 1011 after first frame
//	POST /admin/inject/silence    — next session withholds first frame for ?delay_ms=N
//	POST /admin/inject/quota      — next session returns control error code "quota_exceeded"
//	POST /admin/reset             — clears all injections
//	GET  /healthz                 — 200 OK
//
// The intent is that the same fixtures used by mock-gemini also drive the
// in-memory FakeTransport in e2e_test.go (dry-mode), so the cassettes are
// the single source of truth.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

// Frame is one entry in a cassette. Field names match DES-0009 §2 verbatim.
type Frame struct {
	AfterMS   int            `json:"after_ms"`
	Type      string         `json:"type"`
	Role      string         `json:"role,omitempty"`
	Text      string         `json:"text,omitempty"`
	Final     bool           `json:"final,omitempty"`
	Name      string         `json:"name,omitempty"`
	CallID    string         `json:"call_id,omitempty"`
	Arguments map[string]any `json:"arguments,omitempty"`
	PCM       string         `json:"pcm,omitempty"`
	Code      string         `json:"code,omitempty"`
	Message   string         `json:"message,omitempty"`
}

// Cassette is the JSON file shape.
type Cassette struct {
	Name   string  `json:"name"`
	Frames []Frame `json:"frames"`
}

// Inject is the mutable injection state. All four DES-0009 §5 triggers can be
// armed independently.
type Inject struct {
	mu         sync.Mutex
	fail5xx    bool
	wsDrop     bool
	silenceMS  int
	quota      bool
	once       bool // true → reset after first session
}

func (i *Inject) arm5xx()         { i.mu.Lock(); defer i.mu.Unlock(); i.fail5xx = true; i.once = true }
func (i *Inject) armWSDrop()      { i.mu.Lock(); defer i.mu.Unlock(); i.wsDrop = true; i.once = true }
func (i *Inject) armSilence(d int){ i.mu.Lock(); defer i.mu.Unlock(); i.silenceMS = d; i.once = true }
func (i *Inject) armQuota()       { i.mu.Lock(); defer i.mu.Unlock(); i.quota = true; i.once = true }
func (i *Inject) reset() {
	i.mu.Lock(); defer i.mu.Unlock()
	i.fail5xx, i.wsDrop, i.silenceMS, i.quota, i.once = false, false, 0, false, false
}
// InjectSnapshot is a lock-free, by-value view of the current trigger state.
// Returned by Inject.take() so callers don't copy the mutex.
type InjectSnapshot struct {
	Fail5xx   bool
	WSDrop    bool
	SilenceMS int
	Quota     bool
}

func (i *Inject) take() InjectSnapshot {
	i.mu.Lock(); defer i.mu.Unlock()
	out := InjectSnapshot{Fail5xx: i.fail5xx, WSDrop: i.wsDrop, SilenceMS: i.silenceMS, Quota: i.quota}
	if i.once {
		i.fail5xx, i.wsDrop, i.silenceMS, i.quota, i.once = false, false, 0, false, false
	}
	return out
}

// loadCassettes walks the fixtures dir and returns a name→cassette map.
func loadCassettes(dir string) (map[string]Cassette, error) {
	out := map[string]Cassette{}
	if dir == "" {
		return out, nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return out, err
	}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			log.Printf("mock-gemini: read %s: %v", e.Name(), err)
			continue
		}
		var c Cassette
		if err := json.Unmarshal(b, &c); err != nil {
			log.Printf("mock-gemini: parse %s: %v", e.Name(), err)
			continue
		}
		if c.Name == "" {
			c.Name = e.Name()
		}
		out[c.Name] = c
	}
	return out, nil
}

// server holds the running state.
type server struct {
	addr     string
	fixtures map[string]Cassette
	inj      *Inject
}

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprintln(w, "ok")
	})
	mux.HandleFunc("/gemini", s.handleGemini)
	mux.HandleFunc("/admin/inject/5xx", func(w http.ResponseWriter, _ *http.Request) {
		s.inj.arm5xx(); _, _ = fmt.Fprintln(w, "armed:5xx")
	})
	mux.HandleFunc("/admin/inject/ws_drop", func(w http.ResponseWriter, _ *http.Request) {
		s.inj.armWSDrop(); _, _ = fmt.Fprintln(w, "armed:ws_drop")
	})
	mux.HandleFunc("/admin/inject/silence", func(w http.ResponseWriter, r *http.Request) {
		d, _ := strconv.Atoi(r.URL.Query().Get("delay_ms"))
		if d == 0 {
			d = 2500
		}
		s.inj.armSilence(d)
		_, _ = fmt.Fprintf(w, "armed:silence:%dms\n", d)
	})
	mux.HandleFunc("/admin/inject/quota", func(w http.ResponseWriter, _ *http.Request) {
		s.inj.armQuota(); _, _ = fmt.Fprintln(w, "armed:quota")
	})
	mux.HandleFunc("/admin/reset", func(w http.ResponseWriter, _ *http.Request) {
		s.inj.reset(); _, _ = fmt.Fprintln(w, "reset")
	})
	mux.HandleFunc("/cassettes", func(w http.ResponseWriter, _ *http.Request) {
		names := []string{}
		for n := range s.fixtures {
			names = append(names, n)
		}
		_ = json.NewEncoder(w).Encode(names)
	})
	return mux
}

// handleGemini handles the upstream WebSocket-like channel. To stay
// dependency-light we accept plain HTTP long-poll style frames (POST with
// {cassette: "..."}) plus a streamed text/event-stream response. The Go
// adapter under internal/voice/providers/gemini speaks the same envelope
// through Transport so a real WS upgrade is not required for the contract
// test — the mock is the executable spec for the FRAME SHAPES, not the
// transport.
//
// Behaviour:
//   - POST /gemini?cassette=<name> with armed inj.fail5xx → 503
//   - With inj.quota → write a single error frame with code "quota_exceeded"
//   - With inj.silence_ms > 0 → sleep that long before first frame
//   - With inj.wsDrop → after first frame, close connection (simulating 1011)
//   - Otherwise: replay the named cassette frame-by-frame honouring AfterMS.
func (s *server) handleGemini(w http.ResponseWriter, r *http.Request) {
	inj := s.inj.take()
	if inj.Fail5xx {
		log.Printf("mock-gemini: injected 5xx")
		http.Error(w, "synthetic upstream 5xx", http.StatusServiceUnavailable)
		return
	}
	cassetteName := r.URL.Query().Get("cassette")
	if cassetteName == "" {
		cassetteName = "happy-path"
	}
	cass, ok := s.fixtures[cassetteName]
	if !ok {
		http.Error(w, "unknown cassette: "+cassetteName, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/x-ndjson")
	w.Header().Set("Cache-Control", "no-store")
	flusher, _ := w.(http.Flusher)
	enc := json.NewEncoder(w)

	if inj.Quota {
		_ = enc.Encode(Frame{Type: "error", Code: "quota_exceeded", Message: "tenant monthly quota reached"})
		if flusher != nil { flusher.Flush() }
		return
	}

	for i, fr := range cass.Frames {
		if i == 0 && inj.SilenceMS > 0 {
			time.Sleep(time.Duration(inj.SilenceMS) * time.Millisecond)
		}
		if fr.AfterMS > 0 {
			time.Sleep(time.Duration(fr.AfterMS) * time.Millisecond)
		}
		if err := enc.Encode(fr); err != nil {
			return
		}
		if flusher != nil {
			flusher.Flush()
		}
		if i == 0 && inj.WSDrop {
			log.Printf("mock-gemini: injected ws_drop after first frame")
			// Hijack to close abruptly — net/http will tear down the conn.
			if h, ok := w.(http.Hijacker); ok {
				conn, _, err := h.Hijack()
				if err == nil {
					_ = conn.Close()
				}
			}
			return
		}
	}
}

func main() {
	addr := flag.String("addr", envOr("MOCK_GEMINI_ADDR", ":7070"), "listen addr")
	fixtures := flag.String("fixtures", envOr("MOCK_GEMINI_FIXTURES", "e2e/fixtures/voice"), "fixtures dir")
	flag.Parse()

	cass, err := loadCassettes(*fixtures)
	if err != nil {
		log.Printf("mock-gemini: load cassettes: %v", err)
	}
	log.Printf("mock-gemini: loaded %d cassettes from %s", len(cass), *fixtures)

	srv := &server{addr: *addr, fixtures: cass, inj: &Inject{}}
	log.Printf("mock-gemini: listening on %s", *addr)
	if err := http.ListenAndServe(*addr, srv.routes()); err != nil {
		log.Fatal(err)
	}
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
