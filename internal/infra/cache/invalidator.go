// invalidator.go — runtime side of the cache-tag contract.
// The real implementation issues `DEL` against Redis on every tag passed
// in. For unit tests and offline builds we ship a no-op Invalidator so
// plugin services can call `cache.Inv.PurgeTags(...)` unconditionally.
package cache

import (
	"context"
	"sync"
)

// Invalidator purges cache keys by tag. Both kiosk SSR and the Gemini
// Live function dispatcher (TASK-0010) call into the same instance so
// mutation paths stay consistent (DES-0007 §5 closing paragraph).
type Invalidator interface {
	PurgeTags(ctx context.Context, tags ...string) error
}

// NoopInvalidator records calls in-memory; used by tests and by the
// bootstrap when Redis is not configured.
type NoopInvalidator struct {
	mu   sync.Mutex
	seen []string
}

// PurgeTags appends the tags to the noop log.
func (n *NoopInvalidator) PurgeTags(_ context.Context, tags ...string) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.seen = append(n.seen, tags...)
	return nil
}

// Seen returns the tags purged so far (test-only accessor).
func (n *NoopInvalidator) Seen() []string {
	n.mu.Lock()
	defer n.mu.Unlock()
	out := make([]string, len(n.seen))
	copy(out, n.seen)
	return out
}
