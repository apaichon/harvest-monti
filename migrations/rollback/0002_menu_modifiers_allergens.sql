-- TASK-0009 / DES-0007 §2 rollback
-- 0002_menu_modifiers_allergens.sql rollback

BEGIN;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'cart_items') THEN
        RAISE EXCEPTION 'rollback 0002 refused: rollback 0003 must run first (cart_items table still present)';
    END IF;
END $$;

DROP TABLE IF EXISTS item_allergens;
DROP TABLE IF EXISTS allergens;
DROP TABLE IF EXISTS modifier_options;
DROP TABLE IF EXISTS modifier_groups;

DROP TYPE IF EXISTS modifier_kind;

DELETE FROM schema_migrations WHERE version = '0002_menu_modifiers_allergens';

COMMIT;
