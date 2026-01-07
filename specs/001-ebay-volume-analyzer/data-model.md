# Data Model: Volume Analyzer Phase 1 - eBay

**Date**: 2026-01-07
**Feature Branch**: `001-ebay-volume-analyzer`

## Entity Relationship Diagram

```
┌─────────────────┐       ┌─────────────────┐       ┌─────────────────┐
│    products     │       │     sales       │       │   analyses      │
├─────────────────┤       ├─────────────────┤       ├─────────────────┤
│ id (PK)         │◄──────│ product_id (FK) │       │ id (PK)         │
│ normalized_name │       │ id (PK)         │       │ started_at      │
│ category        │       │ analysis_id(FK) │───────│ completed_at    │
│ set_name        │       │ platform        │       │ status          │
│ msrp_eur        │       │ title           │       │ products_count  │
│ created_at      │       │ price           │       │ sales_count     │
│ updated_at      │       │ sold_at         │       │ search_query    │
└─────────────────┘       │ url             │       │ error_message   │
                          │ scraped_at      │       └─────────────────┘
                          └─────────────────┘

┌─────────────────────────────────────────────────────────────────────┐
│                        product_stats (view)                         │
├─────────────────────────────────────────────────────────────────────┤
│ product_id, normalized_name, category, msrp_eur,                    │
│ sales_count_30d, avg_price, min_price, max_price, margin_eur,       │
│ margin_percent, last_sale_at                                        │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Entities

### 1. Product

Represents a normalized Pokemon TCG product (grouping similar eBay listings).

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| id | UUID | PK, DEFAULT gen_random_uuid() | Unique identifier |
| normalized_name | TEXT | NOT NULL, UNIQUE | Canonical name (e.g., "Display SV 151") |
| category | TEXT | NOT NULL | Product type: booster, display, etb, collection, bundle, tin, single |
| set_name | TEXT | | Pokemon set name (e.g., "Écarlate et Violet 151") |
| set_code | TEXT | | Short code (e.g., "sv-151", "sv-paldean-fates") |
| msrp_eur | DECIMAL(10,2) | | Official retail price in EUR |
| created_at | TIMESTAMPTZ | DEFAULT NOW() | Record creation timestamp |
| updated_at | TIMESTAMPTZ | DEFAULT NOW() | Last update timestamp |

**Indexes**:
- `idx_products_normalized_name` on `normalized_name`
- `idx_products_category` on `category`
- `idx_products_set_code` on `set_code`

**Validation Rules**:
- `category` must be one of: `booster`, `display`, `etb`, `collection`, `bundle`, `tin`, `single`
- `msrp_eur` must be > 0 when set

---

### 2. Sale

Represents a completed sale scraped from eBay.

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| id | UUID | PK, DEFAULT gen_random_uuid() | Unique identifier |
| product_id | UUID | FK → products.id, NOT NULL | Associated product |
| analysis_id | UUID | FK → analyses.id | Analysis session that found this sale |
| platform | TEXT | NOT NULL, DEFAULT 'ebay' | Source platform |
| title | TEXT | NOT NULL | Original eBay listing title |
| price | DECIMAL(10,2) | NOT NULL | Sale price in EUR |
| currency | TEXT | DEFAULT 'EUR' | Currency code |
| sold_at | TIMESTAMPTZ | NOT NULL | Date when item was sold |
| url | TEXT | UNIQUE | eBay listing URL (dedup key) |
| scraped_at | TIMESTAMPTZ | DEFAULT NOW() | When the sale was scraped |

**Indexes**:
- `idx_sales_product_id` on `product_id`
- `idx_sales_sold_at` on `sold_at DESC`
- `idx_sales_url` on `url` (unique)
- `idx_sales_analysis_id` on `analysis_id`

**Validation Rules**:
- `price` must be > 0
- `sold_at` must be within last 30 days (for volume analyzer)
- `url` must be unique (prevents duplicate scraping)

---

### 3. Analysis

Represents an analysis session (one run of the scraper).

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| id | UUID | PK, DEFAULT gen_random_uuid() | Unique identifier |
| started_at | TIMESTAMPTZ | DEFAULT NOW() | When analysis started |
| completed_at | TIMESTAMPTZ | | When analysis completed |
| status | TEXT | NOT NULL, DEFAULT 'running' | Status: running, completed, failed |
| products_count | INT | DEFAULT 0 | Number of distinct products found |
| sales_count | INT | DEFAULT 0 | Number of sales scraped |
| search_query | TEXT | | Search query used |
| error_message | TEXT | | Error details if failed |

**Indexes**:
- `idx_analyses_status` on `status`
- `idx_analyses_started_at` on `started_at DESC`

**State Transitions**:
```
running → completed (success)
running → failed (error)
```

---

## Views

### product_stats

Materialized view for efficient product statistics queries.

```sql
CREATE MATERIALIZED VIEW product_stats AS
SELECT
    p.id AS product_id,
    p.normalized_name,
    p.category,
    p.set_name,
    p.msrp_eur,
    COUNT(s.id) AS sales_count_30d,
    AVG(s.price) AS avg_price,
    MIN(s.price) AS min_price,
    MAX(s.price) AS max_price,
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
```

**Refresh**: Daily via scheduled job or after each analysis completion.

---

## SQL Schema

```sql
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

-- Updated_at trigger
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_products_updated_at
    BEFORE UPDATE ON products
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

---

## Go Models

