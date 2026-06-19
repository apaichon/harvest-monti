// Package switcher is the VoiceProvider Switcher decided in ADR-0005 D3 and
// concretized in DES-0009 §5. It owns the four trigger conditions, the
// state machine, the < 500 ms switch budget, and transcript replay (capped at
// 6 turns OR 1500 tokens — whichever ceiling hits first per DES-0009 §10).
package switcher

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/apaichon/harvest-monti/internal/voice/provider"
)

// State enumerates the switcher's four states (DES-0009 §5).
type State int

const (
	StateGeminiPrimary State = iota
	StateGeminiUnhealthy
	StateGrokActive
	StateDead
)

// String implements fmt.Stringer.
func (s State) String() string {
	switch s {
	case StateGeminiPrimary:
		return "gemini_primary"
	case StateGeminiUnhealthy:
		return "gemini_unhealthy"
	case StateGrokActive:
		return "grok_active"
	case StateDead:
		return "dead"
	default:
		return "unknown"
	}
}

// Trigger names — the four canonical triggers from DES-0009 §5 / ADR-0005 D2.
const (
	TriggerGemini5xx              = "5xx"
	TriggerQuotaExceeded          = "quota"
	TriggerSilenceToFirstToken    = "silence_to_first_token"
	TriggerWSDrop                 = "ws_drop"
	TriggerGrokUnrecoverable      = "grok_unrecoverable"
)

// UnhealthyWindow is the rolling window in which a second Gemini 5xx promotes
// gemini_unhealthy → grok_active.
const UnhealthyWindow = 30 * time.Second

// SwitchBudget is the latency target DES-0009 §5 sets for the trigger →
// first-Grok-audio-frame path.
const SwitchBudget = 500 * time.Millisecond

// Defaults for transcript replay caps (DES-0009 §10).
const (
	DefaultReplayTurnCap  = 6
	DefaultReplayTokenCap = 1500
)

// Turn is a recorded transcript turn carried across a fallback.
type Turn struct {
	Role  string
	Text  string
	Final bool
}

// EventSink receives switcher observability events. Wire to NATS in production.
type EventSink interface {
	FallbackTriggered(ctx context.Context, ev FallbackEvent)
}

// FallbackEvent is the payload of monti.voice.fallback.triggered.
type FallbackEvent struct {
	SessionID       string
	Reason          string
	From            string
	To              string
	SwitchLatencyMS int64
	OccurredAt      time.Time
}

// nopEventSink swallows events when no sink is wired.
type nopEventSink struct{}

func (nopEventSink) FallbackTriggered(context.Context, FallbackEvent) {}

// Config wires the Switcher.
type Config struct {
	Primary  provider.VoiceProvider
	Fallback provider.VoiceProvider
	Sink     EventSink

	ReplayTurnCap  int
	ReplayTokenCap int

	Logger *slog.Logger
	Now    func() time.Time
}

// Switcher implements provider.VoiceProvider by holding both adapters and
// applying DES-0009 §5 trigger logic on a per-session basis.
type Switcher struct {
	cfg Config
}

