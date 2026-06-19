-- TASK-0009 / DES-0007 §2 rollback
-- 0003_cart_lifecycle.sql rollback

BEGIN;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'orders') THEN
        RAISE EXCEPTION 'rollback 0003 refused: rollback 0004 must run first (orders table still present)';
    END IF;
END $$;

DROP TABLE IF EXISTS cart_item_modifiers;
DROP TABLE IF EXISTS cart_items;
DROP TABLE IF EXISTS carts;

DROP TYPE IF EXISTS cart_status;

DELETE FROM schema_migrations WHERE version = '0003_cart_lifecycle';

COMMIT;
