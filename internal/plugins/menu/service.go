// service.go — menu business logic. Thin wrapper around Repository plus
// cache-tag invalidation hooks (DES-0007 §5 row 1+2).
package menu

import (
	"context"

	"github.com/apaichon/harvest-monti/internal/infra/cache"
)

// Service is the menu domain entry-point consumed by the handler.
type Service struct {
	repo Repository
	inv  cache.Invalidator
}

// NewService wires a Repository + Invalidator.
func NewService(r Repository, inv cache.Invalidator) *Service {
	return &Service{repo: r, inv: inv}
}

// List returns the kiosk menu (categories + items) for a tenant.
func (s *Service) List(ctx context.Context, tenantID string, f ItemFilter) (*ListResponse, error) {
	cats, err := s.repo.ListCategories(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	items, err := s.repo.ListItems(ctx, tenantID, f)
	if err != nil {
		return nil, err
	}
	return &ListResponse{Categories: cats, Items: items}, nil
}

// GetItem returns one item with its modifier groups + allergens.
func (s *Service) GetItem(ctx context.Context, tenantID, itemID string) (*Item, error) {
	return s.repo.GetItem(ctx, tenantID, itemID)
}

// Recommendations returns best-sellers ordered by display_order.
func (s *Service) Recommendations(ctx context.Context, tenantID string) ([]Item, error) {
	return s.repo.BestSellers(ctx, tenantID)
}

// PurgeMenuCache is called by admin paths (menu publish / item edit).
func (s *Service) PurgeMenuCache(ctx context.Context, tenantID, itemID string) error {
	if s.inv == nil {
		return nil
	}
	tags := []string{cache.MenuTag(tenantID)}
	if itemID != "" {
		tags = append(tags, cache.MenuItemTag(tenantID, itemID))
	}
	return s.inv.PurgeTags(ctx, tags...)
}
