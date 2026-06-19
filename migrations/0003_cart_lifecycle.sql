-- TASK-0009 / DES-0007 §1, §3
-- 0003_cart_lifecycle.sql

BEGIN;

DO $$ BEGIN
    CREATE TYPE cart_status AS ENUM ('open', 'checking_out', 'submitted', 'abandoned');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

CREATE TABLE IF NOT EXISTS carts (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    outlet_id   UUID NOT NULL REFERENCES outlets(id) ON DELETE CASCADE,
    table_code  TEXT,
    session_id  TEXT NOT NULL,
    status      cart_status NOT NULL DEFAULT 'open',
    promo_code  TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS carts_tenant_idx  ON carts(tenant_id);
CREATE INDEX IF NOT EXISTS carts_outlet_idx  ON carts(outlet_id);
CREATE INDEX IF NOT EXISTS carts_session_idx ON carts(session_id);

CREATE TABLE IF NOT EXISTS cart_items (
    id                   UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    cart_id              UUID NOT NULL REFERENCES carts(id) ON DELETE CASCADE,
    item_id              UUID NOT NULL REFERENCES menu_items(id),
    qty                  INT  NOT NULL,
    unit_price_snapshot  INT  NOT NULL CHECK (unit_price_snapshot >= 0),
    notes                TEXT NOT NULL DEFAULT '',
    line_key             TEXT NOT NULL DEFAULT '',
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT cart_items_qty_check CHECK (qty > 0)
);
CREATE INDEX IF NOT EXISTS cart_items_cart_idx ON cart_items(cart_id);
CREATE INDEX IF NOT EXISTS cart_items_item_idx ON cart_items(item_id);
CREATE UNIQUE INDEX IF NOT EXISTS cart_items_dedupe
    ON cart_items(cart_id, item_id, line_key) WHERE qty > 0;

CREATE TABLE IF NOT EXISTS cart_item_modifiers (
    id                     UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    cart_item_id           UUID NOT NULL REFERENCES cart_items(id) ON DELETE CASCADE,
    modifier_option_id     UUID NOT NULL REFERENCES modifier_options(id),
    price_delta_snapshot   INT  NOT NULL DEFAULT 0,
    CHECK (price_delta_snapshot >= -10000)
);
CREATE INDEX IF NOT EXISTS cart_item_modifiers_line_idx ON cart_item_modifiers(cart_item_id);
CREATE INDEX IF NOT EXISTS cart_item_modifiers_opt_idx  ON cart_item_modifiers(modifier_option_id);

INSERT INTO schema_migrations(version) VALUES ('0003_cart_lifecycle')
    ON CONFLICT DO NOTHING;

COMMIT;
