// Package gemini implements the VoiceProvider interface against Google's
// Gemini Live bidirectional audio API.
//
// SDK availability note (drift): the canonical client library
// google.golang.org/genai is not vendored in this worktree (offline build
// constraints). The adapter is structured so the WS transport is a swappable
// Transport interface — production wiring injects a Gemini-backed transport
// in cmd/api/main.go; tests inject an in-memory fake. This keeps the
// "no SDK imports outside internal/voice/providers/gemini" rule (ADR-0005 D3)
// trivially satisfied today while leaving the SDK seam ready.
package gemini

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/apaichon/harvest-monti/internal/voice/provider"
)

// Transport abstracts the bidirectional Gemini Live connection. A real
// implementation talks to wss://generativelanguage.googleapis.com/.../live;
// the test fake replays cassettes.
type Transport interface {
	// Open establishes the WS connection with the session config.
	Open(ctx context.Context, cfg provider.SessionConfig) error

	// SendAudio writes a PCM frame upstream.
	SendAudio(frame []byte) error

	// RegisterFunctions installs tool declarations.
	RegisterFunctions(decls []provider.FunctionDecl) error

	// SwitchLanguage signals an explicit language change.
	SwitchLanguage(bcp47 string) error

	// SubmitFunctionResult feeds a dispatched function result back to the model.
	SubmitFunctionResult(result provider.FunctionResult) error

	// Receive returns three demultiplexed event streams. The Transport closes
	// all three when the connection terminates.
	Receive() (<-chan provider.AudioFrame, <-chan provider.Transcript, <-chan provider.FunctionCall)

	// Close tears down the connection.
	Close() error
}

// Adapter is the VoiceProvider implementation for Gemini Live.
type Adapter struct {
	// NewTransport builds a fresh Transport per session.
	NewTransport func() Transport
	// FirstTokenTimeout is the silence-to-first-token budget enforced by
	// the Switcher (DES-0009 §5 trigger 3). The adapter exposes it for
	// transparency but does not act on it itself.
	FirstTokenTimeout time.Duration
}

// StartSession opens a Gemini Live session and returns a Session-shaped wrapper.
func (a *Adapter) StartSession(ctx context.Context, cfg provider.SessionConfig) (provider.Session, error) {
	if a.NewTransport == nil {
		return nil, errors.New("gemini: NewTransport not configured")
	}
	t := a.NewTransport()
	if err := t.Open(ctx, cfg); err != nil {
		return nil, err
	}
	return &session{
		cfg: cfg,
		t:   t,
	}, nil
}

type session struct {
	cfg provider.SessionConfig
	t   Transport

	closeOnce sync.Once
}

func (s *session) SendAudio(frame []byte) error {
	return s.t.SendAudio(frame)
}

func (s *session) ReceiveAudio() (<-chan provider.AudioFrame, <-chan provider.Transcript, <-chan provider.FunctionCall) {
	return s.t.Receive()
}

func (s *session) RegisterFunctions(decls []provider.FunctionDecl) error {
	return s.t.RegisterFunctions(decls)
}

func (s *session) SwitchLanguage(bcp47 string) error {
	return s.t.SwitchLanguage(bcp47)
}

func (s *session) SubmitFunctionResult(result provider.FunctionResult) error {
	return s.t.SubmitFunctionResult(result)
}

func (s *session) Close(reason provider.CloseReason) error {
	var err error
	s.closeOnce.Do(func() {
		err = s.t.Close()
	})
	return err
}
