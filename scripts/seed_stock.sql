-- Pass `-v reset=1` to also clear `processed_orders`, so order ids seen in
-- earlier runs aren't short-circuited by the idempotency guard.

\set ON_ERROR_STOP on

BEGIN;

\if :{?reset}
TRUNCATE processed_orders;
\endif

INSERT INTO stock (sku, available_quantity, reserved_quantity) VALUES
    ('SKU-1001', 500, 0),
    ('SKU-1002', 500, 0),
    ('SKU-1003', 300, 0),
    ('SKU-1004', 300, 0),
    ('SKU-1005', 200, 0),
    ('SKU-1006', 200, 0),
    ('SKU-1007', 150, 0),
    ('SKU-1008', 150, 0),
    ('SKU-1009', 400, 0),
    ('SKU-1010', 400, 0),
    ('SKU-2044', 100, 0),
    ('SKU-2045', 100, 0),
    ('SKU-2046', 100, 0),
    ('SKU-2047', 100, 0),
    ('SKU-2048', 100, 0),
    ('SKU-2049', 5, 0),
    ('SKU-2050', 3, 0),
    ('SKU-3954', 2, 0),
    ('SKU-3955', 0, 0)
ON CONFLICT (sku) DO UPDATE
    SET available_quantity = EXCLUDED.available_quantity,
        reserved_quantity  = EXCLUDED.reserved_quantity,
        updated_at         = NOW();

COMMIT;

SELECT sku, available_quantity FROM stock ORDER BY sku;