// New builds a Switcher.
func New(cfg Config) *Switcher {
	if cfg.ReplayTurnCap == 0 {
		cfg.ReplayTurnCap = DefaultReplayTurnCap
	}
	if cfg.ReplayTokenCap == 0 {
		cfg.ReplayTokenCap = DefaultReplayTokenCap
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	if cfg.Sink == nil {
		cfg.Sink = nopEventSink{}
	}
	return &Switcher{cfg: cfg}
}

// StartSession honours provider.VoiceProvider by returning a Session-shaped
// wrapper that owns the per-session state machine.
func (s *Switcher) StartSession(ctx context.Context, cfg provider.SessionConfig) (provider.Session, error) {
	sess := &Session{
		ctx:    ctx,
		cfg:    cfg,
		parent: s,
		state:  StateGeminiPrimary,
	}
	if err := sess.openPrimary(ctx); err != nil {
		// If Gemini won't even start, immediately try fallback once.
		sess.parent.cfg.Logger.Warn("switcher.primary_start_failed",
			"session_id", cfg.SessionID, "err", err.Error())
		if ferr := sess.openFallback(ctx, TriggerGemini5xx); ferr != nil {
			return nil, fmt.Errorf("switcher: both providers failed to start: primary=%v fallback=%v", err, ferr)
		}
	}
	return sess, nil
}

// Session is the per-WS state machine.
type Session struct {
	ctx    context.Context
	cfg    provider.SessionConfig
	parent *Switcher

	mu         sync.Mutex
	state      State
	active     provider.Session
	transcript []Turn
	lastFail   time.Time // for the 30s gemini_unhealthy window

	// Output fan-out: callers Receive once and stay attached across switches.
	audioOut    chan provider.AudioFrame
	textOut     chan provider.Transcript
	funcOut     chan provider.FunctionCall
	closed      bool
	closeOnce   sync.Once
	pumpCancel  context.CancelFunc
	pumpDone    chan struct{}
	receiveOnce sync.Once
}

func (s *Session) openPrimary(ctx context.Context) error {
	if s.parent.cfg.Primary == nil {
		return errors.New("switcher: primary provider not configured")
	}
	sess, err := s.parent.cfg.Primary.StartSession(ctx, s.cfg)
	if err != nil {
		return err
	}
	s.active = sess
	s.state = StateGeminiPrimary
	return nil
}

func (s *Session) openFallback(ctx context.Context, reason string) error {
	if s.parent.cfg.Fallback == nil {
		s.state = StateDead
		return errors.New("switcher: fallback provider not configured")
	}
	// Replay capped transcript as a system-prompt prefix per DES-0009 §10.
	replay := capTurns(s.transcript, s.parent.cfg.ReplayTurnCap, s.parent.cfg.ReplayTokenCap)
	cfg := s.cfg
	cfg.SystemPrompt = injectReplay(cfg.SystemPrompt, replay)
	sess, err := s.parent.cfg.Fallback.StartSession(ctx, cfg)
	if err != nil {
		s.state = StateDead
		return err
	}
	s.active = sess
	s.state = StateGrokActive
	s.parent.cfg.Logger.Info("switcher.fallback_engaged",
		"session_id", s.cfg.SessionID,
		"reason", reason,
		"replay_turns", len(replay),
	)
	return nil
}

// Trigger advances the state machine. Returns the resulting state and the
// measured switch latency (zero if no switch occurred).
func (s *Session) Trigger(reason string) (State, time.Duration, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	start := s.parent.cfg.Now()
	from := s.state
	var switched bool

	switch s.state {
	case StateGeminiPrimary:
		switch reason {
		case TriggerGemini5xx:
			// First 5xx → unhealthy (transient); second within 30s → grok_active.
			s.state = StateGeminiUnhealthy
			s.lastFail = start
		case TriggerQuotaExceeded, TriggerSilenceToFirstToken, TriggerWSDrop:
			if err := s.openFallback(s.ctx, reason); err != nil {
				return s.state, 0, err
			}
			switched = true
		}
	case StateGeminiUnhealthy:
		switch reason {
		case TriggerGemini5xx:
			if start.Sub(s.lastFail) <= UnhealthyWindow {
				if err := s.openFallback(s.ctx, reason); err != nil {
					return s.state, 0, err
				}
				switched = true
			} else {
				// Window elapsed — treat as first 5xx again.
				s.lastFail = start
			}
		case TriggerQuotaExceeded, TriggerSilenceToFirstToken, TriggerWSDrop:
			if err := s.openFallback(s.ctx, reason); err != nil {
				return s.state, 0, err
			}
			switched = true
		}
	case StateGrokActive:
		if reason == TriggerGrokUnrecoverable {
			s.state = StateDead
		}
		// All Gemini triggers are no-ops once Grok is active (no switch-back per DES-0009 §5).
	case StateDead:
		// Terminal.
	}

	// Health-recovery: 30s clean window resets gemini_unhealthy → gemini_primary.
	if s.state == StateGeminiUnhealthy && start.Sub(s.lastFail) > UnhealthyWindow {
		s.state = StateGeminiPrimary
	}

	var latency time.Duration
	if switched {
		latency = s.parent.cfg.Now().Sub(start)
		s.parent.cfg.Sink.FallbackTriggered(s.ctx, FallbackEvent{
			SessionID:       s.cfg.SessionID,
			Reason:          reason,
			From:            from.String(),
			To:              s.state.String(),
			SwitchLatencyMS: latency.Milliseconds(),
			OccurredAt:      start,
		})
		if latency > SwitchBudget {
			s.parent.cfg.Logger.Warn("switcher.budget_exceeded",
				"session_id", s.cfg.SessionID,
				"latency_ms", latency.Milliseconds(),
				"budget_ms", SwitchBudget.Milliseconds(),
				"reason", reason,
			)
		}
	}

	return s.state, latency, nil
}

// State returns the current state (test helper).
func (s *Session) State() State {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state
}

// RecordTurn appends a final transcript turn for replay purposes.
func (s *Session) RecordTurn(t Turn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !t.Final {
		return
	}
	s.transcript = append(s.transcript, t)
}

// Active returns the underlying provider Session (test helper).
func (s *Session) Active() provider.Session {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.active
}

// SendAudio implements provider.Session.
func (s *Session) SendAudio(frame []byte) error {
	s.mu.Lock()
	a := s.active
	s.mu.Unlock()
	if a == nil {
		return errors.New("switcher: no active session")
	}
	return a.SendAudio(frame)
}

// ReceiveAudio implements provider.Session by fanning the active adapter's
// channels through stable proxies that survive a fallback swap.
func (s *Session) ReceiveAudio() (<-chan provider.AudioFrame, <-chan provider.Transcript, <-chan provider.FunctionCall) {
	s.receiveOnce.Do(func() {
		s.audioOut = make(chan provider.AudioFrame, 32)
		s.textOut = make(chan provider.Transcript, 32)
		s.funcOut = make(chan provider.FunctionCall, 8)
		ctx, cancel := context.WithCancel(s.ctx)
		s.pumpCancel = cancel
		s.pumpDone = make(chan struct{})
		go s.pump(ctx)
	})
	return s.audioOut, s.textOut, s.funcOut
}

func (s *Session) pump(ctx context.Context) {
	defer close(s.pumpDone)
	for {
		s.mu.Lock()
		a := s.active
		s.mu.Unlock()
		if a == nil {
			return
		}
		aud, txt, fn := a.ReceiveAudio()
		for {
			select {
			case <-ctx.Done():
				return
			case f, ok := <-aud:
				if !ok {
					aud = nil
				} else {
					s.audioOut <- f
				}
			case t, ok := <-txt:
				if !ok {
					txt = nil
				} else {
					if t.Final {
						s.RecordTurn(Turn{Role: t.Role, Text: t.Text, Final: true})
					}
					s.textOut <- t
				}
			case c, ok := <-fn:
				if !ok {
					fn = nil
				} else {
					s.funcOut <- c
				}
			}
			// If the active session changed under us, break to rebind.
			s.mu.Lock()
			same := s.active == a
			s.mu.Unlock()
			if !same {
				break
			}
			if aud == nil && txt == nil && fn == nil {
				return
			}
		}
	}
}

// RegisterFunctions implements provider.Session.
func (s *Session) RegisterFunctions(decls []provider.FunctionDecl) error {
	s.mu.Lock()
	a := s.active
	s.mu.Unlock()
	if a == nil {
		return errors.New("switcher: no active session")
	}
	return a.RegisterFunctions(decls)
}

// SwitchLanguage implements provider.Session.
func (s *Session) SwitchLanguage(bcp47 string) error {
	s.mu.Lock()
	a := s.active
	s.mu.Unlock()
	if a == nil {
		return errors.New("switcher: no active session")
	}
	return a.SwitchLanguage(bcp47)
}

// SubmitFunctionResult implements provider.Session.
func (s *Session) SubmitFunctionResult(result provider.FunctionResult) error {
	s.mu.Lock()
	a := s.active
	s.mu.Unlock()
	if a == nil {
		return errors.New("switcher: no active session")
	}
	return a.SubmitFunctionResult(result)
}

// Close implements provider.Session.
func (s *Session) Close(reason provider.CloseReason) error {
	var err error
	s.closeOnce.Do(func() {
		s.mu.Lock()
		a := s.active
		s.closed = true
		s.mu.Unlock()
		if a != nil {
			err = a.Close(reason)
		}
		if s.pumpCancel != nil {
			s.pumpCancel()
		}
	})
	return err
}

// CurrentProvider returns "gemini" or "grok" for the provider that is
// currently bound to this session.
func (s *Session) CurrentProvider() string {
	switch s.State() {
	case StateGeminiPrimary, StateGeminiUnhealthy:
		return "gemini"
	case StateGrokActive:
		return "grok"
	default:
		return "none"
	}
}

// ---------- replay helpers ----------

// capTurns implements DES-0009 §10: keep at most N turns OR M tokens,
// whichever ceiling hits first.
func capTurns(turns []Turn, turnCap, tokenCap int) []Turn {
	if len(turns) == 0 {
		return nil
	}
	// Walk back from the end.
	out := make([]Turn, 0, turnCap)
	tokens := 0
	for i := len(turns) - 1; i >= 0; i-- {
		t := turns[i]
		tCount := estimateTokens(t.Text)
		if len(out) >= turnCap || tokens+tCount > tokenCap {
			break
		}
		out = append([]Turn{t}, out...)
		tokens += tCount
	}
	return out
}

// estimateTokens is a deliberately cheap stand-in for a real tokenizer; the
// quota numbers in DES-0009 §10 are order-of-magnitude, so a ~4 chars/token
// approximation is sufficient for the replay cap calculation.
func estimateTokens(text string) int {
	if text == "" {
		return 0
	}
	n := (len(text) + 3) / 4
	if n == 0 {
		return 1
	}
	return n
}

// injectReplay weaves the capped transcript into the system prompt with the
// `[transcript so far]` marker required by DES-0009 §10.
func injectReplay(systemPrompt string, replay []Turn) string {
	if len(replay) == 0 {
		return systemPrompt
	}
	var b strings.Builder
	b.WriteString(systemPrompt)
	if !strings.HasSuffix(systemPrompt, "\n") {
		b.WriteString("\n\n")
	}
	b.WriteString("[transcript so far]\n")
	for _, t := range replay {
		b.WriteString(t.Role)
		b.WriteString(": ")
		b.WriteString(t.Text)
		b.WriteByte('\n')
	}
	return b.String()
}
