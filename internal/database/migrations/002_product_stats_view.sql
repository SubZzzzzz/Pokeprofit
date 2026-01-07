-- Migration: 002_product_stats_view
-- Purpose: Add function to refresh product_stats materialized view
-- Date: 2026-01-07
-- Feature: User Story 2 - Calculate Product Profitability

-- Function to refresh the product_stats materialized view
-- Can be called after each analysis or via scheduled job
CREATE OR REPLACE FUNCTION refresh_product_stats()
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY product_stats;
END;
$$ LANGUAGE plpgsql;

-- Grant execute permission (adjust role as needed)
-- GRANT EXECUTE ON FUNCTION refresh_product_stats() TO pokeprofit_app;

-- Add index to support concurrent refresh (requires unique index)
-- Note: idx_product_stats_product_id already exists from 001_initial_schema.sql

-- Create a helper function to get stats summary
CREATE OR REPLACE FUNCTION get_stats_summary()
RETURNS TABLE (
    total_products INT,
    products_with_sales INT,
    total_sales_30d BIGINT,
    avg_margin_percent DECIMAL(10,2)
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        COUNT(DISTINCT ps.product_id)::INT AS total_products,
        COUNT(DISTINCT CASE WHEN ps.sales_count_30d > 0 THEN ps.product_id END)::INT AS products_with_sales,
        COALESCE(SUM(ps.sales_count_30d), 0) AS total_sales_30d,
        COALESCE(AVG(ps.margin_percent) FILTER (WHERE ps.margin_percent IS NOT NULL), 0)::DECIMAL(10,2) AS avg_margin_percent
    FROM product_stats ps;
END;
$$ LANGUAGE plpgsql;
