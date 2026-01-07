package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Sale represents a completed sale scraped from eBay.
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

// NewSale creates a new Sale with default values.
func NewSale(productID uuid.UUID, title string, price decimal.Decimal, soldAt time.Time) *Sale {
	return &Sale{
		ID:        uuid.New(),
		ProductID: productID,
		Platform:  "ebay",
		Title:     title,
		Price:     price,
		Currency:  "EUR",
		SoldAt:    soldAt,
		ScrapedAt: time.Now(),
	}
}

// SetURL sets the sale URL.
func (s *Sale) SetURL(url string) {
	s.URL = &url
}

// SetAnalysisID sets the analysis ID.
func (s *Sale) SetAnalysisID(analysisID uuid.UUID) {
	s.AnalysisID = &analysisID
}
