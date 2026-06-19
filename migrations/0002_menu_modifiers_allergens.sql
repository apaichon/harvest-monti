-- TASK-0009 / DES-0007 §1, §3
-- 0002_menu_modifiers_allergens.sql

BEGIN;

DO $$ BEGIN
    CREATE TYPE modifier_kind AS ENUM ('single', 'multi');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

CREATE TABLE IF NOT EXISTS modifier_groups (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    item_id      UUID NOT NULL REFERENCES menu_items(id) ON DELETE CASCADE,
    name         TEXT NOT NULL,
    kind         modifier_kind NOT NULL,
    required     BOOLEAN NOT NULL DEFAULT false,
    min_select   INT NOT NULL DEFAULT 0,
    max_select   INT NOT NULL DEFAULT 1,
    display_order INT NOT NULL DEFAULT 0,
    CHECK (min_select >= 0),
    CHECK (min_select <= max_select),
    CHECK (kind <> 'single' OR max_select = 1)
);
CREATE INDEX IF NOT EXISTS modifier_groups_item_idx ON modifier_groups(item_id);

CREATE TABLE IF NOT EXISTS modifier_options (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    group_id            UUID NOT NULL REFERENCES modifier_groups(id) ON DELETE CASCADE,
    name                TEXT NOT NULL,
    price_delta_cents   INT NOT NULL DEFAULT 0,
    display_order       INT NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS modifier_options_group_idx ON modifier_options(group_id);

CREATE TABLE IF NOT EXISTS allergens (
    id        UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code      TEXT NOT NULL,
    label_en  TEXT,
    label_th  TEXT,
    label_zh  TEXT,
    label_ja  TEXT
);
CREATE UNIQUE INDEX IF NOT EXISTS allergens_code_uniq ON allergens(code);

CREATE TABLE IF NOT EXISTS item_allergens (
    item_id      UUID NOT NULL REFERENCES menu_items(id) ON DELETE CASCADE,
    allergen_id  UUID NOT NULL REFERENCES allergens(id)  ON DELETE CASCADE,
    PRIMARY KEY (item_id, allergen_id)
);
CREATE INDEX IF NOT EXISTS item_allergens_allergen_idx ON item_allergens(allergen_id);

INSERT INTO schema_migrations(version) VALUES ('0002_menu_modifiers_allergens')
    ON CONFLICT DO NOTHING;

COMMIT;
