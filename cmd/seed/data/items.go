// Package data declares the deterministic MONTI demo seed payload used by
// `cmd/seed`. All IDs are UUIDv5 derived from the harvest-monti seed namespace
// so re-running the seed against the same database resolves to the same rows
// (idempotency contract per TASK-0013 / DES-0007 §1).
package data

import (
	"github.com/google/uuid"
)

// SeedNamespace is the UUIDv5 namespace anchoring every deterministic ID below.
// Do not change once a tenant has been seeded against this value — the IDs
// become load-bearing for downstream MinIO object keys and cache tags.
var SeedNamespace = uuid.MustParse("8c2f9f9d-7b3a-4e15-9d4f-6c6b2c7e9a01")

// MustID returns a UUIDv5 derived from SeedNamespace + key — deterministic.
func MustID(key string) uuid.UUID {
	return uuid.NewSHA1(SeedNamespace, []byte(key))
}

// Tenant — REQ-0012 demo tenant.
type Tenant struct {
	ID       uuid.UUID
	Name     string
	Locale   string
	Currency string
}

// Outlet — REQ-0012 single outlet for the demo tenant.
type Outlet struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	Name       string
	Timezone   string
	TableCount int
}

// Menu — single active menu.
type Menu struct {
	ID            uuid.UUID
	OutletID      uuid.UUID
	Name          string
	Status        string
	EffectiveFrom string // ISO 8601 date
}

// Category — menu_categories row.
type Category struct {
	ID           uuid.UUID
	MenuID       uuid.UUID
	Name         string
	DisplayOrder int
	AccentColor  [3]uint8 // r,g,b — used for placeholder WebP fill
}

// Item — menu_items row.
type Item struct {
	ID            uuid.UUID
	MenuID        uuid.UUID
	CategoryID    uuid.UUID
	SKU           string
	Name          string
	Description   string
	PriceCents    int
	Currency      string
	IsBestSeller  bool
	DisplayOrder  int
	AccentColor   [3]uint8 // for placeholder image
	AllergenCodes []string
}

// ModifierOption — within a modifier group.
type ModifierOption struct {
	ID               uuid.UUID
	Name             string
	PriceDeltaCents  int
	DisplayOrder     int
}

// ModifierGroup — menu_items → modifier_groups.
type ModifierGroup struct {
	ID         uuid.UUID
	ItemID     uuid.UUID
	Name       string
	Kind       string // single | multi
	Required   bool
	MinSelect  int
	MaxSelect  int
	Options    []ModifierOption
}

// Allergen — reference table row.
type Allergen struct {
	ID      uuid.UUID
	Code    string
	LabelEN string
	LabelTH string
	LabelZH string
	LabelJA string
}

// PromoCode — promo_codes row.
type PromoCode struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	Code      string
	Kind      string // percent | fixed
	Value     int    // percent (0–100) or cents
	MaxUses   int
	ExpiresAt string // ISO date or empty
}

// Seed is the full declarative payload.
type Seed struct {
	Tenant     Tenant
	Outlet     Outlet
	Menu       Menu
	Categories []Category
	Items      []Item
	Modifiers  []ModifierGroup
	Allergens  []Allergen
	ItemAllergens map[uuid.UUID][]string // item_id -> allergen codes
	Promos     []PromoCode
}

