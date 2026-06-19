-- TASK-0009 / DES-0007 §1, §3
-- 0004_order_lifecycle.sql

BEGIN;

DO $$ BEGIN
    CREATE TYPE fulfillment_status AS ENUM ('received', 'preparing', 'ready', 'completed', 'cancelled');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

CREATE TABLE IF NOT EXISTS orders (
    id                     UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id              UUID NOT NULL REFERENCES tenants(id),
    outlet_id              UUID NOT NULL REFERENCES outlets(id),
    table_code             TEXT,
    cart_id                UUID REFERENCES carts(id),
    order_number           TEXT NOT NULL,
    subtotal_cents         INT  NOT NULL CHECK (subtotal_cents >= 0),
    service_charge_cents   INT  NOT NULL DEFAULT 0 CHECK (service_charge_cents >= 0),
    tax_cents              INT  NOT NULL DEFAULT 0 CHECK (tax_cents >= 0),
    promo_discount_cents   INT  NOT NULL DEFAULT 0 CHECK (promo_discount_cents >= 0),
    total_cents            INT  NOT NULL CHECK (total_cents >= 0),
    currency               TEXT NOT NULL DEFAULT 'THB',
    status                 fulfillment_status NOT NULL DEFAULT 'received',
    placed_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    ready_eta_at           TIMESTAMPTZ,
    CONSTRAINT orders_total_formula
        CHECK (total_cents = subtotal_cents + service_charge_cents + tax_cents - promo_discount_cents)
);
CREATE INDEX IF NOT EXISTS orders_tenant_idx    ON orders(tenant_id);
CREATE INDEX IF NOT EXISTS orders_outlet_idx    ON orders(outlet_id);
CREATE INDEX IF NOT EXISTS orders_cart_idx      ON orders(cart_id);
CREATE INDEX IF NOT EXISTS orders_dashboard     ON orders(tenant_id, status, placed_at DESC);
CREATE UNIQUE INDEX IF NOT EXISTS orders_number_per_tenant ON orders(tenant_id, order_number);

CREATE TABLE IF NOT EXISTS order_items (
    id                   UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id             UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    item_id              UUID NOT NULL REFERENCES menu_items(id),
    qty                  INT  NOT NULL CHECK (qty > 0),
    unit_price_cents     INT  NOT NULL CHECK (unit_price_cents >= 0),
    notes                TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS order_items_order_idx ON order_items(order_id);
CREATE INDEX IF NOT EXISTS order_items_item_idx  ON order_items(item_id);

CREATE TABLE IF NOT EXISTS order_item_modifiers (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_item_id       UUID NOT NULL REFERENCES order_items(id) ON DELETE CASCADE,
    modifier_option_id  UUID NOT NULL REFERENCES modifier_options(id),
    price_delta_cents   INT  NOT NULL DEFAULT 0,
    CHECK (price_delta_cents >= -10000)
);
CREATE INDEX IF NOT EXISTS order_item_modifiers_line_idx ON order_item_modifiers(order_item_id);
CREATE INDEX IF NOT EXISTS order_item_modifiers_opt_idx  ON order_item_modifiers(modifier_option_id);

INSERT INTO schema_migrations(version) VALUES ('0004_order_lifecycle')
    ON CONFLICT DO NOTHING;

COMMIT;
