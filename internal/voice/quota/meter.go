// Package quota implements the per-tenant Gemini-minute counter from
// DES-0009 §10 and the ADR-0005 addendum.
//
// Authority split (ADR-0005 2026-06-19 addendum): this Redis counter is the
// OPERATIONAL gauge the Switcher consults for sub-millisecond fallback
// decisions. It MAY drift. The harvest-core billing.usage_events ledger is
// the BILLING authority. This package never touches the ledger; under-counting
// here is acceptable, over-counting only causes spurious fallbacks (also
// acceptable). Month rollover lets the key expire naturally.
package quota

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// KeyFor returns the Redis key for a tenant + month.
// Format: monti:voice:quota:{tenant_id}:{yyyy-mm}.
func KeyFor(tenantID string, t time.Time) string {
	return fmt.Sprintf("monti:voice:quota:%s:%s", tenantID, t.UTC().Format("2006-01"))
}

// Store abstracts the Redis backend so tests can use an in-memory map. The
// real implementation in cmd/api/main.go wraps redis.Client.
type Store interface {
	IncrBy(ctx context.Context, key string, delta int64) (int64, error)
	Get(ctx context.Context, key string) (int64, error)
}

// Meter is the per-tenant quota counter.
type Meter struct {
	store Store
	now   func() time.Time

	// CapResolver yields the per-tenant minutes cap (tenants.monti_voice_minutes_cap).
	CapResolver func(ctx context.Context, tenantID string) (int64, error)

	// SoftWarnRatio fires the quota_warning frame at this fraction of cap (DES-0009 §10).
	SoftWarnRatio float64
}

// New builds a Meter.
func New(store Store, capResolver func(ctx context.Context, tenantID string) (int64, error)) *Meter {
	return &Meter{
		store:         store,
		now:           time.Now,
		CapResolver:   capResolver,
		SoftWarnRatio: 0.8,
	}
}

// WithClock overrides time.Now for tests.
func (m *Meter) WithClock(now func() time.Time) *Meter {
	m.now = now
	return m
}

// IncrementMinute adds one minute of Gemini wall-clock for a tenant and
// returns the new total. Per DES-0009 §10 only Gemini minutes count.
func (m *Meter) IncrementMinute(ctx context.Context, tenantID string) (int64, error) {
	return m.store.IncrBy(ctx, KeyFor(tenantID, m.now()), 1)
}

// Used returns minutes consumed this month.
func (m *Meter) Used(ctx context.Context, tenantID string) (int64, error) {
	return m.store.Get(ctx, KeyFor(tenantID, m.now()))
}

// Check returns the soft/hard signal a Switcher needs to make a fallback
// decision based on the current usage.
type Check struct {
	Used    int64
	Cap     int64
	Soft    bool // crossed SoftWarnRatio
	Hard    bool // at or above cap
}

// Evaluate reads the current usage and resolves the tenant cap.
func (m *Meter) Evaluate(ctx context.Context, tenantID string) (Check, error) {
	used, err := m.Used(ctx, tenantID)
	if err != nil {
		return Check{}, err
	}
	cap, err := m.CapResolver(ctx, tenantID)
	if err != nil {
		return Check{}, err
	}
	if cap <= 0 {
		return Check{Used: used, Cap: cap}, nil
	}
	return Check{
		Used: used,
		Cap:  cap,
		Soft: float64(used)/float64(cap) >= m.SoftWarnRatio,
		Hard: used >= cap,
	}, nil
}

// MemoryStore is a goroutine-safe in-memory Store for unit tests.
type MemoryStore struct {
	mu sync.Mutex
	kv map[string]int64
}

// NewMemoryStore builds an empty MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{kv: make(map[string]int64)}
}

// IncrBy implements Store.
func (m *MemoryStore) IncrBy(_ context.Context, key string, delta int64) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.kv[key] += delta
	return m.kv[key], nil
}

// Get implements Store.
func (m *MemoryStore) Get(_ context.Context, key string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.kv[key], nil
}

// Set is a test helper.
func (m *MemoryStore) Set(key string, v int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.kv[key] = v
}

// ErrNoCap is returned by a Meter whose CapResolver yields zero or a negative number.
var ErrNoCap = errors.New("quota: tenant has no positive minute cap configured")
