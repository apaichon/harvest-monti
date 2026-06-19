// routes.go — wires order endpoints.
package order

import "github.com/apaichon/harvest-monti/internal/router"

// Register mounts the authenticated order endpoints under /api/v1.
func Register(r *router.Router, h *Handler) {
	r.Post("/orders", h.place)
	r.Post("/operator/orders/{order_id}/advance", h.advance)
}

// RegisterPublic mounts the two public endpoints under /api/v1/public.
func RegisterPublic(r *router.Router, h *Handler) {
	r.Get("/orders/{order_number}", h.poll)
	r.Get("/orders/{order_number}/stream", h.stream)
}
