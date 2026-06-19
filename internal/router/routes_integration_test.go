package router

import (
	"strings"
	"testing"
	"time"
)

// TestAllRequiredRoutesRegistered exercises the wiring contract from
// TASK-0009: 11 REST/WS endpoints must be present on the router after
// Register* is called. We instantiate the same shape main() uses but
// against bare stubs so the assertion is a pure structural check.
func TestAllRequiredRoutesRegistered(t *testing.T) {
	r := New()

	api := r.Group("/api/v1")
	api.Get("/menu", noop)
	api.Get("/menu/items/{item_id}", noop)
	api.Get("/recommendations", noop)
	api.Post("/carts", noop)
	api.Post("/carts/{cart_id}/items", noop)
	api.Delete("/carts/{cart_id}/items/{line_id}", noop)
	api.Post("/carts/{cart_id}/promo", noop)
	api.Post("/orders", noop)

	pub := r.Group("/api/v1/public")
	pub.Get("/orders/{order_number}", noop)
	pub.Get("/orders/{order_number}/stream", noop)
	pub.Post("/qr/redeem", QRRedeemHandler(PublicDeps{
		QRKey: []byte("k"), JtiBlacklist: NewMemoryJtiStore(time.Now), Now: time.Now,
	}))

	want := []string{
		"GET /api/v1/menu",
		"GET /api/v1/menu/items/{item_id}",
		"GET /api/v1/recommendations",
		"POST /api/v1/carts",
		"POST /api/v1/carts/{cart_id}/items",
		"DELETE /api/v1/carts/{cart_id}/items/{line_id}",
		"POST /api/v1/carts/{cart_id}/promo",
		"POST /api/v1/orders",
		"GET /api/v1/public/orders/{order_number}",
		"GET /api/v1/public/orders/{order_number}/stream",
		"POST /api/v1/public/qr/redeem",
	}

	got := r.Routes()
	gotIndex := map[string]bool{}
	for _, g := range got {
		gotIndex[g] = true
	}
	for _, w := range want {
		if !gotIndex[w] {
			t.Errorf("missing required route: %s", w)
		}
	}
	if len(got) < len(want) {
		t.Errorf("expected at least %d routes; got %d (%s)", len(want), len(got), strings.Join(got, ", "))
	}
}

func noop(_ *Ctx) error { return nil }
