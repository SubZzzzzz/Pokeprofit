package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/SubZzzzzz/pokeprofit/internal/database"
	apperrors "github.com/SubZzzzzz/pokeprofit/internal/errors"
	"github.com/SubZzzzzz/pokeprofit/internal/models"
	"github.com/google/uuid"
)

// SaleRepository handles sale data persistence.
type SaleRepository struct {
	*Base
}

// NewSaleRepository creates a new SaleRepository.
func NewSaleRepository(db *database.DB) *SaleRepository {
	return &SaleRepository{
		Base: NewBase(db),
	}
}

// SaleListOptions configures sale list queries.
type SaleListOptions struct {
	Since  time.Time
	Limit  int
	Offset int
}

// Create inserts a new sale (ignores if URL already exists).
func (r *SaleRepository) Create(ctx context.Context, sale *models.Sale) error {
	query := `
		INSERT INTO sales (id, product_id, analysis_id, platform, title, price, currency, sold_at, url, scraped_at)
		VALUES (:id, :product_id, :analysis_id, :platform, :title, :price, :currency, :sold_at, :url, :scraped_at)
		ON CONFLICT (url) DO NOTHING
	`

	return r.InsertNamed(ctx, query, sale)
}

// BulkCreate inserts multiple sales efficiently.
// Returns the number of sales actually inserted (excludes duplicates).
func (r *SaleRepository) BulkCreate(ctx context.Context, sales []models.Sale) (int, error) {
	if len(sales) == 0 {
		return 0, nil
	}

	// Build bulk insert query
	valueStrings := make([]string, 0, len(sales))
	valueArgs := make([]interface{}, 0, len(sales)*10)

	for i, sale := range sales {
		base := i * 10
		valueStrings = append(valueStrings, fmt.Sprintf(
			"($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			base+1, base+2, base+3, base+4, base+5, base+6, base+7, base+8, base+9, base+10,
		))
		valueArgs = append(valueArgs,
			sale.ID,
			sale.ProductID,
			sale.AnalysisID,
			sale.Platform,
			sale.Title,
			sale.Price,
			sale.Currency,
			sale.SoldAt,
			sale.URL,
			sale.ScrapedAt,
		)
	}

	query := fmt.Sprintf(`
		INSERT INTO sales (id, product_id, analysis_id, platform, title, price, currency, sold_at, url, scraped_at)
		VALUES %s
		ON CONFLICT (url) DO NOTHING
	`, strings.Join(valueStrings, ","))

	result, err := r.DB().ExecContext(ctx, query, valueArgs...)
	if err != nil {
		return 0, fmt.Errorf("failed to bulk create sales: %w", err)
	}

	inserted, _ := result.RowsAffected()
	return int(inserted), nil
}

// FindByID retrieves a sale by ID.
func (r *SaleRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Sale, error) {
	var sale models.Sale
	query := `
		SELECT id, product_id, analysis_id, platform, title, price, currency, sold_at, url, scraped_at
		FROM sales
		WHERE id = $1
	`

	err := r.QueryRow(ctx, &sale, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.Wrap(apperrors.ErrNoData, "sale not found")
		}
		return nil, fmt.Errorf("failed to find sale by ID: %w", err)
	}

	return &sale, nil
}

// FindByProductID returns sales for a product.
func (r *SaleRepository) FindByProductID(ctx context.Context, productID uuid.UUID, opts SaleListOptions) ([]models.Sale, error) {
	var sales []models.Sale

	query := `
		SELECT id, product_id, analysis_id, platform, title, price, currency, sold_at, url, scraped_at
		FROM sales
		WHERE product_id = $1
	`
	args := []interface{}{productID}
	argIndex := 2

	if !opts.Since.IsZero() {
		query += fmt.Sprintf(" AND sold_at >= $%d", argIndex)
		args = append(args, opts.Since)
		argIndex++
	}

	query += " ORDER BY sold_at DESC"

	if opts.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, opts.Limit)
		argIndex++
	}

	if opts.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, opts.Offset)
	}

	err := r.QueryRows(ctx, &sales, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find sales by product ID: %w", err)
	}

	return sales, nil
}

// FindByAnalysisID returns sales for an analysis session.
func (r *SaleRepository) FindByAnalysisID(ctx context.Context, analysisID uuid.UUID) ([]models.Sale, error) {
	var sales []models.Sale
	query := `
		SELECT id, product_id, analysis_id, platform, title, price, currency, sold_at, url, scraped_at
		FROM sales
		WHERE analysis_id = $1
		ORDER BY sold_at DESC
	`

	err := r.QueryRows(ctx, &sales, query, analysisID)
	if err != nil {
		return nil, fmt.Errorf("failed to find sales by analysis ID: %w", err)
	}

	return sales, nil
}

// CountByProductID returns the number of sales for a product.
func (r *SaleRepository) CountByProductID(ctx context.Context, productID uuid.UUID, since time.Time) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM sales WHERE product_id = $1`
	args := []interface{}{productID}

	if !since.IsZero() {
		query += " AND sold_at >= $2"
		args = append(args, since)
	}

	err := r.QueryRow(ctx, &count, query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to count sales: %w", err)
	}

	return count, nil
}

// CountByAnalysisID returns the number of sales for an analysis.
func (r *SaleRepository) CountByAnalysisID(ctx context.Context, analysisID uuid.UUID) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM sales WHERE analysis_id = $1`

	err := r.QueryRow(ctx, &count, query, analysisID)
	if err != nil {
		return 0, fmt.Errorf("failed to count sales by analysis: %w", err)
	}

	return count, nil
}

// ExistsByURL checks if a sale with the given URL exists.
func (r *SaleRepository) ExistsByURL(ctx context.Context, url string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM sales WHERE url = $1)`

	err := r.QueryRow(ctx, &exists, query, url)
	if err != nil {
		return false, fmt.Errorf("failed to check sale existence: %w", err)
	}

	return exists, nil
}

// DeleteByAnalysisID deletes all sales for an analysis.
func (r *SaleRepository) DeleteByAnalysisID(ctx context.Context, analysisID uuid.UUID) (int, error) {
	query := `DELETE FROM sales WHERE analysis_id = $1`

	affected, err := r.Exec(ctx, query, analysisID)
	if err != nil {
		return 0, fmt.Errorf("failed to delete sales: %w", err)
	}

	return int(affected), nil
}

// GetRecentSales returns the most recent sales across all products.
func (r *SaleRepository) GetRecentSales(ctx context.Context, limit int) ([]models.Sale, error) {
	var sales []models.Sale
	query := `
		SELECT id, product_id, analysis_id, platform, title, price, currency, sold_at, url, scraped_at
		FROM sales
		ORDER BY sold_at DESC
		LIMIT $1
	`

	err := r.QueryRows(ctx, &sales, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent sales: %w", err)
	}

	return sales, nil
}
