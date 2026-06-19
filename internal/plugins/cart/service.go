// service.go — cart business logic. Encapsulates lifecycle transitions,
// promo application, totals computation, and NATS event emission.
package cart

import (
	"context"
	"errors"
	"fmt"

	"github.com/apaichon/harvest-monti/internal/infra/cache"
	"github.com/apaichon/harvest-monti/internal/infra/events"
)

// PromoLookup resolves a promo code into a (kind, value) tuple. The
// service does not own promo persistence — that lives in the order
// plugin so multiple plugins can share the same codes table.
type PromoLookup func(ctx context.Context, tenantID, code string) (kind string, value int, err error)

// ErrInvalidPromo is returned when PromoLookup says the code is unknown.
var ErrInvalidPromo = errors.New("cart: invalid promo code")

// ErrIllegalTransition signals a forbidden status change.
var ErrIllegalTransition = errors.New("cart: illegal status transition")

// Service is the cart domain entry-point.
type Service struct {
	repo  Repository
	inv   cache.Invalidator
	pub   events.Publisher
	promo PromoLookup
}

// NewService wires repo + invalidator + publisher + promo lookup.
func NewService(r Repository, inv cache.Invalidator, pub events.Publisher, p PromoLookup) *Service {
	return &Service{repo: r, inv: inv, pub: pub, promo: p}
}

// Open creates a new cart and emits monti.cart.opened.
func (s *Service) Open(ctx context.Context, req OpenCartRequest) (*OpenCartResponse, error) {
	if req.TenantID == "" || req.OutletID == "" || req.SessionID == "" {
		return nil, fmt.Errorf("tenant_id, outlet_id and session_id are required")
	}
	c, err := s.repo.Open(ctx, Cart{
		TenantID:  req.TenantID,
		OutletID:  req.OutletID,
		SessionID: req.SessionID,
		TableCode: req.TableCode,
	})
	if err != nil {
		return nil, err
	}
	s.invalidate(ctx, c)
	s.publish(ctx, events.SubjectCartOpened, c, nil)
	return &OpenCartResponse{CartID: c.ID, SessionID: c.SessionID}, nil
}

// AddItem appends or de-dupes a line and emits monti.cart.item.added.
func (s *Service) AddItem(ctx context.Context, tenantID, cartID string, req AddItemRequest) (*Cart, error) {
	if req.Qty <= 0 {
		return nil, fmt.Errorf("qty must be > 0")
	}
	line := LineItem{
		ItemID:            req.ItemID,
		Qty:               req.Qty,
		UnitPriceSnapshot: req.UnitPriceSnapshot,
		Notes:             req.Notes,
		Modifiers:         req.Modifiers,
	}
	c, err := s.repo.AddItem(ctx, tenantID, cartID, line)
	if err != nil {
		return nil, err
	}
	s.invalidate(ctx, c)
	s.publish(ctx, events.SubjectCartItemAdded, c, map[string]any{
		"item_id": req.ItemID,
		"qty":     req.Qty,
	})
	return c, nil
}

// RemoveLine drops a line and emits monti.cart.item.removed.
func (s *Service) RemoveLine(ctx context.Context, tenantID, cartID, lineID string) (*Cart, error) {
	c, err := s.repo.RemoveLine(ctx, tenantID, cartID, lineID)
	if err != nil {
		return nil, err
	}
	s.invalidate(ctx, c)
	s.publish(ctx, events.SubjectCartItemRemoved, c, map[string]any{"line_id": lineID})
	return c, nil
}

// ApplyPromo validates a code and records it on the cart.
func (s *Service) ApplyPromo(ctx context.Context, tenantID, cartID, code string) (*Cart, error) {
	if s.promo == nil {
		return nil, ErrInvalidPromo
	}
	if _, _, err := s.promo(ctx, tenantID, code); err != nil {
		return nil, ErrInvalidPromo
	}
	c, err := s.repo.SetPromo(ctx, tenantID, cartID, code)
	if err != nil {
		return nil, err
	}
	s.invalidate(ctx, c)
	return c, nil
}

// Submit transitions the cart to "submitted". Used by order.PlaceFromCart.
func (s *Service) Submit(ctx context.Context, tenantID, cartID string) (*Cart, error) {
	c, err := s.repo.Get(ctx, tenantID, cartID)
	if err != nil {
		return nil, err
	}
	switch c.Status {
	case StatusOpen, StatusCheckingOut:
		// allowed
	default:
		return nil, ErrIllegalTransition
	}
	c2, err := s.repo.SetStatus(ctx, tenantID, cartID, StatusSubmitted)
	if err != nil {
		return nil, err
	}
	s.invalidate(ctx, c2)
	s.publish(ctx, events.SubjectCartSubmitted, c2, nil)
	return c2, nil
}

// Get returns the cart with tenant scope.
func (s *Service) Get(ctx context.Context, tenantID, cartID string) (*Cart, error) {
	return s.repo.Get(ctx, tenantID, cartID)
}

func (s *Service) invalidate(ctx context.Context, c *Cart) {
	if s.inv == nil || c == nil {
		return
	}
	_ = s.inv.PurgeTags(ctx, cache.CartTag(c.TenantID, c.SessionID))
}

func (s *Service) publish(ctx context.Context, subject string, c *Cart, payload any) {
	if s.pub == nil || c == nil {
		return
	}
	_ = s.pub.Publish(ctx, subject, events.Envelope{
		TenantID: c.TenantID,
		OutletID: c.OutletID,
		Actor:    "kiosk",
		Payload:  payload,
	})
}
