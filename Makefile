# harvest-monti — Makefile
# Targets are intentionally thin. TASK-0013 owns `seed`; TASK-0014 owns the
# `e2e` family below.

.PHONY: help seed seed-dryrun e2e e2e-up e2e-down e2e-fixtures e2e-dry \
        e2e-live e2e-mobile e2e-clean

COMPOSE      ?= docker compose
COMPOSE_FILE ?= compose.e2e.yml

help: ## Show this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-18s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# ---- seed (TASK-0013) ----

seed: ## Seed Monti Demo Bistro tenant + menu + images
	go run ./cmd/seed

seed-dryrun: ## Render placeholder images locally without touching Postgres/MinIO
	TEST_NO_INFRA=1 go run ./cmd/seed

# ---- e2e (TASK-0014) ----

e2e: e2e-up e2e-fixtures e2e-down ## Full stack + fixture suite + evidence capture (requires Docker)

e2e-up: ## Bring up compose.e2e.yml and wait for healthy services
	$(COMPOSE) -f $(COMPOSE_FILE) up -d --build
	@echo "waiting for monti-gateway healthcheck..."
	@for i in $$(seq 1 60); do \
	    state=$$($(COMPOSE) -f $(COMPOSE_FILE) ps --format json monti-gateway 2>/dev/null | grep -o '"Health":"[a-z]*"' || true); \
	    if echo "$$state" | grep -q '"Health":"healthy"'; then echo "gateway healthy"; exit 0; fi; \
	    sleep 2; \
	done; \
	echo "gateway never reached healthy"; exit 1

e2e-down: ## Tear down compose.e2e.yml
	$(COMPOSE) -f $(COMPOSE_FILE) down -v --remove-orphans

e2e-fixtures: ## Run the fixture suite against a running stack (stack mode)
	MONTI_E2E_STACK=1 go test ./e2e/... -count=1 -timeout=2m

e2e-dry: ## Run the fixture suite against in-memory fakes (no Docker required)
	go test ./e2e/... -count=1 -timeout=2m

e2e-live: ## Flagged live-Gemini smoke (requires GEMINI_API_KEY + LIVE_VOICE=1)
	LIVE_VOICE=1 MONTI_E2E_STACK=1 go test ./e2e/... -run TestLiveGeminiSmoke -count=1 -timeout=2m

e2e-mobile: ## Flutter integration_test against the running stack
	cd mobile/customer && flutter test integration_test/full_flow_test.dart

e2e-clean: ## Remove evidence artifacts
	rm -rf ../../docs/harvest-monti/sdlc/05-tests/evidence/TEST-*/case-*

# ---- aggregate ----

test: ## Run all Go tests (unit + e2e dry)
	go test ./...
