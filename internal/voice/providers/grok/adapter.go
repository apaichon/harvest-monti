// Package grok wraps the SPRINT-0002 TASK-0004 Grok + Deepgram + ElevenLabs
// pipeline behind the VoiceProvider interface.
//
// Drift note: TASK-0004 has not yet merged into develop. This adapter ships
// in two modes:
//
//   - Stub (default when no PipelineFactory is wired): every Session.method
//     returns provider.ErrNotImplemented or its method equivalent so the
//     Switcher can fall back to "dead" cleanly during integration runs.
//   - Real (wired in cmd/api/main.go once TASK-0004 lands): inject a
//     PipelineFactory that builds a Pipeline talking to the live Deepgram,
//     Grok and ElevenLabs HTTP/WS endpoints. The Pipeline interface mirrors
//     the gemini.Transport seam, which keeps the Switcher contract identical
//     across both providers and lets the TEST-0009 fallback cases run
//     against an in-memory fake (see fake.go) today.
package grok

import (
	"context"
	"errors"
	"sync"

	"github.com/apaichon/harvest-monti/internal/voice/provider"
)

// ErrNotImplemented is returned by the stub adapter when no Pipeline is wired.
var ErrNotImplemented = errors.New("grok: pipeline not implemented (TASK-0004 pending)")

// Pipeline is the transport seam for the three-vendor pipe. The same shape as
// gemini.Transport so tests can share helpers.
type Pipeline interface {
	Open(ctx context.Context, cfg provider.SessionConfig) error
	SendAudio(frame []byte) error
	RegisterFunctions(decls []provider.FunctionDecl) error
	SwitchLanguage(bcp47 string) error
	SubmitFunctionResult(r provider.FunctionResult) error
	Receive() (<-chan provider.AudioFrame, <-chan provider.Transcript, <-chan provider.FunctionCall)
	Close() error
}

// Adapter is the VoiceProvider implementation for the Grok pipeline.
type Adapter struct {
	// NewPipeline builds a fresh Pipeline per session. When nil the adapter
	// behaves as the documented stub: StartSession returns ErrNotImplemented.
	NewPipeline func() Pipeline
}

// StartSession opens a pipeline session.
func (a *Adapter) StartSession(ctx context.Context, cfg provider.SessionConfig) (provider.Session, error) {
	if a.NewPipeline == nil {
		return nil, ErrNotImplemented
	}
	p := a.NewPipeline()
	if err := p.Open(ctx, cfg); err != nil {
		return nil, err
	}
	return &session{cfg: cfg, p: p}, nil
}

type session struct {
	cfg       provider.SessionConfig
	p         Pipeline
	closeOnce sync.Once
}

func (s *session) SendAudio(frame []byte) error { return s.p.SendAudio(frame) }

func (s *session) ReceiveAudio() (<-chan provider.AudioFrame, <-chan provider.Transcript, <-chan provider.FunctionCall) {
	return s.p.Receive()
}

func (s *session) RegisterFunctions(decls []provider.FunctionDecl) error {
	return s.p.RegisterFunctions(decls)
}

func (s *session) SwitchLanguage(bcp47 string) error { return s.p.SwitchLanguage(bcp47) }

func (s *session) SubmitFunctionResult(r provider.FunctionResult) error {
	return s.p.SubmitFunctionResult(r)
}

func (s *session) Close(reason provider.CloseReason) error {
	var err error
	s.closeOnce.Do(func() {
		err = s.p.Close()
	})
	return err
}
