-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Products table
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    normalized_name TEXT NOT NULL UNIQUE,
    category TEXT NOT NULL CHECK (category IN ('booster', 'display', 'etb', 'collection', 'bundle', 'tin', 'single')),
    set_name TEXT,
    set_code TEXT,
    msrp_eur DECIMAL(10,2) CHECK (msrp_eur > 0 OR msrp_eur IS NULL),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_products_normalized_name ON products(normalized_name);
CREATE INDEX idx_products_category ON products(category);
CREATE INDEX idx_products_set_code ON products(set_code);

-- Analyses table
CREATE TABLE analyses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    status TEXT NOT NULL DEFAULT 'running' CHECK (status IN ('running', 'completed', 'failed')),
    products_count INT DEFAULT 0,
    sales_count INT DEFAULT 0,
    search_query TEXT,
    error_message TEXT
);

CREATE INDEX idx_analyses_status ON analyses(status);
CREATE INDEX idx_analyses_started_at ON analyses(started_at DESC);

-- Sales table
CREATE TABLE sales (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    analysis_id UUID REFERENCES analyses(id) ON DELETE SET NULL,
    platform TEXT NOT NULL DEFAULT 'ebay',
    title TEXT NOT NULL,
    price DECIMAL(10,2) NOT NULL CHECK (price > 0),
    currency TEXT DEFAULT 'EUR',
    sold_at TIMESTAMPTZ NOT NULL,
    url TEXT UNIQUE,
    scraped_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sales_product_id ON sales(product_id);
CREATE INDEX idx_sales_sold_at ON sales(sold_at DESC);
CREATE INDEX idx_sales_analysis_id ON sales(analysis_id);

-- Updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger for products table
CREATE TRIGGER update_products_updated_at
    BEFORE UPDATE ON products
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Materialized view for product statistics
CREATE MATERIALIZED VIEW product_stats AS
SELECT
    p.id AS product_id,
    p.normalized_name,
    p.category,
    p.set_name,
    p.msrp_eur,
    COUNT(s.id) AS sales_count_30d,
    COALESCE(AVG(s.price), 0) AS avg_price,
    COALESCE(MIN(s.price), 0) AS min_price,
    COALESCE(MAX(s.price), 0) AS max_price,
    STDDEV(s.price) AS price_stddev,
    AVG(s.price) - COALESCE(p.msrp_eur, 0) AS margin_eur,
    CASE
        WHEN p.msrp_eur > 0 THEN ((AVG(s.price) - p.msrp_eur) / p.msrp_eur * 100)
        ELSE NULL
    END AS margin_percent,
    MAX(s.sold_at) AS last_sale_at
FROM products p
LEFT JOIN sales s ON s.product_id = p.id
    AND s.sold_at > NOW() - INTERVAL '30 days'
GROUP BY p.id, p.normalized_name, p.category, p.set_name, p.msrp_eur;

CREATE UNIQUE INDEX idx_product_stats_product_id ON product_stats(product_id);
CREATE INDEX idx_product_stats_sales_count ON product_stats(sales_count_30d DESC);
CREATE INDEX idx_product_stats_margin ON product_stats(margin_percent DESC NULLS LAST);
