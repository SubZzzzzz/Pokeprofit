package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/SubZzzzzz/pokeprofit/internal/database"
	apperrors "github.com/SubZzzzzz/pokeprofit/internal/errors"
	"github.com/SubZzzzzz/pokeprofit/internal/models"
	"github.com/google/uuid"
)

// ProductRepository handles product data persistence.
type ProductRepository struct {
	*Base
}

// NewProductRepository creates a new ProductRepository.
func NewProductRepository(db *database.DB) *ProductRepository {
	return &ProductRepository{
		Base: NewBase(db),
	}
}

// ProductListOptions configures product list queries.
type ProductListOptions struct {
	Category string
	Limit    int
	Offset   int
}

// FindByID retrieves a product by ID.
func (r *ProductRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	var product models.Product
	query := `
		SELECT id, normalized_name, category, set_name, set_code, msrp_eur, created_at, updated_at
		FROM products
		WHERE id = $1
	`

	err := r.QueryRow(ctx, &product, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.ErrProductNotFound
		}
		return nil, fmt.Errorf("failed to find product by ID: %w", err)
	}

	return &product, nil
}

// FindByNormalizedName finds a product by its canonical name.
func (r *ProductRepository) FindByNormalizedName(ctx context.Context, name string) (*models.Product, error) {
	var product models.Product
	query := `
		SELECT id, normalized_name, category, set_name, set_code, msrp_eur, created_at, updated_at
		FROM products
		WHERE normalized_name = $1
	`

	err := r.QueryRow(ctx, &product, query, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.ErrProductNotFound
		}
		return nil, fmt.Errorf("failed to find product by normalized name: %w", err)
	}

	return &product, nil
}

// FindOrCreate gets existing product or creates a new one.
func (r *ProductRepository) FindOrCreate(ctx context.Context, product *models.Product) (*models.Product, error) {
	// Try to find existing product
	existing, err := r.FindByNormalizedName(ctx, product.NormalizedName)
	if err == nil {
		return existing, nil
	}
	if !apperrors.Is(err, apperrors.ErrProductNotFound) {
		return nil, err
	}

	// Create new product
	if err := r.Create(ctx, product); err != nil {
		// Handle race condition - try to find again
		existing, findErr := r.FindByNormalizedName(ctx, product.NormalizedName)
		if findErr == nil {
			return existing, nil
		}
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return product, nil
}

// Create inserts a new product.
func (r *ProductRepository) Create(ctx context.Context, product *models.Product) error {
	query := `
		INSERT INTO products (id, normalized_name, category, set_name, set_code, msrp_eur, created_at, updated_at)
		VALUES (:id, :normalized_name, :category, :set_name, :set_code, :msrp_eur, :created_at, :updated_at)
	`

	return r.InsertNamed(ctx, query, product)
}

// Update updates an existing product.
func (r *ProductRepository) Update(ctx context.Context, product *models.Product) error {
	query := `
		UPDATE products
		SET normalized_name = :normalized_name,
		    category = :category,
		    set_name = :set_name,
		    set_code = :set_code,
		    msrp_eur = :msrp_eur,
		    updated_at = NOW()
		WHERE id = :id
	`

	affected, err := r.ExecNamed(ctx, query, product)
	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}
	if affected == 0 {
		return apperrors.ErrProductNotFound
	}

	return nil
}

// List returns products with optional filtering.
func (r *ProductRepository) List(ctx context.Context, opts ProductListOptions) ([]models.Product, error) {
	var products []models.Product

	query := `
		SELECT id, normalized_name, category, set_name, set_code, msrp_eur, created_at, updated_at
		FROM products
		WHERE 1=1
	`
	args := []interface{}{}
	argIndex := 1

	if opts.Category != "" {
		query += fmt.Sprintf(" AND category = $%d", argIndex)
		args = append(args, opts.Category)
		argIndex++
	}

	query += " ORDER BY normalized_name"

	if opts.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, opts.Limit)
		argIndex++
	}

	if opts.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, opts.Offset)
	}

	err := r.QueryRows(ctx, &products, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	return products, nil
}

// Count returns the total number of products.
func (r *ProductRepository) Count(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM products`

	err := r.QueryRow(ctx, &count, query)
	if err != nil {
		return 0, fmt.Errorf("failed to count products: %w", err)
	}

	return count, nil
}

// Delete removes a product by ID.
func (r *ProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM products WHERE id = $1`

	affected, err := r.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}
	if affected == 0 {
		return apperrors.ErrProductNotFound
	}

	return nil
}

// FindBySetCode retrieves products by set code.
func (r *ProductRepository) FindBySetCode(ctx context.Context, setCode string) ([]models.Product, error) {
	var products []models.Product
	query := `
		SELECT id, normalized_name, category, set_name, set_code, msrp_eur, created_at, updated_at
		FROM products
		WHERE set_code = $1
		ORDER BY category, normalized_name
	`

	err := r.QueryRows(ctx, &products, query, setCode)
	if err != nil {
		return nil, fmt.Errorf("failed to find products by set code: %w", err)
	}

	return products, nil
}
