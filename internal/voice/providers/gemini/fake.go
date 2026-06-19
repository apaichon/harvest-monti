package gemini

import (
	"context"
	"sync"

	"github.com/apaichon/harvest-monti/internal/voice/provider"
)

// FakeTransport is an in-memory Transport for unit/contract tests. It allows
// the test to push synthetic AudioFrame/Transcript/FunctionCall events and to
// observe what was sent upstream.
type FakeTransport struct {
	mu sync.Mutex

	OpenCfg provider.SessionConfig
	OpenErr error

	SentAudio       [][]byte
	RegisteredFuncs []provider.FunctionDecl
	LangCalls       []string
	FuncResults     []provider.FunctionResult
	Closed          bool

	audio chan provider.AudioFrame
	text  chan provider.Transcript
	fn    chan provider.FunctionCall
}

// NewFakeTransport builds a ready-to-use fake.
func NewFakeTransport() *FakeTransport {
	return &FakeTransport{
		audio: make(chan provider.AudioFrame, 16),
		text:  make(chan provider.Transcript, 16),
		fn:    make(chan provider.FunctionCall, 16),
	}
}

// Open records the session config.
func (f *FakeTransport) Open(_ context.Context, cfg provider.SessionConfig) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.OpenCfg = cfg
	return f.OpenErr
}

// SendAudio appends to the upstream-PCM record.
func (f *FakeTransport) SendAudio(frame []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	cp := make([]byte, len(frame))
	copy(cp, frame)
	f.SentAudio = append(f.SentAudio, cp)
	return nil
}

// RegisterFunctions records the declarations.
func (f *FakeTransport) RegisterFunctions(decls []provider.FunctionDecl) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.RegisteredFuncs = append(f.RegisteredFuncs, decls...)
	return nil
}

// SwitchLanguage records the BCP-47 code.
func (f *FakeTransport) SwitchLanguage(bcp47 string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.LangCalls = append(f.LangCalls, bcp47)
	return nil
}

// SubmitFunctionResult records the dispatched result.
func (f *FakeTransport) SubmitFunctionResult(r provider.FunctionResult) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.FuncResults = append(f.FuncResults, r)
	return nil
}

// Receive exposes the event channels.
func (f *FakeTransport) Receive() (<-chan provider.AudioFrame, <-chan provider.Transcript, <-chan provider.FunctionCall) {
	return f.audio, f.text, f.fn
}

// Close flips the Closed flag and closes the channels.
func (f *FakeTransport) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.Closed {
		return nil
	}
	f.Closed = true
	close(f.audio)
	close(f.text)
	close(f.fn)
	return nil
}

// PushAudio injects a model audio frame.
func (f *FakeTransport) PushAudio(a provider.AudioFrame) { f.audio <- a }

// PushTranscript injects a transcript turn.
func (f *FakeTransport) PushTranscript(t provider.Transcript) { f.text <- t }

// PushFunctionCall injects a model-emitted function call.
func (f *FakeTransport) PushFunctionCall(c provider.FunctionCall) { f.fn <- c }
