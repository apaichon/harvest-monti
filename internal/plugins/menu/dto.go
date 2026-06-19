// Package menu owns the kiosk + mobile menu surface (DES-0007 §1 menu_*).
package menu

// Category mirrors the menu_categories row consumed by kiosks.
type Category struct {
	ID           string `json:"id"`
	MenuID       string `json:"menu_id"`
	Name         string `json:"name"`
	DisplayOrder int    `json:"display_order"`
	Status       string `json:"status"`
}

// Item mirrors menu_items + flattened allergens/modifier-group ids for
// kiosk list responses.
type Item struct {
	ID             string   `json:"id"`
	MenuID         string   `json:"menu_id"`
	CategoryID     string   `json:"category_id"`
	SKU            string   `json:"sku,omitempty"`
	Name           string   `json:"name"`
	Description    string   `json:"description,omitempty"`
	PriceCents     int      `json:"price_cents"`
	Currency       string   `json:"currency"`
	ImagePath      string   `json:"image_path,omitempty"`
	Calories       int      `json:"calories,omitempty"`
	IsBestSeller   bool     `json:"is_best_seller"`
	DisplayOrder   int      `json:"display_order"`
	Status         string   `json:"status"`
	Allergens      []string `json:"allergens,omitempty"`
	ModifierGroups []ModifierGroup `json:"modifier_groups,omitempty"`
	TenantID       string   `json:"-"`
}

// ModifierGroup mirrors modifier_groups joined with modifier_options.
type ModifierGroup struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Kind      string            `json:"kind"`     // single|multi
	Required  bool              `json:"required"`
	MinSelect int               `json:"min_select"`
	MaxSelect int               `json:"max_select"`
	Options   []ModifierOption  `json:"options"`
}

// ModifierOption mirrors modifier_options.
type ModifierOption struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	PriceDeltaCents   int    `json:"price_delta_cents"`
	DisplayOrder      int    `json:"display_order"`
}

// ListResponse is the body of GET /api/v1/menu.
type ListResponse struct {
	Categories []Category `json:"categories"`
	Items      []Item     `json:"items"`
}

// ItemFilter narrows search by allergy exclusion.
type ItemFilter struct {
	ExcludeAllergens []string
}
