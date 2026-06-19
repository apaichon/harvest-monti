// Package cart owns the cart lifecycle (DES-0007 §1 carts/cart_items/cart_item_modifiers).
package cart

import "time"

// Status enum for carts.
type Status string

// Cart lifecycle status values mirror the DB enum cart_status.
const (
	StatusOpen         Status = "open"
	StatusCheckingOut  Status = "checking_out"
	StatusSubmitted    Status = "submitted"
	StatusAbandoned    Status = "abandoned"
)

// Cart mirrors the carts row consumed by the kiosk + mobile.
type Cart struct {
	ID         string     `json:"id"`
	TenantID   string     `json:"tenant_id"`
	OutletID   string     `json:"outlet_id"`
	TableCode  string     `json:"table_code,omitempty"`
	SessionID  string     `json:"session_id"`
	Status     Status     `json:"status"`
	PromoCode  string     `json:"promo_code,omitempty"`
	Items      []LineItem `json:"items"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// LineItem is one cart line.
type LineItem struct {
	ID                 string     `json:"id"`
	ItemID             string     `json:"item_id"`
	Qty                int        `json:"qty"`
	UnitPriceSnapshot  int        `json:"unit_price_snapshot"`
	Notes              string     `json:"notes,omitempty"`
	LineKey            string     `json:"-"`
	Modifiers          []Modifier `json:"modifiers,omitempty"`
}

// Modifier is one selected option on a line.
type Modifier struct {
	OptionID            string `json:"option_id"`
	PriceDeltaSnapshot  int    `json:"price_delta_snapshot"`
}

// OpenCartRequest is the body of POST /api/v1/carts.
type OpenCartRequest struct {
	TenantID  string `json:"tenant_id"`
	OutletID  string `json:"outlet_id"`
	SessionID string `json:"session_id"`
	TableCode string `json:"table_code,omitempty"`
}

// OpenCartResponse is the response to POST /api/v1/carts.
type OpenCartResponse struct {
	CartID    string `json:"cart_id"`
	SessionID string `json:"session_id"`
}

// AddItemRequest is the body of POST /api/v1/carts/{cart_id}/items.
type AddItemRequest struct {
	ItemID            string     `json:"item_id"`
	Qty               int        `json:"qty"`
	UnitPriceSnapshot int        `json:"unit_price_snapshot"`
	Notes             string     `json:"notes,omitempty"`
	Modifiers         []Modifier `json:"modifiers,omitempty"`
}

// PromoRequest is the body of POST /api/v1/carts/{cart_id}/promo.
type PromoRequest struct {
	Code string `json:"code"`
}
