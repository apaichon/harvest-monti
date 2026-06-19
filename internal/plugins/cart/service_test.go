package cart

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/apaichon/harvest-monti/internal/infra/cache"
	"github.com/apaichon/harvest-monti/internal/infra/events"
)

func newSvc(t *testing.T) (*Service, *events.InMemoryPublisher, *cache.NoopInvalidator) {
	t.Helper()
	pub := &events.InMemoryPublisher{}
	inv := &cache.NoopInvalidator{}
	promo := PromoLookup(func(_ context.Context, _, code string) (string, int, error) {
		if code == "LUNCH20" {
			return "percent", 20, nil
		}
		return "", 0, errors.New("not found")
	})
	return NewService(NewMemoryRepo(), inv, pub, promo), pub, inv
}

func TestOpenEmitsCartOpenedAndInvalidates(t *testing.T) {
	s, pub, inv := newSvc(t)
	resp, err := s.Open(context.Background(), OpenCartRequest{
		TenantID: "tA", OutletID: "oA", SessionID: "sess-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.CartID == "" {
		t.Fatal("cart id not assigned")
	}
	if got := pub.SeenSubjects(); len(got) != 1 || got[0] != events.SubjectCartOpened {
		t.Fatalf("expected cart.opened, got %v", got)
	}
	tags := inv.Seen()
	if len(tags) != 1 || !strings.HasPrefix(tags[0], "monti:cart:") {
		t.Fatalf("expected monti:cart: tag, got %v", tags)
	}
}

func TestAddItemDedupesIdenticalLines(t *testing.T) {
	s, pub, _ := newSvc(t)
	open, _ := s.Open(context.Background(), OpenCartRequest{TenantID: "tA", OutletID: "oA", SessionID: "s1"})
	for i := 0; i < 3; i++ {
		_, err := s.AddItem(context.Background(), "tA", open.CartID, AddItemRequest{
			ItemID: "i1", Qty: 1, UnitPriceSnapshot: 24000,
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	c, _ := s.Get(context.Background(), "tA", open.CartID)
	if len(c.Items) != 1 || c.Items[0].Qty != 3 {
		t.Errorf("dedupe failed: %+v", c.Items)
	}
	// Three item.added events should have been emitted (plus the one cart.opened).
	added := 0
	for _, ev := range pub.SeenSubjects() {
		if ev == events.SubjectCartItemAdded {
			added++
		}
	}
	if added != 3 {
		t.Errorf("expected 3 item.added events, got %d", added)
	}
}

func TestAddItemDifferentModifiersAreSeparateLines(t *testing.T) {
	s, _, _ := newSvc(t)
	open, _ := s.Open(context.Background(), OpenCartRequest{TenantID: "tA", OutletID: "oA", SessionID: "s1"})
	_, _ = s.AddItem(context.Background(), "tA", open.CartID, AddItemRequest{ItemID: "i1", Qty: 1, Modifiers: []Modifier{{OptionID: "mo-a"}}})
	_, _ = s.AddItem(context.Background(), "tA", open.CartID, AddItemRequest{ItemID: "i1", Qty: 1, Modifiers: []Modifier{{OptionID: "mo-b"}}})
	c, _ := s.Get(context.Background(), "tA", open.CartID)
	if len(c.Items) != 2 {
		t.Errorf("expected 2 lines for different modifiers, got %d", len(c.Items))
	}
}

func TestApplyPromoRejectsUnknownCode(t *testing.T) {
	s, _, _ := newSvc(t)
	open, _ := s.Open(context.Background(), OpenCartRequest{TenantID: "tA", OutletID: "oA", SessionID: "s1"})
	if _, err := s.ApplyPromo(context.Background(), "tA", open.CartID, "BOGUS"); !errors.Is(err, ErrInvalidPromo) {
		t.Fatalf("want ErrInvalidPromo, got %v", err)
	}
	c, err := s.ApplyPromo(context.Background(), "tA", open.CartID, "LUNCH20")
	if err != nil {
		t.Fatal(err)
	}
	if c.PromoCode != "LUNCH20" {
		t.Errorf("promo not recorded: %+v", c)
	}
}

func TestSubmitRejectsAlreadySubmitted(t *testing.T) {
	s, _, _ := newSvc(t)
	open, _ := s.Open(context.Background(), OpenCartRequest{TenantID: "tA", OutletID: "oA", SessionID: "s1"})
	if _, err := s.Submit(context.Background(), "tA", open.CartID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Submit(context.Background(), "tA", open.CartID); !errors.Is(err, ErrIllegalTransition) {
		t.Fatalf("expected illegal transition, got %v", err)
	}
}

func TestCrossTenantCartAccessIsNotFound(t *testing.T) {
	s, _, _ := newSvc(t)
	open, _ := s.Open(context.Background(), OpenCartRequest{TenantID: "tA", OutletID: "oA", SessionID: "s1"})
	if _, err := s.Get(context.Background(), "tB", open.CartID); err == nil {
		t.Fatal("tenant B should not see tenant A's cart")
	}
}
