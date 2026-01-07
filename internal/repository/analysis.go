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

// AnalysisRepository handles analysis session data.
type AnalysisRepository struct {
	*Base
}

// NewAnalysisRepository creates a new AnalysisRepository.
func NewAnalysisRepository(db *database.DB) *AnalysisRepository {
	return &AnalysisRepository{
		Base: NewBase(db),
	}
}

// Create starts a new analysis session.
func (r *AnalysisRepository) Create(ctx context.Context, analysis *models.Analysis) error {
	query := `
		INSERT INTO analyses (id, started_at, completed_at, status, products_count, sales_count, search_query, error_message)
		VALUES (:id, :started_at, :completed_at, :status, :products_count, :sales_count, :search_query, :error_message)
	`

	return r.InsertNamed(ctx, query, analysis)
}

// Update modifies an existing analysis.
func (r *AnalysisRepository) Update(ctx context.Context, analysis *models.Analysis) error {
	query := `
		UPDATE analyses
		SET completed_at = :completed_at,
		    status = :status,
		    products_count = :products_count,
		    sales_count = :sales_count,
		    search_query = :search_query,
		    error_message = :error_message
		WHERE id = :id
	`

	affected, err := r.ExecNamed(ctx, query, analysis)
	if err != nil {
		return fmt.Errorf("failed to update analysis: %w", err)
	}
	if affected == 0 {
		return apperrors.Wrap(apperrors.ErrNoData, "analysis not found")
	}

	return nil
}

// GetByID retrieves an analysis by ID.
func (r *AnalysisRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Analysis, error) {
	var analysis models.Analysis
	query := `
		SELECT id, started_at, completed_at, status, products_count, sales_count, search_query, error_message
		FROM analyses
		WHERE id = $1
	`

	err := r.QueryRow(ctx, &analysis, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.ErrNoData
		}
		return nil, fmt.Errorf("failed to get analysis by ID: %w", err)
	}

	return &analysis, nil
}

// GetLatest returns the most recent completed analysis.
func (r *AnalysisRepository) GetLatest(ctx context.Context) (*models.Analysis, error) {
	var analysis models.Analysis
	query := `
		SELECT id, started_at, completed_at, status, products_count, sales_count, search_query, error_message
		FROM analyses
		WHERE status = 'completed'
		ORDER BY completed_at DESC
		LIMIT 1
	`

	err := r.QueryRow(ctx, &analysis, query)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.ErrNoData
		}
		return nil, fmt.Errorf("failed to get latest analysis: %w", err)
	}

	return &analysis, nil
}

// GetRunning returns the currently running analysis, if any.
func (r *AnalysisRepository) GetRunning(ctx context.Context) (*models.Analysis, error) {
	var analysis models.Analysis
	query := `
		SELECT id, started_at, completed_at, status, products_count, sales_count, search_query, error_message
		FROM analyses
		WHERE status = 'running'
		ORDER BY started_at DESC
		LIMIT 1
	`

	err := r.QueryRow(ctx, &analysis, query)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No running analysis is not an error
		}
		return nil, fmt.Errorf("failed to get running analysis: %w", err)
	}

	return &analysis, nil
}

// List returns analyses with optional filtering.
func (r *AnalysisRepository) List(ctx context.Context, limit, offset int) ([]models.Analysis, error) {
	var analyses []models.Analysis
	query := `
		SELECT id, started_at, completed_at, status, products_count, sales_count, search_query, error_message
		FROM analyses
		ORDER BY started_at DESC
		LIMIT $1 OFFSET $2
	`

	err := r.QueryRows(ctx, &analyses, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list analyses: %w", err)
	}

	return analyses, nil
}

// Count returns the total number of analyses.
func (r *AnalysisRepository) Count(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM analyses`

	err := r.QueryRow(ctx, &count, query)
	if err != nil {
		return 0, fmt.Errorf("failed to count analyses: %w", err)
	}

	return count, nil
}

// CountByStatus returns the number of analyses with a specific status.
func (r *AnalysisRepository) CountByStatus(ctx context.Context, status models.AnalysisStatus) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM analyses WHERE status = $1`

	err := r.QueryRow(ctx, &count, query, status)
	if err != nil {
		return 0, fmt.Errorf("failed to count analyses by status: %w", err)
	}

	return count, nil
}

// Delete removes an analysis by ID.
func (r *AnalysisRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM analyses WHERE id = $1`

	affected, err := r.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete analysis: %w", err)
	}
	if affected == 0 {
		return apperrors.Wrap(apperrors.ErrNoData, "analysis not found")
	}

	return nil
}

// MarkStaleAsFailes marks any running analyses older than the threshold as failed.
// This is useful for cleaning up stuck analyses on startup.
func (r *AnalysisRepository) MarkStaleAsFailed(ctx context.Context, thresholdMinutes int) (int, error) {
	query := `
		UPDATE analyses
		SET status = 'failed',
		    completed_at = NOW(),
		    error_message = 'Analysis timed out'
		WHERE status = 'running'
		AND started_at < NOW() - INTERVAL '1 minute' * $1
	`

	affected, err := r.Exec(ctx, query, thresholdMinutes)
	if err != nil {
		return 0, fmt.Errorf("failed to mark stale analyses as failed: %w", err)
	}

	return int(affected), nil
}
