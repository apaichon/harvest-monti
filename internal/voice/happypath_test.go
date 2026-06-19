package voice_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/apaichon/harvest-monti/internal/voice/dispatch"
	"github.com/apaichon/harvest-monti/internal/voice/dispatch/ports"
	"github.com/apaichon/harvest-monti/internal/voice/provider"
	"github.com/apaichon/harvest-monti/internal/voice/providers/gemini"
	"github.com/apaichon/harvest-monti/internal/voice/providers/grok"
	"github.com/apaichon/harvest-monti/internal/voice/switcher"
)

// in-memory ports

type memMenu struct{ items []ports.MenuItemLite }

func (m *memMenu) Search(_ context.Context, q ports.MenuSearchQuery) ([]ports.MenuItemLite, error) {
	out := []ports.MenuItemLite{}
	for _, it := range m.items {
		if q.Query != "" && containsIgnoreCase(it.Name, q.Query) {
			out = append(out, it)
		}
	}
	return out, nil
}

type memCart struct {
	mu    sync.Mutex
	state ports.CartState
}

func (c *memCart) Add(_ context.Context, r ports.CartAddRequest) (ports.CartState, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.state.SessionID = r.SessionID
	c.state.Currency = "THB"
	c.state.Lines = append(c.state.Lines, ports.CartLine{
		LineID: "L1", ItemID: r.ItemID, Name: "Truffle Pasta", Qty: r.Qty, UnitCents: 38000,
	})
	c.state.Subtotal += 38000 * int64(r.Qty)
	return c.state, nil
}

func (c *memCart) Remove(_ context.Context, r ports.CartRemoveRequest) (ports.CartState, error) {
	return c.state, nil
}

type stubOrder struct{}

func (stubOrder) Submit(_ context.Context, r ports.OrderSubmitRequest) (ports.OrderTicket, error) {
	return ports.OrderTicket{OrderID: "o-1", SessionID: r.SessionID, PaymentStatus: "stubbed"}, nil
}

type stubStaff struct{}

func (stubStaff) Call(_ context.Context, r ports.StaffCallRequest) (ports.StaffPing, error) {
	return ports.StaffPing{PingID: "p-1", NotifiedAt: 1}, nil
}

type stubLang struct{}

func (stubLang) Switch(_ context.Context, r ports.LanguageSwitchRequest) (ports.LanguageState, error) {
	return ports.LanguageState{SessionID: r.SessionID, BCP47: r.BCP47}, nil
}

func containsIgnoreCase(s, sub string) bool {
	if len(sub) == 0 {
		return true
	}
	// trivial case-insensitive substring without importing strings to keep
	// the file dependency-light
	lower := func(b byte) byte {
		if b >= 'A' && b <= 'Z' {
			return b + 32
		}
		return b
	}
	for i := 0; i+len(sub) <= len(s); i++ {
		ok := true
		for j := 0; j < len(sub); j++ {
			if lower(s[i+j]) != lower(sub[j]) {
				ok = false
				break
			}
		}
		if ok {
			return true
		}
	}
	return false
}

