package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ProductStats represents aggregated statistics for a product.
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

// HasMSRP returns true if the product has MSRP data.
func (ps *ProductStats) HasMSRP() bool {
	return ps.MSRPEUR != nil && ps.MSRPEUR.IsPositive()
}

// HasMargin returns true if margin data is available.
func (ps *ProductStats) HasMargin() bool {
	return ps.MarginPercent != nil
}

// IsProfitable returns true if the average price is above MSRP.
func (ps *ProductStats) IsProfitable() bool {
	if !ps.HasMargin() {
		return false
	}
	return ps.MarginPercent.IsPositive()
}

// PriceRange returns the min-max price range.
func (ps *ProductStats) PriceRange() decimal.Decimal {
	return ps.MaxPrice.Sub(ps.MinPrice)
}

// FormatMarginPercent returns a formatted margin percentage string.
func (ps *ProductStats) FormatMarginPercent() string {
	if !ps.HasMargin() {
		return "N/A"
	}
	sign := ""
	if ps.MarginPercent.IsPositive() {
		sign = "+"
	}
	return sign + ps.MarginPercent.Round(1).String() + "%"
}

// FormatAvgPrice returns a formatted average price string in EUR.
func (ps *ProductStats) FormatAvgPrice() string {
	return ps.AvgPrice.Round(2).String() + "€"
}

// FormatMSRP returns a formatted MSRP string in EUR.
func (ps *ProductStats) FormatMSRP() string {
	if !ps.HasMSRP() {
		return "N/A"
	}
	return ps.MSRPEUR.Round(2).String() + "€"
}
