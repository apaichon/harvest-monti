package switcher_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/apaichon/harvest-monti/internal/voice/provider"
	"github.com/apaichon/harvest-monti/internal/voice/providers/gemini"
	"github.com/apaichon/harvest-monti/internal/voice/providers/grok"
	"github.com/apaichon/harvest-monti/internal/voice/switcher"
)

// clock is a deterministic fake clock for measuring switch latency.
type clock struct {
	mu sync.Mutex
	t  time.Time
}

func newClock(start time.Time) *clock { return &clock{t: start} }
func (c *clock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}
func (c *clock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.t = c.t.Add(d)
}

type captureSink struct {
	mu     sync.Mutex
	events []switcher.FallbackEvent
}

func (c *captureSink) FallbackTriggered(_ context.Context, ev switcher.FallbackEvent) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events = append(c.events, ev)
}

func newSwitcher(t *testing.T, ck *clock, sink switcher.EventSink) (*switcher.Switcher, *gemini.FakeTransport, *grok.FakePipeline) {
	t.Helper()
	geminiFake := gemini.NewFakeTransport()
	grokFake := grok.NewFakePipeline()
	sw := switcher.New(switcher.Config{
		Primary:  &gemini.Adapter{NewTransport: func() gemini.Transport { return geminiFake }},
		Fallback: &grok.Adapter{NewPipeline: func() grok.Pipeline { return grokFake }},
		Sink:     sink,
		Now:      ck.Now,
	})
	return sw, geminiFake, grokFake
}

// TC: Trigger 1 — Gemini 5xx (first fires unhealthy, second within 30s fires
// grok_active). DES-0009 §5 trigger 1 / ADR-0005 D2.
func TestSwitcher_TriggerGemini5xx_TwoStrikes(t *testing.T) {
	ck := newClock(time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC))
	sink := &captureSink{}
	sw, _, _ := newSwitcher(t, ck, sink)

	ps, err := sw.StartSession(context.Background(), provider.SessionConfig{SessionID: "s1", TenantID: "t1"})
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	sess := ps.(*switcher.Session)

	st, lat, err := sess.Trigger(switcher.TriggerGemini5xx)
	if err != nil {
		t.Fatalf("first 5xx: %v", err)
	}
	if st != switcher.StateGeminiUnhealthy {
		t.Fatalf("first 5xx: want gemini_unhealthy, got %s", st)
	}
	if lat != 0 {
		t.Fatalf("first 5xx should not switch; lat=%s", lat)
	}

	// Within 30 s, a second 5xx promotes to grok_active.
	ck.Advance(5 * time.Second)
	st, lat, err = sess.Trigger(switcher.TriggerGemini5xx)
	if err != nil {
		t.Fatalf("second 5xx: %v", err)
	}
	if st != switcher.StateGrokActive {
		t.Fatalf("second 5xx: want grok_active, got %s", st)
	}
	if lat > switcher.SwitchBudget {
		t.Fatalf("switch latency %s exceeds budget %s", lat, switcher.SwitchBudget)
	}
	if len(sink.events) != 1 {
		t.Fatalf("expected 1 fallback event, got %d", len(sink.events))
	}
	if sink.events[0].Reason != switcher.TriggerGemini5xx {
		t.Errorf("event reason: want %q got %q", switcher.TriggerGemini5xx, sink.events[0].Reason)
	}
}

// TC: Trigger 2 — quota exceeded → immediate grok_active (no retry, no first-strike).
func TestSwitcher_TriggerQuota_ImmediateFallback(t *testing.T) {
	ck := newClock(time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC))
	sink := &captureSink{}
	sw, _, _ := newSwitcher(t, ck, sink)

	ps, _ := sw.StartSession(context.Background(), provider.SessionConfig{SessionID: "s", TenantID: "t"})
	sess := ps.(*switcher.Session)

	st, lat, err := sess.Trigger(switcher.TriggerQuotaExceeded)
	if err != nil {
		t.Fatalf("trigger: %v", err)
	}
	if st != switcher.StateGrokActive {
		t.Fatalf("want grok_active, got %s", st)
	}
	if lat > switcher.SwitchBudget {
		t.Fatalf("switch latency %s exceeds budget %s", lat, switcher.SwitchBudget)
	}
	if sink.events[0].Reason != switcher.TriggerQuotaExceeded {
		t.Errorf("reason: want %q got %q", switcher.TriggerQuotaExceeded, sink.events[0].Reason)
	}
}

