// routes.go — wires the three REST endpoints from TASK-0009 onto a Router.
package menu

import "github.com/apaichon/harvest-monti/internal/router"

// Register mounts the menu read endpoints. The auth group covers
// /api/v1/menu/* and /api/v1/recommendations (operator/staff JWT for now;
// kiosk reads land on the public router's mirror endpoints).
func Register(r *router.Router, h *Handler) {
	r.Get("/menu", h.list)
	r.Get("/menu/items/{item_id}", h.getItem)
	r.Get("/recommendations", h.recommendations)
}

// RegisterPublic mounts the same readers under the public group so
// unauthenticated kiosk + mobile UI can read tenants by ?tenant_id=.
func RegisterPublic(r *router.Router, h *Handler) {
	r.Get("/menu", h.list)
	r.Get("/menu/items/{item_id}", h.getItem)
	r.Get("/recommendations", h.recommendations)
}
