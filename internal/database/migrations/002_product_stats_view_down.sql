-- Down migration: 002_product_stats_view
-- Rollback the materialized view refresh function

DROP FUNCTION IF EXISTS get_stats_summary();
DROP FUNCTION IF EXISTS refresh_product_stats();