```go
package models

import (
    "time"

    "github.com/google/uuid"
    "github.com/shopspring/decimal"
)

// ProductCategory represents the type of Pokemon TCG product
type ProductCategory string

const (
    CategoryBooster    ProductCategory = "booster"
    CategoryDisplay    ProductCategory = "display"
    CategoryETB        ProductCategory = "etb"
    CategoryCollection ProductCategory = "collection"
    CategoryBundle     ProductCategory = "bundle"
    CategoryTin        ProductCategory = "tin"
    CategorySingle     ProductCategory = "single"
)

// Product represents a normalized Pokemon TCG product
type Product struct {
    ID             uuid.UUID        `db:"id" json:"id"`
    NormalizedName string           `db:"normalized_name" json:"normalized_name"`
    Category       ProductCategory  `db:"category" json:"category"`
    SetName        *string          `db:"set_name" json:"set_name,omitempty"`
    SetCode        *string          `db:"set_code" json:"set_code,omitempty"`
    MSRPEUR        *decimal.Decimal `db:"msrp_eur" json:"msrp_eur,omitempty"`
    CreatedAt      time.Time        `db:"created_at" json:"created_at"`
    UpdatedAt      time.Time        `db:"updated_at" json:"updated_at"`
}

// AnalysisStatus represents the state of an analysis session
type AnalysisStatus string

const (
    StatusRunning   AnalysisStatus = "running"
    StatusCompleted AnalysisStatus = "completed"
    StatusFailed    AnalysisStatus = "failed"
)

// Analysis represents an analysis session
type Analysis struct {
    ID            uuid.UUID      `db:"id" json:"id"`
    StartedAt     time.Time      `db:"started_at" json:"started_at"`
    CompletedAt   *time.Time     `db:"completed_at" json:"completed_at,omitempty"`
    Status        AnalysisStatus `db:"status" json:"status"`
    ProductsCount int            `db:"products_count" json:"products_count"`
    SalesCount    int            `db:"sales_count" json:"sales_count"`
    SearchQuery   *string        `db:"search_query" json:"search_query,omitempty"`
    ErrorMessage  *string        `db:"error_message" json:"error_message,omitempty"`
}

// Sale represents a completed sale from eBay
type Sale struct {
    ID         uuid.UUID       `db:"id" json:"id"`
    ProductID  uuid.UUID       `db:"product_id" json:"product_id"`
    AnalysisID *uuid.UUID      `db:"analysis_id" json:"analysis_id,omitempty"`
    Platform   string          `db:"platform" json:"platform"`
    Title      string          `db:"title" json:"title"`
    Price      decimal.Decimal `db:"price" json:"price"`
    Currency   string          `db:"currency" json:"currency"`
    SoldAt     time.Time       `db:"sold_at" json:"sold_at"`
    URL        *string         `db:"url" json:"url,omitempty"`
    ScrapedAt  time.Time       `db:"scraped_at" json:"scraped_at"`
}

// ProductStats represents aggregated statistics for a product
type ProductStats struct {
    ProductID      uuid.UUID        `db:"product_id" json:"product_id"`
    NormalizedName string           `db:"normalized_name" json:"normalized_name"`
    Category       ProductCategory  `db:"category" json:"category"`
    SetName        *string          `db:"set_name" json:"set_name,omitempty"`
    MSRPEUR        *decimal.Decimal `db:"msrp_eur" json:"msrp_eur,omitempty"`
    SalesCount30d  int              `db:"sales_count_30d" json:"sales_count_30d"`
    AvgPrice       decimal.Decimal  `db:"avg_price" json:"avg_price"`
    MinPrice       decimal.Decimal  `db:"min_price" json:"min_price"`
    MaxPrice       decimal.Decimal  `db:"max_price" json:"max_price"`
    PriceStddev    *decimal.Decimal `db:"price_stddev" json:"price_stddev,omitempty"`
    MarginEUR      *decimal.Decimal `db:"margin_eur" json:"margin_eur,omitempty"`
    MarginPercent  *decimal.Decimal `db:"margin_percent" json:"margin_percent,omitempty"`
    LastSaleAt     *time.Time       `db:"last_sale_at" json:"last_sale_at,omitempty"`
}
```

---

## Seed Data (MSRP Reference)

Initial MSRP values for Pokemon TCG products (2024-2025 sets):

```sql
INSERT INTO products (normalized_name, category, set_name, set_code, msrp_eur) VALUES
-- Displays (Booster Boxes)
('Display Écarlate et Violet 151', 'display', 'Écarlate et Violet 151', 'sv-151', 159.99),
('Display Destinées de Paldea', 'display', 'Destinées de Paldea', 'sv-paldean-fates', 159.99),
('Display Évolutions Prismatiques', 'display', 'Évolutions Prismatiques', 'sv-prismatic-evo', 159.99),
('Display Masques du Crépuscule', 'display', 'Masques du Crépuscule', 'sv-twilight-masque', 159.99),
('Display Forces Temporelles', 'display', 'Forces Temporelles', 'sv-temporal-forces', 159.99),

-- ETBs (Elite Trainer Boxes)
('ETB Écarlate et Violet 151', 'etb', 'Écarlate et Violet 151', 'sv-151', 54.99),
('ETB Destinées de Paldea', 'etb', 'Destinées de Paldea', 'sv-paldean-fates', 54.99),
('ETB Évolutions Prismatiques', 'etb', 'Évolutions Prismatiques', 'sv-prismatic-evo', 54.99),

-- Collections / Coffrets
('Coffret Dracaufeu Ultra Premium', 'collection', 'Écarlate et Violet', 'sv-charizard-upc', 119.99),
('Coffret Mew Ultra Premium', 'collection', 'Écarlate et Violet 151', 'sv-151-upc', 119.99)
ON CONFLICT (normalized_name) DO UPDATE SET
    msrp_eur = EXCLUDED.msrp_eur,
    updated_at = NOW();
```
