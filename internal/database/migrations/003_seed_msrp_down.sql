-- Down migration: 003_seed_msrp
-- Note: This does not delete the products, only clears MSRP values
-- To avoid data loss during rollback

UPDATE products SET msrp_eur = NULL WHERE msrp_eur IS NOT NULL;

-- Refresh the materialized view
SELECT refresh_product_stats();
