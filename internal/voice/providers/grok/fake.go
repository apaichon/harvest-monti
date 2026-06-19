package grok

import (
	"context"
	"sync"

	"github.com/apaichon/harvest-monti/internal/voice/provider"
)

// FakePipeline is an in-memory Pipeline for unit/contract tests.
type FakePipeline struct {
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

// NewFakePipeline builds a ready-to-use fake.
func NewFakePipeline() *FakePipeline {
	return &FakePipeline{
		audio: make(chan provider.AudioFrame, 16),
		text:  make(chan provider.Transcript, 16),
		fn:    make(chan provider.FunctionCall, 16),
	}
}

func (f *FakePipeline) Open(_ context.Context, cfg provider.SessionConfig) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.OpenCfg = cfg
	return f.OpenErr
}

func (f *FakePipeline) SendAudio(frame []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	cp := make([]byte, len(frame))
	copy(cp, frame)
	f.SentAudio = append(f.SentAudio, cp)
	return nil
}

func (f *FakePipeline) RegisterFunctions(decls []provider.FunctionDecl) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.RegisteredFuncs = append(f.RegisteredFuncs, decls...)
	return nil
}

func (f *FakePipeline) SwitchLanguage(bcp47 string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.LangCalls = append(f.LangCalls, bcp47)
	return nil
}

func (f *FakePipeline) SubmitFunctionResult(r provider.FunctionResult) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.FuncResults = append(f.FuncResults, r)
	return nil
}

func (f *FakePipeline) Receive() (<-chan provider.AudioFrame, <-chan provider.Transcript, <-chan provider.FunctionCall) {
	return f.audio, f.text, f.fn
}

func (f *FakePipeline) Close() error {
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

func (f *FakePipeline) PushAudio(a provider.AudioFrame)       { f.audio <- a }
func (f *FakePipeline) PushTranscript(t provider.Transcript)  { f.text <- t }
func (f *FakePipeline) PushFunctionCall(c provider.FunctionCall) { f.fn <- c }
