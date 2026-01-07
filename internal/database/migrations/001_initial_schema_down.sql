-- Drop materialized view
DROP MATERIALIZED VIEW IF EXISTS product_stats;

-- Drop trigger
DROP TRIGGER IF EXISTS update_products_updated_at ON products;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables in reverse order
DROP TABLE IF EXISTS sales;
DROP TABLE IF EXISTS analyses;
DROP TABLE IF EXISTS products;
