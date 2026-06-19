-- TASK-0009 / DES-0007 §2 rollback
-- 0004_order_lifecycle.sql rollback

BEGIN;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'payments') THEN
        RAISE EXCEPTION 'rollback 0004 refused: rollback 0005 must run first (payments table still present)';
    END IF;
END $$;

DROP TABLE IF EXISTS order_item_modifiers;
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;

DROP TYPE IF EXISTS fulfillment_status;

DELETE FROM schema_migrations WHERE version = '0004_order_lifecycle';

COMMIT;
