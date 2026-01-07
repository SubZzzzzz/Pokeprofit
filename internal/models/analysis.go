package models

import (
	"time"

	"github.com/google/uuid"
)

// AnalysisStatus represents the state of an analysis session.
type AnalysisStatus string

const (
	StatusRunning   AnalysisStatus = "running"
	StatusCompleted AnalysisStatus = "completed"
	StatusFailed    AnalysisStatus = "failed"
)

// String returns the string representation of the status.
func (s AnalysisStatus) String() string {
	return string(s)
}

// IsTerminal returns true if the status is a terminal state.
func (s AnalysisStatus) IsTerminal() bool {
	return s == StatusCompleted || s == StatusFailed
}

// Analysis represents an analysis session.
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

// NewAnalysis creates a new Analysis in running state.
func NewAnalysis() *Analysis {
	return &Analysis{
		ID:            uuid.New(),
		StartedAt:     time.Now(),
		Status:        StatusRunning,
		ProductsCount: 0,
		SalesCount:    0,
	}
}

// SetSearchQuery sets the search query.
func (a *Analysis) SetSearchQuery(query string) {
	a.SearchQuery = &query
}

// Complete marks the analysis as completed.
func (a *Analysis) Complete(productsCount, salesCount int) {
	now := time.Now()
	a.CompletedAt = &now
	a.Status = StatusCompleted
	a.ProductsCount = productsCount
	a.SalesCount = salesCount
}

// Fail marks the analysis as failed.
func (a *Analysis) Fail(err error) {
	now := time.Now()
	a.CompletedAt = &now
	a.Status = StatusFailed
	errMsg := err.Error()
	a.ErrorMessage = &errMsg
}

// Duration returns the analysis duration if completed.
func (a *Analysis) Duration() time.Duration {
	if a.CompletedAt == nil {
		return time.Since(a.StartedAt)
	}
	return a.CompletedAt.Sub(a.StartedAt)
}
