// repository.go — order persistence + status transitions. The promo
// catalog lives here too so other plugins can resolve codes by sharing one
// store (cart.ApplyPromo wires in via order.Service.LookupPromo).
package order

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/apaichon/harvest-monti/internal/infra/db"
	"github.com/apaichon/harvest-monti/internal/infra/idgen"
)

// Promo is a tenant-scoped promo code.
type Promo struct {
	Code       string
	TenantID   string
	Kind       string // percent|fixed
	Value      int
	MaxUses    int
	UsedCount  int
	ExpiresAt  *time.Time
	Status     string
}

// Repository is the order data API.
type Repository interface {
	Create(ctx context.Context, o Order) (*Order, error)
	Get(ctx context.Context, tenantID, orderID string) (*Order, error)
	GetByNumber(ctx context.Context, orderNumber string) (*Order, error)
	SetStatus(ctx context.Context, tenantID, orderID string, s Status) (*Order, error)

	GetPromo(ctx context.Context, tenantID, code string) (*Promo, error)
	SeedPromo(ctx context.Context, p Promo) error
}

// MemoryRepo is the default in-memory store.
type MemoryRepo struct {
	mu       sync.Mutex
	orders   map[string]*Order // by id
	byNumber map[string]string // order_number -> id
	promos   map[string]*Promo // by tenant|code
}

// NewMemoryRepo returns an empty repo.
func NewMemoryRepo() *MemoryRepo {
	return &MemoryRepo{
		orders:   map[string]*Order{},
		byNumber: map[string]string{},
		promos:   map[string]*Promo{},
	}
}

// Create assigns ids and stores the order.
func (r *MemoryRepo) Create(_ context.Context, o Order) (*Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if o.ID == "" {
		o.ID = idgen.NewUUID()
	}
	if o.OrderNumber == "" {
		o.OrderNumber = fmt.Sprintf("M-%s", strings.ToUpper(idgen.NewUUID()[:8]))
	}
	if o.Status == "" {
		o.Status = StatusReceived
	}
	if o.PlacedAt.IsZero() {
		o.PlacedAt = time.Now().UTC()
	}
	cp := o
	r.orders[o.ID] = &cp
	r.byNumber[o.OrderNumber] = o.ID
	return &cp, nil
}

// Get returns one order with tenant scope.
func (r *MemoryRepo) Get(_ context.Context, tenantID, orderID string) (*Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	o, ok := r.orders[orderID]
	if !ok || (tenantID != "" && o.TenantID != tenantID) {
		return nil, db.ErrNotFound
	}
	cp := *o
	return &cp, nil
}

// GetByNumber returns one order by order_number. Used by the public
// poll/WS endpoint — tenant scope is not enforced here because the
// order_number itself is unique-per-tenant and unguessable; the public
// handler treats it as a capability.
func (r *MemoryRepo) GetByNumber(_ context.Context, orderNumber string) (*Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id, ok := r.byNumber[orderNumber]
	if !ok {
		return nil, db.ErrNotFound
	}
	o := r.orders[id]
	cp := *o
	return &cp, nil
}

// SetStatus transitions the order. Forward-only except cancelled, which
// can fire from any pre-completed state.
func (r *MemoryRepo) SetStatus(_ context.Context, tenantID, orderID string, next Status) (*Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	o, ok := r.orders[orderID]
	if !ok || o.TenantID != tenantID {
		return nil, db.ErrNotFound
	}
	if !canTransition(o.Status, next) {
		return nil, ErrIllegalTransition
	}
	o.Status = next
	cp := *o
	return &cp, nil
}

// GetPromo resolves a tenant-scoped promo code.
func (r *MemoryRepo) GetPromo(_ context.Context, tenantID, code string) (*Promo, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.promos[tenantID+"|"+code]
	if !ok {
		return nil, db.ErrNotFound
	}
	cp := *p
	return &cp, nil
}

// SeedPromo loads a promo (test helper / demo seed).
func (r *MemoryRepo) SeedPromo(_ context.Context, p Promo) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := p
	r.promos[p.TenantID+"|"+p.Code] = &cp
	return nil
}

// canTransition enforces REQ-0012 AC-6 + DES-0007 §1 status guard.
func canTransition(cur, next Status) bool {
	if cur == StatusCompleted || cur == StatusCancelled {
		return false // terminal states
	}
	if next == StatusCancelled {
		return true // cancel from any pre-terminal
	}
	// Forward-only chain.
	order := map[Status]int{StatusReceived: 0, StatusPreparing: 1, StatusReady: 2, StatusCompleted: 3}
	cv, ok1 := order[cur]
	nv, ok2 := order[next]
	if !ok1 || !ok2 {
		return false
	}
	return nv == cv+1
}
