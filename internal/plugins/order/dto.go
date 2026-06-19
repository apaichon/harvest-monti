// Package order owns order placement + status lifecycle (DES-0007 §1 orders/order_items/order_item_modifiers).
package order

import "time"

// Status mirrors the DB enum fulfillment_status.
type Status string

// Status values; forward-only except cancelled (REQ-0012 AC-6).
const (
	StatusReceived   Status = "received"
	StatusPreparing  Status = "preparing"
	StatusReady      Status = "ready"
	StatusCompleted  Status = "completed"
	StatusCancelled  Status = "cancelled"
)

// Order is the public order shape returned by GET endpoints + the WS stream.
type Order struct {
	ID                  string    `json:"id"`
	TenantID            string    `json:"-"`
	OutletID            string    `json:"outlet_id"`
	OrderNumber         string    `json:"order_number"`
	TableCode           string    `json:"table_code,omitempty"`
	SubtotalCents       int       `json:"subtotal_cents"`
	ServiceChargeCents  int       `json:"service_charge_cents"`
	TaxCents            int       `json:"tax_cents"`
	PromoDiscountCents  int       `json:"promo_discount_cents"`
	TotalCents          int       `json:"total_cents"`
	Currency            string    `json:"currency"`
	Status              Status    `json:"status"`
	Items               []Line    `json:"items"`
	PlacedAt            time.Time `json:"placed_at"`
	ReadyETAAt          *time.Time `json:"ready_eta_at,omitempty"`
}

// Line is one order_items row.
type Line struct {
	ID              string     `json:"id"`
	ItemID          string     `json:"item_id"`
	Qty             int        `json:"qty"`
	UnitPriceCents  int        `json:"unit_price_cents"`
	Notes           string     `json:"notes,omitempty"`
	Modifiers       []Modifier `json:"modifiers,omitempty"`
}

// Modifier is one order_item_modifiers row.
type Modifier struct {
	OptionID         string `json:"option_id"`
	PriceDeltaCents  int    `json:"price_delta_cents"`
}

// PlaceRequest is the body of POST /api/v1/orders.
type PlaceRequest struct {
	CartID              string `json:"cart_id"`
	SubtotalCents       int    `json:"subtotal_cents"`
	ServiceChargeCents  int    `json:"service_charge_cents"`
	TaxCents            int    `json:"tax_cents"`
	PromoDiscountCents  int    `json:"promo_discount_cents"`
	Currency            string `json:"currency"`
	PaymentMethod       string `json:"payment_method,omitempty"`
}

// AdvanceRequest is the body of POST /api/v1/operator/orders/{id}/advance.
type AdvanceRequest struct {
	NextStatus Status `json:"next_status"`
}