// TestHappyPath_GeminiMocked exercises a recorded-fixture happy path:
// user says "add truffle pasta" → fake Gemini emits
// function_call menu_search → dispatcher → function_call cart_add → dispatcher →
// model emits final TTS frame. Verifies dispatcher injection, FIFO,
// SubmitFunctionResult delivery, and channel fan-out.
func TestHappyPath_GeminiMocked(t *testing.T) {
	menu := &memMenu{items: []ports.MenuItemLite{
		{ID: "i-truffle", Name: "Truffle Pasta", PriceCents: 38000, Currency: "THB"},
		{ID: "i-greencurry", Name: "Green Curry", PriceCents: 18000, Currency: "THB"},
	}}
	cart := &memCart{}
	d := dispatch.New(dispatch.Config{
		Menu:     menu,
		Cart:     cart,
		Order:    stubOrder{},
		Staff:    stubStaff{},
		Language: stubLang{},
	})

	geminiFake := gemini.NewFakeTransport()
	sw := switcher.New(switcher.Config{
		Primary:  &gemini.Adapter{NewTransport: func() gemini.Transport { return geminiFake }},
		Fallback: &grok.Adapter{NewPipeline: func() grok.Pipeline { return grok.NewFakePipeline() }},
	})

	id := dispatch.Identity{
		TenantID: "tenant-monti-demo", OutletID: "outlet-1",
		TableCode: "A12", SessionID: "sess-happy-1",
	}
	ps, err := sw.StartSession(context.Background(), provider.SessionConfig{
		TenantID: id.TenantID, OutletID: id.OutletID,
		TableCode: id.TableCode, SessionID: id.SessionID,
		LanguageHint: "en-US",
	})
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	sess := ps.(*switcher.Session)
	defer sess.Close(provider.CloseReason{Code: provider.CloseNormal})

	audioCh, textCh, fnCh := sess.ReceiveAudio()

	// Driver goroutine: emulate the gateway's pumpFunctionCalls loop.
	dispatched := make(chan provider.FunctionResult, 4)
	go func() {
		for call := range fnCh {
			res := d.Dispatch(context.Background(), id, call)
			_ = sess.SubmitFunctionResult(res)
			dispatched <- res
		}
	}()

	// Step 1: model emits the user's transcribed utterance.
	geminiFake.PushTranscript(provider.Transcript{
		Role: "user", Text: "add truffle pasta please", Final: true,
	})
	// Step 2: model emits menu_search function call.
	geminiFake.PushFunctionCall(provider.FunctionCall{
		Name:      dispatch.FnMenuSearch,
		CallID:    "fc1",
		Arguments: map[string]any{"query": "truffle"},
	})
	// Step 3: model emits cart_add function call.
	geminiFake.PushFunctionCall(provider.FunctionCall{
		Name:      dispatch.FnCartAdd,
		CallID:    "fc2",
		Arguments: map[string]any{"item_id": "i-truffle", "qty": 1},
	})
	// Step 4: model emits assistant transcript + TTS audio.
	geminiFake.PushTranscript(provider.Transcript{
		Role: "assistant", Text: "Added one truffle pasta.", Final: true,
	})
	geminiFake.PushAudio(provider.AudioFrame{
		PCM: make([]byte, 640), IsFinal: true,
	})

	// Drain expectations with a timeout.
	deadline := time.After(2 * time.Second)
	gotUserTranscript := false
	gotAssistantTranscript := false
	gotAudio := false
	dispatchedNames := []string{}

	for !(gotUserTranscript && gotAssistantTranscript && gotAudio && len(dispatchedNames) == 2) {
		select {
		case t1 := <-textCh:
			if t1.Role == "user" {
				gotUserTranscript = true
			}
			if t1.Role == "assistant" {
				gotAssistantTranscript = true
			}
		case <-audioCh:
			gotAudio = true
		case r := <-dispatched:
			if !r.Success {
				t.Fatalf("dispatch failed: %s", r.Error)
			}
			dispatchedNames = append(dispatchedNames, r.CallID)
		case <-deadline:
			t.Fatalf("timeout waiting; user=%v assistant=%v audio=%v dispatched=%v",
				gotUserTranscript, gotAssistantTranscript, gotAudio, dispatchedNames)
		}
	}

	// Final cart should have one Truffle Pasta line.
	if len(cart.state.Lines) != 1 {
		t.Errorf("expected 1 cart line, got %d", len(cart.state.Lines))
	}
	if cart.state.Lines[0].ItemID != "i-truffle" {
		t.Errorf("cart line item: %s", cart.state.Lines[0].ItemID)
	}
	if cart.state.SessionID != id.SessionID {
		t.Errorf("cart session_id: got %s want %s", cart.state.SessionID, id.SessionID)
	}

	// The model must have received the function results back (so it could
	// speak the answer).
	if got := len(geminiFake.FuncResults); got != 2 {
		t.Errorf("expected 2 SubmitFunctionResult calls, got %d", got)
	}
}
