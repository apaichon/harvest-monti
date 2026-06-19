-- TASK-0009 / DES-0007 §2 rollback
-- 0005_payments_promos.sql rollback

BEGIN;

-- Guard: refuse if downstream payment rows exist (DES-0007 §2 policy)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM payments) THEN
        RAISE EXCEPTION 'rollback 0005 refused: payments rows exist (DES-0007 §2)';
    END IF;
END $$;

DROP TABLE IF EXISTS qr_table_codes;
DROP TABLE IF EXISTS order_promos;
DROP TABLE IF EXISTS promo_codes;
DROP TABLE IF EXISTS payments;

DROP TYPE IF EXISTS promo_status;
DROP TYPE IF EXISTS promo_kind;
DROP TYPE IF EXISTS payment_status;
DROP TYPE IF EXISTS payment_method;

DELETE FROM schema_migrations WHERE version = '0005_payments_promos';

COMMIT;
