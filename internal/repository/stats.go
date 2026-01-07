package repository

import (
	"context"
	"fmt"

	"github.com/SubZzzzzz/pokeprofit/internal/database"
	"github.com/SubZzzzzz/pokeprofit/internal/models"
)

// StatsRepository handles aggregated statistics queries.
type StatsRepository struct {
	*Base
}

// NewStatsRepository creates a new StatsRepository.
func NewStatsRepository(db *database.DB) *StatsRepository {
	return &StatsRepository{
		Base: NewBase(db),
	}
}

// StatsOptions configures stats queries.
type StatsOptions struct {
	Category  string
	SortBy    string // "sales_count", "margin_percent", "avg_price"
	SortOrder string // "asc", "desc"
	MinSales  int
	Limit     int
	Offset    int
}

// DefaultStatsOptions returns default stats options.
func DefaultStatsOptions() StatsOptions {
	return StatsOptions{
		SortBy:    "sales_count",
		SortOrder: "desc",
		MinSales:  1,
		Limit:     10,
		Offset:    0,
	}
}

// GetProductStats returns volume and profit stats for all products.
// This query directly calculates stats without using a materialized view.
func (r *StatsRepository) GetProductStats(ctx context.Context, opts StatsOptions) ([]models.ProductStats, error) {
	// Set defaults
	if opts.SortBy == "" {
		opts.SortBy = "sales_count"
	}
	if opts.SortOrder == "" {
		opts.SortOrder = "desc"
	}
	if opts.Limit <= 0 {
		opts.Limit = 10
	}

	// Validate sort column to prevent SQL injection
	validSortColumns := map[string]string{
		"sales_count":    "sales_count_30d",
		"margin_percent": "margin_percent",
		"avg_price":      "avg_price",
	}
	sortColumn, ok := validSortColumns[opts.SortBy]
	if !ok {
		sortColumn = "sales_count_30d"
	}

	// Build the query
	query := `
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
			COALESCE(AVG(s.price), 0) - COALESCE(p.msrp_eur, 0) AS margin_eur,
			CASE
				WHEN p.msrp_eur > 0 THEN ((COALESCE(AVG(s.price), 0) - p.msrp_eur) / p.msrp_eur * 100)
				ELSE NULL
			END AS margin_percent,
			MAX(s.sold_at) AS last_sale_at
		FROM products p
		LEFT JOIN sales s ON s.product_id = p.id
			AND s.sold_at > NOW() - INTERVAL '30 days'
		WHERE 1=1
	`

	args := []interface{}{}
	argIndex := 1

	if opts.Category != "" {
		query += fmt.Sprintf(" AND p.category = $%d", argIndex)
		args = append(args, opts.Category)
		argIndex++
	}

	query += " GROUP BY p.id, p.normalized_name, p.category, p.set_name, p.msrp_eur"

	if opts.MinSales > 0 {
		query += fmt.Sprintf(" HAVING COUNT(s.id) >= $%d", argIndex)
		args = append(args, opts.MinSales)
		argIndex++
	}

	// Add ORDER BY with NULLS LAST for margin_percent
	if sortColumn == "margin_percent" {
		query += fmt.Sprintf(" ORDER BY %s %s NULLS LAST", sortColumn, opts.SortOrder)
	} else {
		query += fmt.Sprintf(" ORDER BY %s %s", sortColumn, opts.SortOrder)
	}

	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, opts.Limit, opts.Offset)

	var stats []models.ProductStats
	err := r.QueryRows(ctx, &stats, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get product stats: %w", err)
	}

	return stats, nil
}

// RefreshStats refreshes the materialized view (if using one).
// For direct query approach, this is a no-op but kept for interface compatibility.
func (r *StatsRepository) RefreshStats(ctx context.Context) error {
	// If using a materialized view, uncomment:
	// query := `REFRESH MATERIALIZED VIEW CONCURRENTLY product_stats`
	// _, err := r.Exec(ctx, query)
	// return err
	return nil
}

// GetTopProductsByVolume returns the top N products by sales volume.
func (r *StatsRepository) GetTopProductsByVolume(ctx context.Context, limit int) ([]models.ProductStats, error) {
	return r.GetProductStats(ctx, StatsOptions{
		SortBy:    "sales_count",
		SortOrder: "desc",
		MinSales:  1,
		Limit:     limit,
	})
}

// GetTopProductsByMargin returns the top N products by margin percentage.
func (r *StatsRepository) GetTopProductsByMargin(ctx context.Context, limit int) ([]models.ProductStats, error) {
	return r.GetProductStats(ctx, StatsOptions{
		SortBy:    "margin_percent",
		SortOrder: "desc",
		MinSales:  1,
		Limit:     limit,
	})
}

// GetStatsByCategory returns stats for a specific product category.
func (r *StatsRepository) GetStatsByCategory(ctx context.Context, category string, limit int) ([]models.ProductStats, error) {
	return r.GetProductStats(ctx, StatsOptions{
		Category:  category,
		SortBy:    "margin_percent",
		SortOrder: "desc",
		MinSales:  1,
		Limit:     limit,
	})
}

// GetTotalSalesCount returns the total number of sales in the last 30 days.
func (r *StatsRepository) GetTotalSalesCount(ctx context.Context) (int, error) {
	var count int
	query := `
		SELECT COUNT(*)
		FROM sales
		WHERE sold_at > NOW() - INTERVAL '30 days'
	`

	err := r.QueryRow(ctx, &count, query)
	if err != nil {
		return 0, fmt.Errorf("failed to get total sales count: %w", err)
	}

	return count, nil
}

// GetUniqueProductCount returns the number of products with sales in the last 30 days.
func (r *StatsRepository) GetUniqueProductCount(ctx context.Context) (int, error) {
	var count int
	query := `
		SELECT COUNT(DISTINCT product_id)
		FROM sales
		WHERE sold_at > NOW() - INTERVAL '30 days'
	`

	err := r.QueryRow(ctx, &count, query)
	if err != nil {
		return 0, fmt.Errorf("failed to get unique product count: %w", err)
	}

	return count, nil
}
