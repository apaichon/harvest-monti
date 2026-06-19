// repository.go — menu persistence. The default implementation is the
// in-memory store used by tests + the initial demo; the production wiring
// replaces it with a *sql.DB-backed Repo (same interface) once the
// harvest-deployment Postgres is reachable.
package menu

import (
	"context"
	"sort"
	"strings"
	"sync"

	"github.com/apaichon/harvest-monti/internal/infra/db"
)

// Repository is the tenant-scoped data API consumed by Service.
type Repository interface {
	ListCategories(ctx context.Context, tenantID string) ([]Category, error)
	ListItems(ctx context.Context, tenantID string, f ItemFilter) ([]Item, error)
	GetItem(ctx context.Context, tenantID, itemID string) (*Item, error)
	BestSellers(ctx context.Context, tenantID string) ([]Item, error)
	Seed(ctx context.Context, tenantID string, cats []Category, items []Item) error
}

// MemoryRepo is the in-memory implementation. Safe for concurrent use.
type MemoryRepo struct {
	mu         sync.RWMutex
	categories map[string][]Category // by tenant
	items      map[string][]Item     // by tenant
}

// NewMemoryRepo returns an empty in-memory repo.
func NewMemoryRepo() *MemoryRepo {
	return &MemoryRepo{
		categories: map[string][]Category{},
		items:      map[string][]Item{},
	}
}

// Seed loads categories and items for one tenant; replaces existing data.
func (r *MemoryRepo) Seed(_ context.Context, tenantID string, cats []Category, items []Item) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.categories[tenantID] = append([]Category{}, cats...)
	tagged := make([]Item, len(items))
	for i := range items {
		tagged[i] = items[i]
		tagged[i].TenantID = tenantID
	}
	r.items[tenantID] = tagged
	return nil
}

// ListCategories returns categories ordered by display_order.
func (r *MemoryRepo) ListCategories(_ context.Context, tenantID string) ([]Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	src := r.categories[tenantID]
	out := append([]Category{}, src...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].DisplayOrder < out[j].DisplayOrder })
	return out, nil
}

// ListItems returns items, filtered by allergy exclusion if requested.
func (r *MemoryRepo) ListItems(_ context.Context, tenantID string, f ItemFilter) ([]Item, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	src := r.items[tenantID]
	out := make([]Item, 0, len(src))
itemLoop:
	for _, it := range src {
		for _, ex := range f.ExcludeAllergens {
			for _, a := range it.Allergens {
				if strings.EqualFold(a, ex) {
					continue itemLoop
				}
			}
		}
		out = append(out, it)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].DisplayOrder < out[j].DisplayOrder })
	return out, nil
}

// GetItem returns one item, scoped by tenant — never reveals cross-tenant
// existence (TEST-0008 TC-10).
func (r *MemoryRepo) GetItem(_ context.Context, tenantID, itemID string) (*Item, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, it := range r.items[tenantID] {
		if it.ID == itemID {
			cp := it
			return &cp, nil
		}
	}
	return nil, db.ErrNotFound
}

// BestSellers returns is_best_seller items ordered by display_order
// (DES-0007 §7 — recommendation ranking source of truth for MVP).
func (r *MemoryRepo) BestSellers(_ context.Context, tenantID string) ([]Item, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	src := r.items[tenantID]
	out := make([]Item, 0, len(src))
	for _, it := range src {
		if it.IsBestSeller {
			out = append(out, it)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].DisplayOrder < out[j].DisplayOrder })
	return out, nil
}
