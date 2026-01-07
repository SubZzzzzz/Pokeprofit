package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ProductCategory represents the type of Pokemon TCG product.
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

// ValidCategories returns all valid product categories.
func ValidCategories() []ProductCategory {
	return []ProductCategory{
		CategoryBooster,
		CategoryDisplay,
		CategoryETB,
		CategoryCollection,
		CategoryBundle,
		CategoryTin,
		CategorySingle,
	}
}

// IsValid checks if the category is valid.
func (c ProductCategory) IsValid() bool {
	for _, valid := range ValidCategories() {
		if c == valid {
			return true
		}
	}
	return false
}

// String returns the string representation of the category.
func (c ProductCategory) String() string {
	return string(c)
}

// DisplayName returns a human-readable name for the category.
func (c ProductCategory) DisplayName() string {
	switch c {
	case CategoryBooster:
		return "Booster"
	case CategoryDisplay:
		return "Display"
	case CategoryETB:
		return "ETB"
	case CategoryCollection:
		return "Coffret"
	case CategoryBundle:
		return "Bundle"
	case CategoryTin:
		return "Tin"
	case CategorySingle:
		return "Single"
	default:
		return string(c)
	}
}

// Product represents a normalized Pokemon TCG product.
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

// NewProduct creates a new Product with default values.
func NewProduct(normalizedName string, category ProductCategory) *Product {
	return &Product{
		ID:             uuid.New(),
		NormalizedName: normalizedName,
		Category:       category,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// SetMSRP sets the MSRP value.
func (p *Product) SetMSRP(msrp decimal.Decimal) {
	p.MSRPEUR = &msrp
}

// SetSetInfo sets the set name and code.
func (p *Product) SetSetInfo(name, code string) {
	p.SetName = &name
	p.SetCode = &code
}
