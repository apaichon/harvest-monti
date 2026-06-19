// repository.go — cart persistence (in-memory default).
package cart

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/apaichon/harvest-monti/internal/infra/db"
	"github.com/apaichon/harvest-monti/internal/infra/idgen"
)

// Repository is the cart data API.
type Repository interface {
	Open(ctx context.Context, c Cart) (*Cart, error)
	Get(ctx context.Context, tenantID, cartID string) (*Cart, error)
	AddItem(ctx context.Context, tenantID, cartID string, line LineItem) (*Cart, error)
	RemoveLine(ctx context.Context, tenantID, cartID, lineID string) (*Cart, error)
	SetPromo(ctx context.Context, tenantID, cartID, code string) (*Cart, error)
	SetStatus(ctx context.Context, tenantID, cartID string, s Status) (*Cart, error)
}

// MemoryRepo is the default in-memory implementation.
type MemoryRepo struct {
	mu    sync.Mutex
	store map[string]*Cart // by cart id
}

// NewMemoryRepo returns an empty repo.
func NewMemoryRepo() *MemoryRepo { return &MemoryRepo{store: map[string]*Cart{}} }

// Open creates a new cart, assigning an id if missing.
func (r *MemoryRepo) Open(_ context.Context, c Cart) (*Cart, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if c.ID == "" {
		c.ID = idgen.NewUUID()
	}
	c.Status = StatusOpen
	c.CreatedAt = time.Now().UTC()
	c.UpdatedAt = c.CreatedAt
	cp := c
	r.store[c.ID] = &cp
	return &cp, nil
}

// Get fetches a cart; enforces tenant scope.
func (r *MemoryRepo) Get(_ context.Context, tenantID, cartID string) (*Cart, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.store[cartID]
	if !ok || c.TenantID != tenantID {
		return nil, db.ErrNotFound
	}
	cp := *c
	return &cp, nil
}

// AddItem appends or dedupes a line based on LineKey.
func (r *MemoryRepo) AddItem(_ context.Context, tenantID, cartID string, line LineItem) (*Cart, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.store[cartID]
	if !ok || c.TenantID != tenantID {
		return nil, db.ErrNotFound
	}
	line.LineKey = LineKey(line)
	if line.ID == "" {
		line.ID = idgen.NewUUID()
	}
	// De-dupe: identical (item_id, line_key) lines collapse by qty.
	for i, existing := range c.Items {
		if existing.ItemID == line.ItemID && existing.LineKey == line.LineKey {
			c.Items[i].Qty += line.Qty
			c.UpdatedAt = time.Now().UTC()
			cp := *c
			return &cp, nil
		}
	}
	c.Items = append(c.Items, line)
	c.UpdatedAt = time.Now().UTC()
	cp := *c
	return &cp, nil
}

// RemoveLine drops one line.
func (r *MemoryRepo) RemoveLine(_ context.Context, tenantID, cartID, lineID string) (*Cart, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.store[cartID]
	if !ok || c.TenantID != tenantID {
		return nil, db.ErrNotFound
	}
	for i, l := range c.Items {
		if l.ID == lineID {
			c.Items = append(c.Items[:i], c.Items[i+1:]...)
			c.UpdatedAt = time.Now().UTC()
			cp := *c
			return &cp, nil
		}
	}
	return nil, db.ErrNotFound
}

// SetPromo records a promo code on the cart.
func (r *MemoryRepo) SetPromo(_ context.Context, tenantID, cartID, code string) (*Cart, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.store[cartID]
	if !ok || c.TenantID != tenantID {
		return nil, db.ErrNotFound
	}
	c.PromoCode = code
	c.UpdatedAt = time.Now().UTC()
	cp := *c
	return &cp, nil
}

// SetStatus transitions the cart status.
func (r *MemoryRepo) SetStatus(_ context.Context, tenantID, cartID string, s Status) (*Cart, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.store[cartID]
	if !ok || c.TenantID != tenantID {
		return nil, db.ErrNotFound
	}
	c.Status = s
	c.UpdatedAt = time.Now().UTC()
	cp := *c
	return &cp, nil
}

// LineKey returns the stable per-line dedupe hash combining notes +
// modifier option ids. Mirrors the DB-side line_key column populated by
// the application path (DES-0007 §3 partial unique index).
func LineKey(l LineItem) string {
	ids := make([]string, 0, len(l.Modifiers))
	for _, m := range l.Modifiers {
		ids = append(ids, m.OptionID)
	}
	sort.Strings(ids)
	h := md5.New()
	fmt.Fprintf(h, "notes=%s|mods=%s", strings.ToLower(l.Notes), strings.Join(ids, ","))
	return hex.EncodeToString(h.Sum(nil))
}
