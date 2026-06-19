// Package objectstore bootstraps the three MinIO buckets MONTI requires
// (DES-0007 §4). The implementation is interface-driven so unit tests can
// run without a live MinIO and the production binary can swap in a real
// minio-go client when online (left as a small wiring change in cmd/monti-api).
//
// Buckets:
//   - monti-menu-images        public-read, immutable cache 1y
//   - monti-voice-sessions     private, 30-day expiration lifecycle
//   - monti-order-receipts     private, 7-year retention
package objectstore

import (
	"context"
	"fmt"
	"sync"
)

// BucketSpec captures the design intent for one bucket.
type BucketSpec struct {
	Name        string
	Public      bool
	CacheControl string // applied as object metadata via signed PUT; empty = none
	LifecycleDays int   // 0 = no expiration; positive = days to expiration
	RetentionYears int  // 0 = none; positive = retention floor in years
}

// MontiBuckets returns the three buckets from DES-0007 §4 in deterministic
// order so bootstrap logs are stable across restarts.
func MontiBuckets() []BucketSpec {
	return []BucketSpec{
		{
			Name:         "monti-menu-images",
			Public:       true,
			CacheControl: "public, max-age=31536000, immutable",
		},
		{
			Name:          "monti-voice-sessions",
			Public:        false,
			LifecycleDays: 30,
		},
		{
			Name:           "monti-order-receipts",
			Public:         false,
			RetentionYears: 7,
		},
	}
}

// Client is the small slice of minio-go we exercise during bootstrap.
// The real impl wraps *minio.Client; tests use FakeClient.
type Client interface {
	EnsureBucket(ctx context.Context, spec BucketSpec) error
}

// FakeClient records ensure-calls in memory. Tests assert against Seen().
type FakeClient struct {
	mu   sync.Mutex
	seen []string
}

// EnsureBucket records the bucket name; never errors.
func (f *FakeClient) EnsureBucket(_ context.Context, spec BucketSpec) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.seen = append(f.seen, spec.Name)
	return nil
}

// Seen returns the bucket names ensured so far.
func (f *FakeClient) Seen() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]string, len(f.seen))
	copy(out, f.seen)
	return out
}

// Bootstrap ensures every bucket in MontiBuckets() exists via the supplied
// client. Callers invoke this from main() after the MinIO connection is
// healthy; non-blocking on individual failures so the API still serves the
// public menu endpoints when MinIO is degraded (DES-0007 §4 last paragraph).
func Bootstrap(ctx context.Context, c Client) error {
	if c == nil {
		return fmt.Errorf("objectstore: nil client passed to Bootstrap")
	}
	var firstErr error
	for _, b := range MontiBuckets() {
		if err := c.EnsureBucket(ctx, b); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("ensure bucket %s: %w", b.Name, err)
		}
	}
	return firstErr
}
