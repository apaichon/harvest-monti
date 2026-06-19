-- TASK-0009 / DES-0007 §2 rollback
-- 0001_menu_core.sql rollback

BEGIN;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'modifier_groups') THEN
        RAISE EXCEPTION 'rollback 0001 refused: rollback 0002 must run first (modifier_groups table still present)';
    END IF;
END $$;

DROP TABLE IF EXISTS menu_items;
DROP TABLE IF EXISTS menu_categories;
DROP TABLE IF EXISTS menus;
DROP TABLE IF EXISTS outlets;
DROP TABLE IF EXISTS tenants;

DROP TYPE IF EXISTS item_status;
DROP TYPE IF EXISTS category_status;
DROP TYPE IF EXISTS menu_status;

DELETE FROM schema_migrations WHERE version = '0001_menu_core';

COMMIT;