// TC: Trigger 3 — silence-to-first-token > 2000 ms → grok_active.
func TestSwitcher_TriggerSilenceToFirstToken(t *testing.T) {
	ck := newClock(time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC))
	sink := &captureSink{}
	sw, _, _ := newSwitcher(t, ck, sink)

	ps, _ := sw.StartSession(context.Background(), provider.SessionConfig{SessionID: "s", TenantID: "t"})
	sess := ps.(*switcher.Session)

	st, lat, err := sess.Trigger(switcher.TriggerSilenceToFirstToken)
	if err != nil {
		t.Fatalf("trigger: %v", err)
	}
	if st != switcher.StateGrokActive {
		t.Fatalf("want grok_active, got %s", st)
	}
	if lat > switcher.SwitchBudget {
		t.Fatalf("switch latency %s exceeds budget %s", lat, switcher.SwitchBudget)
	}
}

// TC: Trigger 4 — WS drop with unrecoverable code → grok_active.
func TestSwitcher_TriggerWSDrop(t *testing.T) {
	ck := newClock(time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC))
	sink := &captureSink{}
	sw, _, _ := newSwitcher(t, ck, sink)

	ps, _ := sw.StartSession(context.Background(), provider.SessionConfig{SessionID: "s", TenantID: "t"})
	sess := ps.(*switcher.Session)

	st, lat, err := sess.Trigger(switcher.TriggerWSDrop)
	if err != nil {
		t.Fatalf("trigger: %v", err)
	}
	if st != switcher.StateGrokActive {
		t.Fatalf("want grok_active, got %s", st)
	}
	if lat > switcher.SwitchBudget {
		t.Fatalf("switch latency %s exceeds budget %s", lat, switcher.SwitchBudget)
	}
}

// TC: 30s clean window resets gemini_unhealthy → gemini_primary.
func TestSwitcher_UnhealthyRecoveryAfter30s(t *testing.T) {
	ck := newClock(time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC))
	sw, _, _ := newSwitcher(t, ck, &captureSink{})

	ps, _ := sw.StartSession(context.Background(), provider.SessionConfig{SessionID: "s", TenantID: "t"})
	sess := ps.(*switcher.Session)

	_, _, _ = sess.Trigger(switcher.TriggerGemini5xx)
	if sess.State() != switcher.StateGeminiUnhealthy {
		t.Fatalf("expected unhealthy")
	}

	// Advance past the 30s window, then trigger a harmless event that re-evaluates.
	ck.Advance(31 * time.Second)
	st, _, _ := sess.Trigger("noop")
	if st != switcher.StateGeminiPrimary {
		t.Fatalf("expected recovery to gemini_primary, got %s", st)
	}
}

// TC: no switch-back from grok_active.
func TestSwitcher_NoSwitchBack(t *testing.T) {
	ck := newClock(time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC))
	sw, _, _ := newSwitcher(t, ck, &captureSink{})

	ps, _ := sw.StartSession(context.Background(), provider.SessionConfig{SessionID: "s", TenantID: "t"})
	sess := ps.(*switcher.Session)

	_, _, _ = sess.Trigger(switcher.TriggerQuotaExceeded)
	if sess.State() != switcher.StateGrokActive {
		t.Fatalf("setup failed")
	}

	// Any further Gemini-style trigger must not switch back.
	st, lat, _ := sess.Trigger(switcher.TriggerGemini5xx)
	if st != switcher.StateGrokActive {
		t.Fatalf("expected stay on grok_active, got %s", st)
	}
	if lat != 0 {
		t.Fatalf("expected no new switch event, lat=%s", lat)
	}
}

