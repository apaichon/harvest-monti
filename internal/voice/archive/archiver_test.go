package archive_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/apaichon/harvest-monti/internal/voice/archive"
)

func TestArchiver_KeysAndAppend(t *testing.T) {
	sink := archive.NewMemorySink()
	when := time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC)
	a := archive.New(sink, archive.PassthroughOpusEncoder{},
		"tenant-monti-demo", "01HXYZSESSION", when)

	if got := a.AudioKey(); got != "tenant-monti-demo/2026-06-19/01HXYZSESSION.opus" {
		t.Errorf("audio key: %s", got)
	}
	if got := a.TranscriptKey(); got != "tenant-monti-demo/2026-06-19/01HXYZSESSION.transcript.jsonl" {
		t.Errorf("transcript key: %s", got)
	}

	ctx := context.Background()
	if err := a.AppendPCM(ctx, []byte{0x00, 0x10, 0x00, 0x10}); err != nil {
		t.Fatalf("append pcm: %v", err)
	}
	if err := a.AppendTranscript(ctx, archive.TranscriptLine{
		TS: 1, Role: "user", Text: "hello", Final: true,
	}); err != nil {
		t.Fatalf("append transcript: %v", err)
	}

	if err := a.Close(ctx); err != nil {
		t.Fatalf("close: %v", err)
	}
	if !sink.Sealed[a.AudioKey()] || !sink.Sealed[a.TranscriptKey()] {
		t.Errorf("expected both objects sealed on close")
	}

	txt := string(sink.Get(a.TranscriptKey()))
	if !strings.Contains(txt, `"role":"user"`) || !strings.Contains(txt, `"text":"hello"`) {
		t.Errorf("transcript jsonl missing fields: %q", txt)
	}
}

// After Close, further appends fail (no half-written tails).
func TestArchiver_NoAppendAfterClose(t *testing.T) {
	a := archive.New(archive.NewMemorySink(), archive.PassthroughOpusEncoder{},
		"t", "s", time.Date(2026, 6, 19, 0, 0, 0, 0, time.UTC))
	_ = a.Close(context.Background())
	if err := a.AppendPCM(context.Background(), []byte{0x01}); err == nil {
		t.Error("expected error appending after close")
	}
	if err := a.AppendTranscript(context.Background(), archive.TranscriptLine{Role: "user"}); err == nil {
		t.Error("expected error appending transcript after close")
	}
}
