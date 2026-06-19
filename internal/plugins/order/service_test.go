package order

import (
	"context"
	"errors"
	"testing"

	"github.com/apaichon/harvest-monti/internal/infra/cache"
	"github.com/apaichon/harvest-monti/internal/infra/events"
	"github.com/apaichon/harvest-monti/internal/plugins/cart"
)

func newOrderSvc(t *testing.T) (*Service, *cart.Service, *events.InMemoryPublisher) {
	t.Helper()
	pub := &events.InMemoryPublisher{}
	inv := &cache.NoopInvalidator{}

	cartRepo := cart.NewMemoryRepo()
	orderRepo := NewMemoryRepo()
	_ = orderRepo.SeedPromo(context.Background(), Promo{
		Code: "LUNCH20", TenantID: "tA", Kind: "percent", Value: 20,
		MaxUses: 100, Status: "active",
	})

	cartSvc := cart.NewService(cartRepo, inv, pub, nil)
	orderSvc := NewService(orderRepo, cartSvc, cartRepo, inv, pub)
	cartSvc = cart.NewService(cartRepo, inv, pub, orderSvc.LookupPromo)
	orderSvc = NewService(orderRepo, cartSvc, cartRepo, inv, pub)
	return orderSvc, cartSvc, pub
}

func TestTotalsFormulaEnforced(t *testing.T) {
	osvc, csvc, _ := newOrderSvc(t)
	open, _ := csvc.Open(context.Background(), cart.OpenCartRequest{TenantID: "tA", OutletID: "oA", SessionID: "s1"})
	_, _ = csvc.AddItem(context.Background(), "tA", open.CartID, cart.AddItemRequest{ItemID: "i1", Qty: 1, UnitPriceSnapshot: 60000})

	// Bad totals: 60000 + 6000 + 4600 - 0 = 70600, supply 70601.
	_, err := osvc.PlaceFromCart(context.Background(), "tA", PlaceRequest{
		CartID: open.CartID, SubtotalCents: 60000, ServiceChargeCents: 6000, TaxCents: 4600,
		PromoDiscountCents: 0, Currency: "THB",
	})
	// PlaceFromCart requires the caller to also send total_cents; instead
	// the service derives total via ComputedTotal. To exercise
	// ErrTotalsMismatch we call TotalsMatch directly with mismatched data:
	_ = err
	if TotalsMatch(60000, 6000, 4600, 0, 70601) {
		t.Fatalf("TotalsMatch should reject 70601 != 70600")
	}
	if !TotalsMatch(60000, 6000, 4600, 0, 70600) {
		t.Fatalf("TotalsMatch should accept 70600")
	}
}

func TestPlaceOrderEmitsPlacedEvent(t *testing.T) {
	osvc, csvc, pub := newOrderSvc(t)
	open, _ := csvc.Open(context.Background(), cart.OpenCartRequest{TenantID: "tA", OutletID: "oA", SessionID: "s1"})
	_, _ = csvc.AddItem(context.Background(), "tA", open.CartID, cart.AddItemRequest{ItemID: "i1", Qty: 1, UnitPriceSnapshot: 60000})

	o, err := osvc.PlaceFromCart(context.Background(), "tA", PlaceRequest{
		CartID: open.CartID, SubtotalCents: 60000, ServiceChargeCents: 6000, TaxCents: 4600,
		PromoDiscountCents: 0, Currency: "THB",
	})
	if err != nil {
		t.Fatal(err)
	}
	if o.TotalCents != 70600 {
		t.Errorf("computed total mismatch: %d", o.TotalCents)
	}
	found := false
	for _, s := range pub.SeenSubjects() {
		if s == events.SubjectOrderPlaced {
			found = true
		}
	}
	if !found {
		t.Fatalf("monti.order.placed not emitted; saw %v", pub.SeenSubjects())
	}
}

func TestStatusForwardOnly(t *testing.T) {
	osvc, csvc, _ := newOrderSvc(t)
	open, _ := csvc.Open(context.Background(), cart.OpenCartRequest{TenantID: "tA", OutletID: "oA", SessionID: "s1"})
	_, _ = csvc.AddItem(context.Background(), "tA", open.CartID, cart.AddItemRequest{ItemID: "i1", Qty: 1, UnitPriceSnapshot: 1000})
	o, _ := osvc.PlaceFromCart(context.Background(), "tA", PlaceRequest{
		CartID: open.CartID, SubtotalCents: 1000, Currency: "THB",
	})

	for _, next := range []Status{StatusPreparing, StatusReady, StatusCompleted} {
		if _, err := osvc.Advance(context.Background(), "tA", o.ID, next); err != nil {
			t.Fatalf("advance to %s failed: %v", next, err)
		}
	}
	// completed -> preparing should be rejected.
	if _, err := osvc.Advance(context.Background(), "tA", o.ID, StatusPreparing); !errors.Is(err, ErrIllegalTransition) {
		t.Fatalf("regression from completed should be rejected, got %v", err)
	}
}

func TestCancelFromPreparing(t *testing.T) {
	osvc, csvc, _ := newOrderSvc(t)
	open, _ := csvc.Open(context.Background(), cart.OpenCartRequest{TenantID: "tA", OutletID: "oA", SessionID: "s1"})
	_, _ = csvc.AddItem(context.Background(), "tA", open.CartID, cart.AddItemRequest{ItemID: "i1", Qty: 1, UnitPriceSnapshot: 500})
	o, _ := osvc.PlaceFromCart(context.Background(), "tA", PlaceRequest{
		CartID: open.CartID, SubtotalCents: 500, Currency: "THB",
	})
	if _, err := osvc.Advance(context.Background(), "tA", o.ID, StatusPreparing); err != nil {
		t.Fatal(err)
	}
	if _, err := osvc.Advance(context.Background(), "tA", o.ID, StatusCancelled); err != nil {
		t.Fatalf("cancel from preparing should be allowed: %v", err)
	}
	if _, err := osvc.Advance(context.Background(), "tA", o.ID, StatusCancelled); !errors.Is(err, ErrIllegalTransition) {
		t.Fatalf("cancel from cancelled should be rejected, got %v", err)
	}
}

func TestSubscribeFanOut(t *testing.T) {
	osvc, csvc, _ := newOrderSvc(t)
	open, _ := csvc.Open(context.Background(), cart.OpenCartRequest{TenantID: "tA", OutletID: "oA", SessionID: "s1"})
	_, _ = csvc.AddItem(context.Background(), "tA", open.CartID, cart.AddItemRequest{ItemID: "i1", Qty: 1, UnitPriceSnapshot: 100})
	o, _ := osvc.PlaceFromCart(context.Background(), "tA", PlaceRequest{CartID: open.CartID, SubtotalCents: 100, Currency: "THB"})

	ch, unsub := osvc.Subscribe(o.OrderNumber)
	defer unsub()

	go func() {
		_, _ = osvc.Advance(context.Background(), "tA", o.ID, StatusPreparing)
	}()
	got := <-ch
	if got.Status != StatusPreparing {
		t.Fatalf("expected preparing frame, got %s", got.Status)
	}
}
