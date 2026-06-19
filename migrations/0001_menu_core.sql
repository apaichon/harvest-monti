-- TASK-0009 / DES-0007 §1, §3
-- 0001_menu_core.sql — tenants, outlets, menus, menu_categories, menu_items + enums

BEGIN;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

DO $$ BEGIN
    CREATE TYPE menu_status     AS ENUM ('draft', 'active', 'archived');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE category_status AS ENUM ('active', 'hidden', 'archived');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE item_status     AS ENUM ('active', 'sold_out', 'hidden', 'archived');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

CREATE TABLE IF NOT EXISTS tenants (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        TEXT NOT NULL,
    locale      TEXT NOT NULL DEFAULT 'en',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS outlets (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id    UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name         TEXT NOT NULL,
    table_count  INT  NOT NULL DEFAULT 0 CHECK (table_count >= 0),
    timezone     TEXT NOT NULL DEFAULT 'Asia/Bangkok',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS outlets_tenant_idx ON outlets(tenant_id);

CREATE TABLE IF NOT EXISTS menus (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    outlet_id       UUID NOT NULL REFERENCES outlets(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    status          menu_status NOT NULL DEFAULT 'draft',
    effective_from  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS menus_outlet_idx ON menus(outlet_id);

CREATE TABLE IF NOT EXISTS menu_categories (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    menu_id           UUID NOT NULL REFERENCES menus(id) ON DELETE CASCADE,
    name              TEXT NOT NULL,
    display_order     INT  NOT NULL DEFAULT 0,
    image_minio_path  TEXT,
    status            category_status NOT NULL DEFAULT 'active'
);
CREATE INDEX IF NOT EXISTS menu_categories_menu_idx ON menu_categories(menu_id);

CREATE TABLE IF NOT EXISTS menu_items (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    menu_id           UUID NOT NULL REFERENCES menus(id) ON DELETE CASCADE,
    category_id       UUID REFERENCES menu_categories(id) ON DELETE SET NULL,
    sku               TEXT,
    name              TEXT NOT NULL,
    description       TEXT,
    price_cents       INT  NOT NULL CHECK (price_cents >= 0),
    currency          TEXT NOT NULL DEFAULT 'THB',
    image_minio_path  TEXT,
    calories          INT,
    is_best_seller    BOOLEAN NOT NULL DEFAULT false,
    display_order     INT  NOT NULL DEFAULT 0,
    status            item_status NOT NULL DEFAULT 'active',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS menu_items_menu_idx     ON menu_items(menu_id);
CREATE INDEX IF NOT EXISTS menu_items_category_idx ON menu_items(category_id);
CREATE INDEX IF NOT EXISTS menu_items_browse       ON menu_items(menu_id, category_id, status);

INSERT INTO schema_migrations(version) VALUES ('0001_menu_core')
    ON CONFLICT DO NOTHING;

COMMIT;
