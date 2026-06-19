package quota_test

import (
	"context"
	"testing"
	"time"

	"github.com/apaichon/harvest-monti/internal/voice/quota"
)

func TestMeter_KeyFormat(t *testing.T) {
	tt := time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC)
	got := quota.KeyFor("tenant-monti-demo", tt)
	want := "monti:voice:quota:tenant-monti-demo:2026-06"
	if got != want {
		t.Fatalf("key format: want %q got %q", want, got)
	}
}

func TestMeter_IncrementAndCheck(t *testing.T) {
	store := quota.NewMemoryStore()
	caps := map[string]int64{"tenant-A": 100}
	m := quota.New(store, func(_ context.Context, tid string) (int64, error) {
		return caps[tid], nil
	}).WithClock(func() time.Time { return time.Date(2026, 6, 19, 0, 0, 0, 0, time.UTC) })

	for i := 0; i < 79; i++ {
		if _, err := m.IncrementMinute(context.Background(), "tenant-A"); err != nil {
			t.Fatalf("incr: %v", err)
		}
	}
	check, err := m.Evaluate(context.Background(), "tenant-A")
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	if check.Used != 79 || check.Cap != 100 {
		t.Fatalf("counts: used=%d cap=%d", check.Used, check.Cap)
	}
	if check.Soft {
		t.Errorf("soft should not yet fire at 79/100")
	}

	// Cross the 80% threshold.
	_, _ = m.IncrementMinute(context.Background(), "tenant-A")
	check, _ = m.Evaluate(context.Background(), "tenant-A")
	if !check.Soft {
		t.Errorf("soft should fire at 80/100")
	}
	if check.Hard {
		t.Errorf("hard should not yet fire at 80/100")
	}

	// Cross the cap.
	for i := 0; i < 25; i++ {
		_, _ = m.IncrementMinute(context.Background(), "tenant-A")
	}
	check, _ = m.Evaluate(context.Background(), "tenant-A")
	if !check.Hard {
		t.Errorf("hard should fire at 105/100")
	}
}

// Monthly rollover yields a fresh key (under-counting per the ADR-0005
// addendum is acceptable; we never reset the previous key).
func TestMeter_RolloverKeyIsFresh(t *testing.T) {
	store := quota.NewMemoryStore()
	caps := map[string]int64{"tenant-A": 100}
	clock := time.Date(2026, 6, 30, 23, 59, 0, 0, time.UTC)
	m := quota.New(store, func(_ context.Context, tid string) (int64, error) {
		return caps[tid], nil
	}).WithClock(func() time.Time { return clock })

	_, _ = m.IncrementMinute(context.Background(), "tenant-A")
	juneUsed, _ := m.Used(context.Background(), "tenant-A")
	if juneUsed != 1 {
		t.Fatalf("expected 1 minute in june, got %d", juneUsed)
	}

	clock = time.Date(2026, 7, 1, 0, 0, 1, 0, time.UTC)
	julyUsed, _ := m.Used(context.Background(), "tenant-A")
	if julyUsed != 0 {
		t.Errorf("expected fresh july key starts at 0, got %d", julyUsed)
	}
}
