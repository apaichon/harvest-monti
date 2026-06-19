// Package archive implements the MinIO session archiver from DES-0007 §4
// and ADR-0005 resolved-ambiguities: one .opus audio file + one
// .transcript.jsonl per session under
// monti-voice-sessions/{tenant_id}/{yyyy-mm-dd}/{session_id}.{ext}.
//
// The opus encoder is exposed as an OpusEncoder interface — the production
// wiring injects a libopus-backed encoder; tests use a passthrough that
// records the PCM bytes verbatim so assertions stay deterministic.
package archive

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

// Bucket is the canonical destination per DES-0007 §4.
const Bucket = "monti-voice-sessions"

// OpusEncoder converts PCM 16k mono frames into opus chunks.
type OpusEncoder interface {
	Encode(pcm []byte) ([]byte, error)
	Flush() ([]byte, error)
}

// PassthroughOpusEncoder is a test encoder that emits PCM bytes verbatim.
// Production swaps in a libopus-backed implementation.
type PassthroughOpusEncoder struct{}

// Encode returns the PCM bytes verbatim.
func (PassthroughOpusEncoder) Encode(pcm []byte) ([]byte, error) { return pcm, nil }

// Flush returns nil.
func (PassthroughOpusEncoder) Flush() ([]byte, error) { return nil, nil }

// Sink writes byte streams to durable storage (MinIO in production).
type Sink interface {
	// AppendObject writes bytes to a named key, creating it on first call.
	AppendObject(ctx context.Context, key string, data []byte) error
	// Seal closes the object (no-op for sinks that always seal on AppendObject).
	Seal(ctx context.Context, key string) error
}

// TranscriptLine is one JSONL row.
type TranscriptLine struct {
	TS          int64  `json:"ts"`
	Role        string `json:"role"`
	Text        string `json:"text"`
	Final       bool   `json:"final"`
	Interrupted bool   `json:"interrupted,omitempty"`
}

// Archiver writes one session's audio + transcript to the sink.
type Archiver struct {
	sink     Sink
	encoder  OpusEncoder
	tenantID string
	day      string
	sessionID string

	mu     sync.Mutex
	closed bool
}

// New builds an Archiver for a single session.
func New(sink Sink, encoder OpusEncoder, tenantID, sessionID string, when time.Time) *Archiver {
	return &Archiver{
		sink:      sink,
		encoder:   encoder,
		tenantID:  tenantID,
		sessionID: sessionID,
		day:       when.UTC().Format("2006-01-02"),
	}
}

// AudioKey returns the .opus object key.
func (a *Archiver) AudioKey() string {
	return fmt.Sprintf("%s/%s/%s.opus", a.tenantID, a.day, a.sessionID)
}

// TranscriptKey returns the .transcript.jsonl object key.
func (a *Archiver) TranscriptKey() string {
	return fmt.Sprintf("%s/%s/%s.transcript.jsonl", a.tenantID, a.day, a.sessionID)
}

// AppendPCM encodes a PCM frame to opus and writes it to the audio object.
func (a *Archiver) AppendPCM(ctx context.Context, pcm []byte) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.closed {
		return errors.New("archive: sealed")
	}
	chunk, err := a.encoder.Encode(pcm)
	if err != nil {
		return err
	}
	if len(chunk) == 0 {
		return nil
	}
	return a.sink.AppendObject(ctx, a.AudioKey(), chunk)
}

// AppendTranscript serializes a transcript line and appends it.
func (a *Archiver) AppendTranscript(ctx context.Context, line TranscriptLine) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.closed {
		return errors.New("archive: sealed")
	}
	if line.TS == 0 {
		line.TS = time.Now().UnixMilli()
	}
	b, err := json.Marshal(line)
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return a.sink.AppendObject(ctx, a.TranscriptKey(), b)
}

// Close flushes the encoder and seals both objects.
func (a *Archiver) Close(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.closed {
		return nil
	}
	a.closed = true
	tail, err := a.encoder.Flush()
	if err == nil && len(tail) > 0 {
		if werr := a.sink.AppendObject(ctx, a.AudioKey(), tail); werr != nil {
			err = werr
		}
	}
	if sealErr := a.sink.Seal(ctx, a.AudioKey()); sealErr != nil && err == nil {
		err = sealErr
	}
	if sealErr := a.sink.Seal(ctx, a.TranscriptKey()); sealErr != nil && err == nil {
		err = sealErr
	}
	return err
}

// MemorySink is an in-memory Sink for tests.
type MemorySink struct {
	mu     sync.Mutex
	Blobs  map[string][]byte
	Sealed map[string]bool
}

// NewMemorySink builds an empty MemorySink.
func NewMemorySink() *MemorySink {
	return &MemorySink{
		Blobs:  make(map[string][]byte),
		Sealed: make(map[string]bool),
	}
}

// AppendObject implements Sink.
func (m *MemorySink) AppendObject(_ context.Context, key string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Blobs[key] = append(m.Blobs[key], data...)
	return nil
}

// Seal implements Sink.
func (m *MemorySink) Seal(_ context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Sealed[key] = true
	return nil
}

// Get returns a deep copy of an object.
func (m *MemorySink) Get(key string) []byte {
	m.mu.Lock()
	defer m.mu.Unlock()
	b := m.Blobs[key]
	out := make([]byte, len(b))
	copy(out, b)
	return out
}
