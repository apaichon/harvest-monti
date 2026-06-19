// Package ports declares the abstract collaborator interfaces the
// function-call dispatcher needs. The concrete implementations come from
// TASK-0009 (internal/plugins/menu, cart, order) and are wired in
// cmd/api/main.go. Keeping them as interfaces here means TASK-0010 can ship
// and be tested before TASK-0009 lands.
package ports

import "context"

// MenuPort is the read-only menu search collaborator.
type MenuPort interface {
	Search(ctx context.Context, q MenuSearchQuery) ([]MenuItemLite, error)
}

// MenuSearchQuery is the structured request.
type MenuSearchQuery struct {
	TenantID  string
	OutletID  string
	TableCode string
	SessionID string
	Query     string
	Allergens []string
	Category  string
}

// MenuItemLite is the projection the voice agent speaks back.
type MenuItemLite struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	PriceCents  int64   `json:"price_cents"`
	Currency    string  `json:"currency"`
	Allergens   []string `json:"allergens,omitempty"`
}

// CartPort mutates the per-session cart.
type CartPort interface {
	Add(ctx context.Context, req CartAddRequest) (CartState, error)
	Remove(ctx context.Context, req CartRemoveRequest) (CartState, error)
}

// CartAddRequest carries the trusted session identity.
type CartAddRequest struct {
	TenantID  string
	OutletID  string
	TableCode string
	SessionID string
	ItemID    string
	Qty       int
	Modifiers map[string]string
}

// CartRemoveRequest carries the trusted session identity.
type CartRemoveRequest struct {
	TenantID  string
	OutletID  string
	TableCode string
	SessionID string
	LineID    string
}

// CartState is the projection sent back to the model + UI.
type CartState struct {
	SessionID string     `json:"session_id"`
	Lines     []CartLine `json:"lines"`
	Subtotal  int64      `json:"subtotal_cents"`
	Currency  string     `json:"currency"`
}

// CartLine is a single cart row.
type CartLine struct {
	LineID    string            `json:"line_id"`
	ItemID    string            `json:"item_id"`
	Name      string            `json:"name"`
	Qty       int               `json:"qty"`
	Modifiers map[string]string `json:"modifiers,omitempty"`
	UnitCents int64             `json:"unit_cents"`
}

// OrderPort submits the order as a stubbed demo ticket per ADR-0005.
type OrderPort interface {
	Submit(ctx context.Context, req OrderSubmitRequest) (OrderTicket, error)
}

// OrderSubmitRequest captures trusted identity + payment method.
type OrderSubmitRequest struct {
	TenantID      string
	OutletID      string
	TableCode     string
	SessionID     string
	PaymentMethod string
}

// OrderTicket is the demo order envelope.
type OrderTicket struct {
	OrderID       string `json:"order_id"`
	SessionID     string `json:"session_id"`
	PaymentStatus string `json:"payment_status"` // "stubbed" per ADR-0005 resolved-ambiguities
	TotalCents    int64  `json:"total_cents"`
	Currency      string `json:"currency"`
}

// StaffPort emits a call-staff escalation. Wired to Slack webhook per DES-0008 §7.
type StaffPort interface {
	Call(ctx context.Context, req StaffCallRequest) (StaffPing, error)
}

// StaffCallRequest carries trusted identity + free-form reason.
type StaffCallRequest struct {
	TenantID  string
	OutletID  string
	TableCode string
	SessionID string
	Reason    string
}

// StaffPing is the escalation receipt.
type StaffPing struct {
	PingID    string `json:"ping_id"`
	NotifiedAt int64 `json:"notified_at"`
}

// LanguagePort handles explicit session language switches.
type LanguagePort interface {
	Switch(ctx context.Context, req LanguageSwitchRequest) (LanguageState, error)
}

// LanguageSwitchRequest carries trusted identity + the new BCP-47 code.
type LanguageSwitchRequest struct {
	TenantID  string
	OutletID  string
	TableCode string
	SessionID string
	BCP47     string
}

// LanguageState is the resulting session-language record.
type LanguageState struct {
	SessionID string `json:"session_id"`
	BCP47     string `json:"bcp47"`
}