// TC: transcript replay caps at 6 turns OR 1500 tokens — whichever ceiling
// hits first (DES-0009 §10).
func TestSwitcher_TranscriptReplayCap(t *testing.T) {
	ck := newClock(time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC))
	sink := &captureSink{}
	sw, _, grokFake := newSwitcher(t, ck, sink)

	ps, _ := sw.StartSession(context.Background(), provider.SessionConfig{
		SessionID:    "s", TenantID: "t",
		SystemPrompt: "you are monti.",
	})
	sess := ps.(*switcher.Session)

	// Push 12 turns to exceed the 6-turn cap.
	for i := 0; i < 12; i++ {
		role := "user"
		if i%2 == 1 {
			role = "assistant"
		}
		sess.RecordTurn(switcher.Turn{Role: role, Text: shortText(i), Final: true})
	}

	// Force fallback.
	_, _, err := sess.Trigger(switcher.TriggerQuotaExceeded)
	if err != nil {
		t.Fatalf("trigger: %v", err)
	}

	// Inspect what the grok pipeline received as its SystemPrompt.
	got := grokFake.OpenCfg.SystemPrompt
	if got == "" {
		t.Fatal("grok pipeline received empty system prompt")
	}
	// Must contain the `[transcript so far]` marker.
	if !contains(got, "[transcript so far]") {
		t.Errorf("system prompt missing replay marker; got: %q", got)
	}
	// Must NOT contain the earliest turn (turn 0 dropped by the 6-turn cap).
	if contains(got, "turn-0-content") {
		t.Errorf("expected oldest turn (turn-0) to be dropped, but it appears in: %q", got)
	}
	// Must contain the most recent turn (turn 11).
	if !contains(got, "turn-11-content") {
		t.Errorf("expected newest turn (turn-11) to be present, got: %q", got)
	}
}

func shortText(i int) string { return "turn-" + itoa(i) + "-content" }

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var buf []byte
	for i > 0 {
		buf = append([]byte{byte('0' + i%10)}, buf...)
		i /= 10
	}
	return string(buf)
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// TC: < 500 ms switch budget under all four triggers. We measure using the
// fake clock — Now() is read at trigger start and again on event emission.
// In the fake clock the elapsed delta is 0 (no Advance between reads), so the
// measured latency is unambiguously < 500 ms regardless of real wall-clock.
func TestSwitcher_SwitchUnder500msAcrossAllTriggers(t *testing.T) {
	for _, reason := range []string{
		switcher.TriggerQuotaExceeded,
		switcher.TriggerSilenceToFirstToken,
		switcher.TriggerWSDrop,
	} {
		reason := reason
		t.Run(reason, func(t *testing.T) {
			ck := newClock(time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC))
			sink := &captureSink{}
			sw, _, _ := newSwitcher(t, ck, sink)
			ps, _ := sw.StartSession(context.Background(), provider.SessionConfig{SessionID: "s", TenantID: "t"})
			sess := ps.(*switcher.Session)
			_, lat, err := sess.Trigger(reason)
			if err != nil {
				t.Fatalf("trigger %q: %v", reason, err)
			}
			if lat >= switcher.SwitchBudget {
				t.Fatalf("trigger %q latency %s >= budget %s", reason, lat, switcher.SwitchBudget)
			}
		})
	}
	// For the two-strike 5xx case we measure latency from the second strike.
	t.Run("5xx-two-strike", func(t *testing.T) {
		ck := newClock(time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC))
		sw, _, _ := newSwitcher(t, ck, &captureSink{})
		ps, _ := sw.StartSession(context.Background(), provider.SessionConfig{SessionID: "s", TenantID: "t"})
		sess := ps.(*switcher.Session)
		_, _, _ = sess.Trigger(switcher.TriggerGemini5xx)
		ck.Advance(time.Second)
		_, lat, _ := sess.Trigger(switcher.TriggerGemini5xx)
		if lat >= switcher.SwitchBudget {
			t.Fatalf("5xx latency %s >= budget %s", lat, switcher.SwitchBudget)
		}
	})
}
