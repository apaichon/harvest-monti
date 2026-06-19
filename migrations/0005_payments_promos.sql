-- TASK-0009 / DES-0007 §1, §3
-- 0005_payments_promos.sql — payments, promo_codes, order_promos, qr_table_codes

BEGIN;

DO $$ BEGIN
    CREATE TYPE payment_method AS ENUM ('credit_card', 'apple_pay', 'wallet', 'cash', 'stub');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE payment_status AS ENUM ('recorded', 'pending', 'failed', 'refunded');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE promo_kind   AS ENUM ('percent', 'fixed');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE promo_status AS ENUM ('active', 'expired', 'disabled');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

CREATE TABLE IF NOT EXISTS payments (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id      UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    method        payment_method NOT NULL DEFAULT 'stub',
    amount_cents  INT  NOT NULL CHECK (amount_cents >= 0),
    status        payment_status NOT NULL DEFAULT 'recorded',
    reference     TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS payments_order_idx ON payments(order_id);

CREATE TABLE IF NOT EXISTS promo_codes (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id     UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code          TEXT NOT NULL,
    kind          promo_kind NOT NULL,
    value         INT  NOT NULL CHECK (value >= 0),
    max_uses      INT  NOT NULL DEFAULT 0,
    used_count    INT  NOT NULL DEFAULT 0,
    expires_at    TIMESTAMPTZ,
    status        promo_status NOT NULL DEFAULT 'active',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (used_count <= max_uses OR max_uses = 0),
    CHECK (expires_at IS NULL OR expires_at > created_at)
);
CREATE INDEX IF NOT EXISTS promo_codes_tenant_idx ON promo_codes(tenant_id);
CREATE UNIQUE INDEX IF NOT EXISTS promo_codes_per_tenant ON promo_codes(tenant_id, code);

CREATE TABLE IF NOT EXISTS order_promos (
    id                 UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id           UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    promo_code_id      UUID NOT NULL REFERENCES promo_codes(id),
    discount_cents     INT  NOT NULL CHECK (discount_cents >= 0),
    kind               promo_kind NOT NULL
);
CREATE INDEX IF NOT EXISTS order_promos_order_idx ON order_promos(order_id);
CREATE INDEX IF NOT EXISTS order_promos_code_idx  ON order_promos(promo_code_id);
-- At most one percentage + one fixed per order (DES-0007 §1)
CREATE UNIQUE INDEX IF NOT EXISTS order_promos_one_per_kind ON order_promos(order_id, kind);

CREATE TABLE IF NOT EXISTS qr_table_codes (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    outlet_id   UUID NOT NULL REFERENCES outlets(id) ON DELETE CASCADE,
    table_code  TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS qr_table_codes_tenant_idx ON qr_table_codes(tenant_id);
CREATE INDEX IF NOT EXISTS qr_table_codes_outlet_idx ON qr_table_codes(outlet_id);
CREATE UNIQUE INDEX IF NOT EXISTS qr_table_codes_per_outlet ON qr_table_codes(outlet_id, table_code);

INSERT INTO schema_migrations(version) VALUES ('0005_payments_promos')
    ON CONFLICT DO NOTHING;

COMMIT;
