---
name: database
description: Patterns pour PostgreSQL en Go (sqlx, migrations, queries optimisées). Pour stockage des ventes, produits, alertes.
allowed-tools: Read, Grep, Glob, Edit, Write, Bash
---

# Database Patterns (PostgreSQL + Go)

## Stack
- **sqlx** : Extension de database/sql avec named queries
- **pgx** : Driver PostgreSQL performant
- **migrate** : Gestion des migrations
- **squirrel** : Query builder (optionnel)

## Structure de base

```go
package db

import (
    "github.com/jmoiron/sqlx"
    _ "github.com/lib/pq"
)

type Database struct {
    *sqlx.DB
}

func Connect(dsn string) (*Database, error) {
    db, err := sqlx.Connect("postgres", dsn)
    if err != nil {
        return nil, fmt.Errorf("connect: %w", err)
    }

    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(5)
    db.SetConnMaxLifetime(5 * time.Minute)

    return &Database{db}, nil
}
```

## Schéma pour ton projet

```sql
-- Products (cartes Pokemon)
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    set_name TEXT,
    card_number TEXT,
    rarity TEXT,
    tcgplayer_id TEXT UNIQUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_products_name ON products(name);
CREATE INDEX idx_products_tcgplayer ON products(tcgplayer_id);

-- Sales (ventes scrapées)
CREATE TABLE sales (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID REFERENCES products(id),
    platform TEXT NOT NULL, -- 'ebay', 'vinted', 'cardmarket'
    title TEXT NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    currency TEXT DEFAULT 'EUR',
    sold_at TIMESTAMPTZ NOT NULL,
    url TEXT,
    scraped_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_sales_product ON sales(product_id);
CREATE INDEX idx_sales_platform ON sales(platform);
CREATE INDEX idx_sales_sold_at ON sales(sold_at DESC);

-- Opportunities (opportunités identifiées)
CREATE TABLE opportunities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID REFERENCES products(id),
    type TEXT NOT NULL, -- 'volume', 'arbitrage', 'spike', 'restock'
    buy_price DECIMAL(10,2),
    sell_price DECIMAL(10,2),
    roi_percent DECIMAL(5,2),
    volume_last_7d INT,
    confidence_score DECIMAL(3,2), -- 0.00 to 1.00
    metadata JSONB,
    detected_at TIMESTAMPTZ DEFAULT NOW(),
    notified BOOLEAN DEFAULT FALSE
);

CREATE INDEX idx_opportunities_type ON opportunities(type);
CREATE INDEX idx_opportunities_roi ON opportunities(roi_percent DESC);
CREATE INDEX idx_opportunities_notified ON opportunities(notified) WHERE NOT notified;
```

## Repository Pattern

```go
type SalesRepository struct {
    db *Database
}

func NewSalesRepository(db *Database) *SalesRepository {
    return &SalesRepository{db: db}
}

// Insert avec conflit handling
func (r *SalesRepository) Create(ctx context.Context, sale *Sale) error {
    query := `
        INSERT INTO sales (product_id, platform, title, price, currency, sold_at, url)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        ON CONFLICT (url) DO NOTHING
        RETURNING id
    `
    return r.db.GetContext(ctx, &sale.ID, query,
        sale.ProductID, sale.Platform, sale.Title,
        sale.Price, sale.Currency, sale.SoldAt, sale.URL,
    )
}

// Bulk insert (plus performant pour scraping)
func (r *SalesRepository) BulkCreate(ctx context.Context, sales []Sale) error {
    query := `
        INSERT INTO sales (product_id, platform, title, price, currency, sold_at, url)
        VALUES (:product_id, :platform, :title, :price, :currency, :sold_at, :url)
        ON CONFLICT (url) DO NOTHING
    `
    _, err := r.db.NamedExecContext(ctx, query, sales)
    return err
}

// Volume analysis query
func (r *SalesRepository) GetVolumeStats(ctx context.Context, days int) ([]VolumeStats, error) {
    query := `
        SELECT
            p.id as product_id,
            p.name,
            COUNT(*) as sales_count,
            AVG(s.price) as avg_price,
            STDDEV(s.price) as price_stddev,
            MIN(s.price) as min_price,
            MAX(s.price) as max_price
        FROM sales s
        JOIN products p ON p.id = s.product_id
        WHERE s.sold_at > NOW() - INTERVAL '$1 days'
        GROUP BY p.id, p.name
        HAVING COUNT(*) >= 5
        ORDER BY sales_count DESC
    `
    var stats []VolumeStats
    err := r.db.SelectContext(ctx, &stats, query, days)
    return stats, err
}
```

## Transactions

```go
func (db *Database) WithTransaction(ctx context.Context, fn func(*sqlx.Tx) error) error {
    tx, err := db.BeginTxx(ctx, nil)
    if err != nil {
        return err
    }

    defer func() {
        if p := recover(); p != nil {
            tx.Rollback()
            panic(p)
        } else if err != nil {
            tx.Rollback()
        } else {
            err = tx.Commit()
        }
    }()

    err = fn(tx)
    return err
}
```

## Migrations (golang-migrate)

```bash
# Installer
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Créer une migration
migrate create -ext sql -dir migrations -seq create_sales_table

# Appliquer
migrate -path migrations -database "postgres://user:pass@localhost/dbname?sslmode=disable" up
```

## Best Practices

1. **Toujours utiliser context** pour timeout/cancellation
2. **Prepared statements** pour queries répétées
3. **Bulk inserts** pour scraping (pas de loop INSERT)
4. **Indexes** sur colonnes fréquemment filtrées
5. **JSONB** pour metadata flexible
6. **Partitioning** si plus de 10M rows (par date)
