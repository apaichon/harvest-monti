// Command monti-api boots the harvest-monti HTTP surface.
//
// Wiring order (DES-0007 §2 / §4 / §6):
//   1. MinIO bucket bootstrap (objectstore.Bootstrap).
//   2. NATS JetStream subject assertion (events.AllSubjects()).
//   3. Cache invalidator (Redis in prod, NoopInvalidator otherwise).
//   4. Plugin services: menu, cart, order — wired in dependency order
//      because cart.PromoLookup is supplied by order.LookupPromo.
//   5. Router groups:
//        /api/v1            — authenticated (Tenant + JWT)
//        /api/v1/public     — unauthenticated (QR redeem, order poll, WS)
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/apaichon/harvest-monti/internal/infra/cache"
	"github.com/apaichon/harvest-monti/internal/infra/events"
	"github.com/apaichon/harvest-monti/internal/infra/objectstore"
	"github.com/apaichon/harvest-monti/internal/middleware"
	"github.com/apaichon/harvest-monti/internal/plugins/cart"
	"github.com/apaichon/harvest-monti/internal/plugins/menu"
	"github.com/apaichon/harvest-monti/internal/plugins/order"
	"github.com/apaichon/harvest-monti/internal/router"
)

func main() {
	ctx := context.Background()

	// 1. MinIO bucket bootstrap. The real client lives behind objectstore.Client;
	// for the initial scaffold we use the FakeClient so a missing MinIO does
	// not block local boot. cmd/monti-api wiring swaps in the minio-go client.
	if err := objectstore.Bootstrap(ctx, &objectstore.FakeClient{}); err != nil {
		log.Printf("objectstore bootstrap: %v", err)
	}

	// 2. NATS stream + subject assertion is a smoke log here; the production
	// wiring uses nats.go to declare the HARVEST_MONTI stream if absent.
	log.Printf("nats stream=%s subjects=%v", events.StreamName, events.AllSubjects())

	// 3. Shared infrastructure.
	inv := &cache.NoopInvalidator{}
	pub := &events.InMemoryPublisher{}

	// 4. Plugin wiring (order ← cart ← promo loop).
	menuRepo := menu.NewMemoryRepo()
	cartRepo := cart.NewMemoryRepo()
	orderRepo := order.NewMemoryRepo()

	orderSvc := order.NewService(orderRepo, nil, cartRepo, inv, pub)
	cartSvc := cart.NewService(cartRepo, inv, pub, orderSvc.LookupPromo)
	orderSvc = order.NewService(orderRepo, cartSvc, cartRepo, inv, pub)

	menuSvc := menu.NewService(menuRepo, inv)

	menuH := menu.NewHandler(menuSvc)
	cartH := cart.NewHandler(cartSvc)
	orderH := order.NewHandler(orderSvc)

	// 5. Router.
	r := router.New()

	// Authenticated /api/v1 group (tenant + jwt).
	api := r.Group("/api/v1")
	api.Use(middleware.JWTAuth(), middleware.Tenant())
	menu.Register(api, menuH)
	cart.Register(api, cartH)
	order.Register(api, orderH)

	// Public /api/v1/public group (no auth).
	pubGroup := r.Group("/api/v1/public")
	menu.RegisterPublic(pubGroup, menuH)
	order.RegisterPublic(pubGroup, orderH)
	pubGroup.Post("/qr/redeem", router.QRRedeemHandler(router.PublicDeps{
		QRKey:        []byte(envOr("MONTI_QR_KEY", "dev-only-key-do-not-use")),
		JtiBlacklist: router.NewMemoryJtiStore(time.Now),
		Now:          time.Now,
	}))

	addr := envOr("MONTI_HTTP_ADDR", ":8081")
	log.Printf("monti-api listening on %s; %d routes registered", addr, len(r.Routes()))
	for _, rt := range r.Routes() {
		log.Printf("  %s", rt)
	}
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
