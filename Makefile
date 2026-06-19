# harvest-monti — Makefile
# Targets are intentionally thin; the seed target is owned by TASK-0013.

.PHONY: help seed seed-dryrun

help: ## Show this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-18s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

seed: ## Seed Monti Demo Bistro tenant + menu + images
	go run ./cmd/seed

seed-dryrun: ## Render placeholder images locally without touching Postgres/MinIO
	TEST_NO_INFRA=1 go run ./cmd/seed
