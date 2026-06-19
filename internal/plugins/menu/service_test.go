package menu

import (
	"context"
	"errors"
	"testing"

	"github.com/apaichon/harvest-monti/internal/infra/cache"
	"github.com/apaichon/harvest-monti/internal/infra/db"
)

func seededRepo(t *testing.T) *MemoryRepo {
	t.Helper()
	r := NewMemoryRepo()
	cats := []Category{
		{ID: "c1", Name: "Mains", DisplayOrder: 2, Status: "active"},
		{ID: "c2", Name: "Starters", DisplayOrder: 1, Status: "active"},
	}
	items := []Item{
		{ID: "i1", Name: "Pad Thai", CategoryID: "c1", PriceCents: 24000, Currency: "THB", IsBestSeller: true, DisplayOrder: 2, Status: "active", Allergens: []string{"peanut"}},
		{ID: "i2", Name: "Spring Roll", CategoryID: "c2", PriceCents: 9000, Currency: "THB", IsBestSeller: true, DisplayOrder: 1, Status: "active"},
		{ID: "i3", Name: "Soup", CategoryID: "c2", PriceCents: 12000, Currency: "THB", DisplayOrder: 3, Status: "active"},
	}
	if err := r.Seed(context.Background(), "tenant-A", cats, items); err != nil {
		t.Fatal(err)
	}
	// Tenant B has a separately-keyed item with the same id as tenant A's i1.
	_ = r.Seed(context.Background(), "tenant-B", []Category{{ID: "c1", Name: "X"}}, []Item{{ID: "i1", Name: "Other"}})
	return r
}

func TestListSortsCategoriesByDisplayOrder(t *testing.T) {
	s := NewService(seededRepo(t), &cache.NoopInvalidator{})
	out, err := s.List(context.Background(), "tenant-A", ItemFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if out.Categories[0].ID != "c2" {
		t.Errorf("categories not sorted by display_order: %v", out.Categories)
	}
}

func TestListExcludeAllergens(t *testing.T) {
	s := NewService(seededRepo(t), &cache.NoopInvalidator{})
	out, _ := s.List(context.Background(), "tenant-A", ItemFilter{ExcludeAllergens: []string{"peanut"}})
	for _, it := range out.Items {
		if it.ID == "i1" {
			t.Errorf("peanut-tagged item not filtered out")
		}
	}
}

func TestGetItemCrossTenantIsNotFound(t *testing.T) {
	s := NewService(seededRepo(t), &cache.NoopInvalidator{})
	// Tenant A asking for an item id that ALSO exists under tenant B should
	// see tenant A's row, not tenant B's (per-tenant isolation).
	got, err := s.GetItem(context.Background(), "tenant-A", "i1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Pad Thai" {
		t.Errorf("got tenant B's row leaking into tenant A: %+v", got)
	}
	// Tenant A asking for a non-existent id returns ErrNotFound.
	if _, err := s.GetItem(context.Background(), "tenant-A", "i999"); !errors.Is(err, db.ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

func TestRecommendationsOnlyBestSellers(t *testing.T) {
	s := NewService(seededRepo(t), &cache.NoopInvalidator{})
	out, _ := s.Recommendations(context.Background(), "tenant-A")
	if len(out) != 2 {
		t.Fatalf("expected 2 best-sellers, got %d", len(out))
	}
	if out[0].ID != "i2" {
		t.Errorf("expected display-order 1 item first, got %s", out[0].ID)
	}
}

func TestPurgeMenuCachePushesMontiTags(t *testing.T) {
	inv := &cache.NoopInvalidator{}
	s := NewService(seededRepo(t), inv)
	_ = s.PurgeMenuCache(context.Background(), "tenant-A", "i1")
	tags := inv.Seen()
	if len(tags) != 2 {
		t.Fatalf("expected 2 tags, got %v", tags)
	}
	for _, tg := range tags {
		if tg[:6] != "monti:" {
			t.Errorf("tag missing monti: prefix: %s", tg)
		}
	}
}