// Build returns the populated demo seed. Deterministic.
func Build() Seed {
	tenantID := MustID("tenant:monti-demo")
	outletID := MustID("outlet:monti-bangkok-sukhumvit")
	menuID := MustID("menu:dinner-2026")

	tenant := Tenant{
		ID:       tenantID,
		Name:     "Monti Demo Bistro",
		Locale:   "en-US",
		Currency: "USD",
	}
	outlet := Outlet{
		ID:         outletID,
		TenantID:   tenantID,
		Name:       "Monti Bangkok Sukhumvit",
		Timezone:   "Asia/Bangkok",
		TableCount: 24,
	}
	menu := Menu{
		ID:            menuID,
		OutletID:      outletID,
		Name:          "Dinner Menu",
		Status:        "active",
		EffectiveFrom: "2026-06-19",
	}

	// Categories
	catPasta := Category{ID: MustID("category:pasta"), MenuID: menuID, Name: "Pasta", DisplayOrder: 1, AccentColor: [3]uint8{0xC8, 0xA8, 0x6A}}
	catSteak := Category{ID: MustID("category:steak"), MenuID: menuID, Name: "Steak", DisplayOrder: 2, AccentColor: [3]uint8{0x7B, 0x3F, 0x2A}}
	catSalmon := Category{ID: MustID("category:salmon"), MenuID: menuID, Name: "Salmon", DisplayOrder: 3, AccentColor: [3]uint8{0xE0, 0x7A, 0x5F}}
	catPizza := Category{ID: MustID("category:pizza"), MenuID: menuID, Name: "Pizza", DisplayOrder: 4, AccentColor: [3]uint8{0xD6, 0x28, 0x28}}
	catBurger := Category{ID: MustID("category:burger"), MenuID: menuID, Name: "Burger", DisplayOrder: 5, AccentColor: [3]uint8{0x8B, 0x45, 0x13}}

	categories := []Category{catPasta, catSteak, catSalmon, catPizza, catBurger}

	// Items
	itemTruffle := Item{
		ID:           MustID("item:truffle-pasta"),
		MenuID:       menuID,
		CategoryID:   catPasta.ID,
		SKU:          "MNT-PAS-001",
		Name:         "Truffle Pasta",
		Description:  "Handmade pasta tossed in a rich black-truffle cream sauce. Finished with shaved Parmigiano-Reggiano and fresh thyme.",
		PriceCents:   1850,
		Currency:     "USD",
		IsBestSeller: true,
		DisplayOrder: 1,
		AccentColor:  [3]uint8{0xC8, 0xA8, 0x6A},
		AllergenCodes: []string{"gluten", "dairy"},
	}
	itemSteak := Item{
		ID:           MustID("item:grilled-wagyu-steak"),
		MenuID:       menuID,
		CategoryID:   catSteak.ID,
		SKU:          "MNT-STK-001",
		Name:         "Grilled Wagyu Steak",
		Description:  "A5 Wagyu ribeye seared over Japanese charcoal and basted with garlic butter. Served with rosemary jus and grilled seasonal vegetables.",
		PriceCents:   3280,
		Currency:     "USD",
		IsBestSeller: true,
		DisplayOrder: 1,
		AccentColor:  [3]uint8{0x7B, 0x3F, 0x2A},
	}
	itemSalmon := Item{
		ID:           MustID("item:grilled-salmon"),
		MenuID:       menuID,
		CategoryID:   catSalmon.ID,
		SKU:          "MNT-SAL-001",
		Name:         "Grilled Salmon",
		Description:  "Atlantic salmon fillet glazed with miso-honey and grilled to a crisp finish. Plated over saffron rice with a citrus beurre blanc.",
		PriceCents:   2450,
		Currency:     "USD",
		IsBestSeller: false,
		DisplayOrder: 1,
		AccentColor:  [3]uint8{0xE0, 0x7A, 0x5F},
		AllergenCodes: []string{"fish"},
	}
	itemPizza := Item{
		ID:           MustID("item:margherita-pizza"),
		MenuID:       menuID,
		CategoryID:   catPizza.ID,
		SKU:          "MNT-PIZ-001",
		Name:         "Margherita Pizza",
		Description:  "Naples-style hand-stretched dough topped with San Marzano tomato sugo, fior di latte, and fresh basil. Wood-fired at 480 C for sixty seconds.",
		PriceCents:   1950,
		Currency:     "USD",
		IsBestSeller: false,
		DisplayOrder: 1,
		AccentColor:  [3]uint8{0xD6, 0x28, 0x28},
		AllergenCodes: []string{"gluten", "dairy"},
	}
	itemBurger := Item{
		ID:           MustID("item:wagyu-burger"),
		MenuID:       menuID,
		CategoryID:   catBurger.ID,
		SKU:          "MNT-BRG-001",
		Name:         "Wagyu Burger",
		Description:  "Hand-formed Wagyu patty layered with aged cheddar, smoked bacon, and house pickle on a toasted sesame brioche. Served with truffle fries.",
		PriceCents:   1890,
		Currency:     "USD",
		IsBestSeller: false,
		DisplayOrder: 1,
		AccentColor:  [3]uint8{0x8B, 0x45, 0x13},
		AllergenCodes: []string{"gluten", "sesame"},
	}

	items := []Item{itemTruffle, itemSteak, itemSalmon, itemPizza, itemBurger}

	// Modifier groups (per DES-0007 §1 and TASK-0013 spec)
	mgTruffle := ModifierGroup{
		ID:        MustID("mg:truffle:pasta-type"),
		ItemID:    itemTruffle.ID,
		Name:      "Pasta type",
		Kind:      "single",
		Required:  true,
		MinSelect: 1,
		MaxSelect: 1,
		Options: []ModifierOption{
			{ID: MustID("mo:truffle:penne"), Name: "Penne", PriceDeltaCents: 0, DisplayOrder: 1},
			{ID: MustID("mo:truffle:spaghetti"), Name: "Spaghetti", PriceDeltaCents: 0, DisplayOrder: 2},
			{ID: MustID("mo:truffle:linguine"), Name: "Linguine", PriceDeltaCents: 0, DisplayOrder: 3},
		},
	}
	mgSteak := ModifierGroup{
		ID:        MustID("mg:steak:doneness"),
		ItemID:    itemSteak.ID,
		Name:      "Doneness",
		Kind:      "single",
		Required:  true,
		MinSelect: 1,
		MaxSelect: 1,
		Options: []ModifierOption{
			{ID: MustID("mo:steak:rare"), Name: "Rare", DisplayOrder: 1},
			{ID: MustID("mo:steak:medium-rare"), Name: "Medium-Rare", DisplayOrder: 2},
			{ID: MustID("mo:steak:medium"), Name: "Medium", DisplayOrder: 3},
			{ID: MustID("mo:steak:medium-well"), Name: "Medium-Well", DisplayOrder: 4},
			{ID: MustID("mo:steak:well-done"), Name: "Well-Done", DisplayOrder: 5},
		},
	}
	mgPizza := ModifierGroup{
		ID:        MustID("mg:pizza:toppings"),
		ItemID:    itemPizza.ID,
		Name:      "Toppings",
		Kind:      "multi",
		Required:  false,
		MinSelect: 0,
		MaxSelect: 4,
		Options: []ModifierOption{
			{ID: MustID("mo:pizza:mushrooms"), Name: "Mushrooms", PriceDeltaCents: 200, DisplayOrder: 1},
			{ID: MustID("mo:pizza:olives"), Name: "Olives", PriceDeltaCents: 150, DisplayOrder: 2},
			{ID: MustID("mo:pizza:basil"), Name: "Basil", PriceDeltaCents: 50, DisplayOrder: 3},
			{ID: MustID("mo:pizza:bell-pepper"), Name: "Bell Pepper", PriceDeltaCents: 100, DisplayOrder: 4},
		},
	}
	mgBurger := ModifierGroup{
		ID:        MustID("mg:burger:add-side"),
		ItemID:    itemBurger.ID,
		Name:      "Add Side",
		Kind:      "single",
		Required:  false,
		MinSelect: 0,
		MaxSelect: 1,
		Options: []ModifierOption{
			{ID: MustID("mo:burger:salad"), Name: "Salad", PriceDeltaCents: 300, DisplayOrder: 1},
			{ID: MustID("mo:burger:soup"), Name: "Soup", PriceDeltaCents: 300, DisplayOrder: 2},
			{ID: MustID("mo:burger:bread"), Name: "Bread", PriceDeltaCents: 0, DisplayOrder: 3},
		},
	}

	modifiers := []ModifierGroup{mgTruffle, mgSteak, mgPizza, mgBurger}

	// Allergens — top-10 most common
	allergens := []Allergen{
		{ID: MustID("allergen:peanut"), Code: "peanut", LabelEN: "Peanut", LabelTH: "ถั่วลิสง", LabelZH: "花生", LabelJA: "ピーナッツ"},
		{ID: MustID("allergen:tree_nut"), Code: "tree_nut", LabelEN: "Tree Nut", LabelTH: "ถั่วเปลือกแข็ง", LabelZH: "树坚果", LabelJA: "ナッツ"},
		{ID: MustID("allergen:gluten"), Code: "gluten", LabelEN: "Gluten", LabelTH: "กลูเตน", LabelZH: "麸质", LabelJA: "グルテン"},
		{ID: MustID("allergen:dairy"), Code: "dairy", LabelEN: "Dairy", LabelTH: "ผลิตภัณฑ์นม", LabelZH: "乳制品", LabelJA: "乳製品"},
		{ID: MustID("allergen:egg"), Code: "egg", LabelEN: "Egg", LabelTH: "ไข่", LabelZH: "鸡蛋", LabelJA: "卵"},
		{ID: MustID("allergen:soy"), Code: "soy", LabelEN: "Soy", LabelTH: "ถั่วเหลือง", LabelZH: "大豆", LabelJA: "大豆"},
		{ID: MustID("allergen:shellfish"), Code: "shellfish", LabelEN: "Shellfish", LabelTH: "หอย/กุ้ง", LabelZH: "贝类", LabelJA: "甲殻類"},
		{ID: MustID("allergen:fish"), Code: "fish", LabelEN: "Fish", LabelTH: "ปลา", LabelZH: "鱼", LabelJA: "魚"},
		{ID: MustID("allergen:sesame"), Code: "sesame", LabelEN: "Sesame", LabelTH: "งา", LabelZH: "芝麻", LabelJA: "ごま"},
		{ID: MustID("allergen:sulfite"), Code: "sulfite", LabelEN: "Sulfite", LabelTH: "ซัลไฟต์", LabelZH: "亚硫酸盐", LabelJA: "亜硫酸塩"},
	}

	itemAllergens := map[uuid.UUID][]string{}
	for _, it := range items {
		if len(it.AllergenCodes) > 0 {
			itemAllergens[it.ID] = it.AllergenCodes
		}
	}

	// Promo codes
	promos := []PromoCode{
		{
			ID:        MustID("promo:appetit10"),
			TenantID:  tenantID,
			Code:      "APPETIT10",
			Kind:      "percent",
			Value:     10,
			MaxUses:   1000,
			ExpiresAt: "2026-12-31",
		},
		{
			ID:        MustID("promo:welcome"),
			TenantID:  tenantID,
			Code:      "WELCOME",
			Kind:      "fixed",
			Value:     500, // 500 cents = $5
			MaxUses:   100,
			ExpiresAt: "2026-12-31",
		},
	}

	return Seed{
		Tenant:        tenant,
		Outlet:        outlet,
		Menu:          menu,
		Categories:    categories,
		Items:         items,
		Modifiers:     modifiers,
		Allergens:     allergens,
		ItemAllergens: itemAllergens,
		Promos:        promos,
	}
}
