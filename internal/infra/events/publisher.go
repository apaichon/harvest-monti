// publisher.go — small Publisher interface so plugin services can call
// `pub.Publish(ctx, SubjectCartItemAdded, envelope)` without knowing which
// transport sits behind it. Production wires this to nats.go JetStream;
// tests wire to InMemoryPublisher and assert on Seen().
package events

import (
	"context"
	"sync"
	"time"
)

// Envelope mirrors the DES-0007 §6 envelope schema. Payload is left as
// any so each plugin can carry its own typed event body — the worker /
// contract tests are the authority on payload shape.
type Envelope struct {
	EventID    string    `json:"event_id"`
	OccurredAt time.Time `json:"occurred_at"`
	TenantID   string    `json:"tenant_id"`
	OutletID   string    `json:"outlet_id,omitempty"`
	Actor      string    `json:"actor"`            // kiosk|mobile|voice|operator
	Payload    any       `json:"payload"`
}

// Publisher is the boot-injected event sink.
type Publisher interface {
	Publish(ctx context.Context, subject string, env Envelope) error
}

// InMemoryPublisher records every publish for tests.
type InMemoryPublisher struct {
	mu    sync.Mutex
	calls []Recorded
}

// Recorded is one publish event.
type Recorded struct {
	Subject  string
	Envelope Envelope
}

// Publish appends the call to the in-memory log.
func (p *InMemoryPublisher) Publish(_ context.Context, subject string, env Envelope) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.calls = append(p.calls, Recorded{Subject: subject, Envelope: env})
	return nil
}

// Seen returns recorded calls in publish order.
func (p *InMemoryPublisher) Seen() []Recorded {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]Recorded, len(p.calls))
	copy(out, p.calls)
	return out
}

// SeenSubjects is a small convenience for asserting subject sets in tests.
func (p *InMemoryPublisher) SeenSubjects() []string {
	rs := p.Seen()
	out := make([]string, len(rs))
	for i, r := range rs {
		out[i] = r.Subject
	}
	return out
}
