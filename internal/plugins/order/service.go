// service.go — order business logic: totals enforcement, status machine,
// cart-to-order placement, NATS events.
package order

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/apaichon/harvest-monti/internal/infra/cache"
	"github.com/apaichon/harvest-monti/internal/infra/events"
	"github.com/apaichon/harvest-monti/internal/plugins/cart"
)

// ErrIllegalTransition signals a forbidden status change (REQ-0012 AC-6).
var ErrIllegalTransition = errors.New("order: illegal status transition")

// ErrTotalsMismatch is returned when the submitted totals do not satisfy
// the DES-0007 §3 CHECK formula. Caught at the service layer so the
// kiosk/mobile see a clean 422 instead of a Postgres CHECK violation.
var ErrTotalsMismatch = errors.New("order: total_cents != subtotal + service_charge + tax - promo_discount")

// Service is the order domain entry-point.
type Service struct {
	repo      Repository
	carts     *cart.Service
	cartRepo  cart.Repository
	inv       cache.Invalidator
	pub       events.Publisher

	subMu  sync.Mutex
	subs   map[string][]chan Order // keyed by order_number; status stream subscribers
}

// NewService wires dependencies.
func NewService(r Repository, c *cart.Service, cr cart.Repository, inv cache.Invalidator, pub events.Publisher) *Service {
	return &Service{
		repo:     r,
		carts:    c,
		cartRepo: cr,
		inv:      inv,
		pub:      pub,
		subs:     map[string][]chan Order{},
	}
}

// LookupPromo implements cart.PromoLookup so the cart plugin can resolve
// codes without owning the promo table.
func (s *Service) LookupPromo(ctx context.Context, tenantID, code string) (string, int, error) {
	p, err := s.repo.GetPromo(ctx, tenantID, code)
	if err != nil {
		return "", 0, err
	}
	if p.Status != "" && p.Status != "active" {
		return "", 0, fmt.Errorf("promo inactive")
	}
	if p.MaxUses > 0 && p.UsedCount >= p.MaxUses {
		return "", 0, fmt.Errorf("promo exhausted")
	}
	return p.Kind, p.Value, nil
}

// PlaceFromCart converts a submitted cart into an order, enforcing the
// totals formula from DES-0007 §3 and emitting monti.order.placed.
func (s *Service) PlaceFromCart(ctx context.Context, tenantID string, req PlaceRequest) (*Order, error) {
	// Enforce server-side totals (mirrors the DB CHECK constraint).
	if !TotalsMatch(req.SubtotalCents, req.ServiceChargeCents, req.TaxCents, req.PromoDiscountCents, ComputedTotal(req)) {
		return nil, ErrTotalsMismatch
	}
	c, err := s.cartRepo.Get(ctx, tenantID, req.CartID)
	if err != nil {
		return nil, err
	}
	lines := make([]Line, 0, len(c.Items))
	for _, ci := range c.Items {
		mods := make([]Modifier, 0, len(ci.Modifiers))
		for _, m := range ci.Modifiers {
			mods = append(mods, Modifier{OptionID: m.OptionID, PriceDeltaCents: m.PriceDeltaSnapshot})
		}
		lines = append(lines, Line{
			ItemID:         ci.ItemID,
			Qty:            ci.Qty,
			UnitPriceCents: ci.UnitPriceSnapshot,
			Notes:          ci.Notes,
			Modifiers:      mods,
		})
	}

	o, err := s.repo.Create(ctx, Order{
		TenantID:           tenantID,
		OutletID:           c.OutletID,
		TableCode:          c.TableCode,
		SubtotalCents:      req.SubtotalCents,
		ServiceChargeCents: req.ServiceChargeCents,
		TaxCents:           req.TaxCents,
		PromoDiscountCents: req.PromoDiscountCents,
		TotalCents:         ComputedTotal(req),
		Currency:           req.Currency,
		Items:              lines,
	})
	if err != nil {
		return nil, err
	}
	// Mark the cart submitted.
	if _, err := s.carts.Submit(ctx, tenantID, req.CartID); err != nil && !errors.Is(err, cart.ErrIllegalTransition) {
		return nil, err
	}
	s.invalidate(ctx, o)
	s.publishOrder(ctx, events.SubjectOrderPlaced, o)
	return o, nil
}

// Advance pushes the order to next status (operator action).
func (s *Service) Advance(ctx context.Context, tenantID, orderID string, next Status) (*Order, error) {
	o, err := s.repo.SetStatus(ctx, tenantID, orderID, next)
	if errors.Is(err, ErrIllegalTransition) {
		return nil, ErrIllegalTransition
	}
	if err != nil {
		return nil, err
	}
	s.invalidate(ctx, o)
	s.publishOrder(ctx, events.SubjectOrderStatusChanged, o)
	if o.Status == StatusCompleted {
		s.publishOrder(ctx, events.SubjectOrderCompleted, o)
	}
	s.fanOut(*o)
	return o, nil
}

// GetByNumber returns one order for the public poll endpoint.
func (s *Service) GetByNumber(ctx context.Context, orderNumber string) (*Order, error) {
	return s.repo.GetByNumber(ctx, orderNumber)
}

// Subscribe registers a channel for order-status frames keyed by
// order_number. Returns an unsubscribe func the caller defers.
func (s *Service) Subscribe(orderNumber string) (<-chan Order, func()) {
	ch := make(chan Order, 8)
	s.subMu.Lock()
	s.subs[orderNumber] = append(s.subs[orderNumber], ch)
	s.subMu.Unlock()
	unsub := func() {
		s.subMu.Lock()
		defer s.subMu.Unlock()
		list := s.subs[orderNumber]
		for i, c := range list {
			if c == ch {
				s.subs[orderNumber] = append(list[:i], list[i+1:]...)
				close(ch)
				return
			}
		}
	}
	return ch, unsub
}

func (s *Service) fanOut(o Order) {
	s.subMu.Lock()
	defer s.subMu.Unlock()
	for _, ch := range s.subs[o.OrderNumber] {
		select {
		case ch <- o:
		default: // drop on full buffer; clients re-poll
		}
	}
}

func (s *Service) invalidate(ctx context.Context, o *Order) {
	if s.inv == nil || o == nil {
		return
	}
	_ = s.inv.PurgeTags(ctx, cache.OrderTag(o.TenantID, o.ID))
}

func (s *Service) publishOrder(ctx context.Context, subject string, o *Order) {
	if s.pub == nil || o == nil {
		return
	}
	_ = s.pub.Publish(ctx, subject, events.Envelope{
		TenantID: o.TenantID,
		OutletID: o.OutletID,
		Actor:    "operator",
		Payload: map[string]any{
			"order_id":     o.ID,
			"order_number": o.OrderNumber,
			"status":       string(o.Status),
		},
	})
}

// ComputedTotal returns subtotal + service + tax - discount.
func ComputedTotal(req PlaceRequest) int {
	return req.SubtotalCents + req.ServiceChargeCents + req.TaxCents - req.PromoDiscountCents
}

// TotalsMatch checks whether the wire-supplied total equals the formula.
func TotalsMatch(subtotal, service, tax, discount, total int) bool {
	return total == subtotal+service+tax-discount
}
