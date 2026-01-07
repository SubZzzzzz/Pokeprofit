package analyzer

import (
	"github.com/SubZzzzzz/pokeprofit/internal/models"
	"github.com/shopspring/decimal"
)

// ProfitCalculator calculates profitability metrics for products.
type ProfitCalculator struct {
	// defaultFees represents the default fees (eBay fees, shipping, etc.)
	// as a percentage (e.g., 0.13 for 13%)
	defaultFees decimal.Decimal
}

// NewProfitCalculator creates a new ProfitCalculator with default fees.
func NewProfitCalculator() *ProfitCalculator {
	return &ProfitCalculator{
		// Default eBay fees ~13% (10% final value + 3% payment processing)
		defaultFees: decimal.NewFromFloat(0.13),
	}
}

// NewProfitCalculatorWithFees creates a ProfitCalculator with custom fees.
func NewProfitCalculatorWithFees(feesPercent float64) *ProfitCalculator {
	return &ProfitCalculator{
		defaultFees: decimal.NewFromFloat(feesPercent),
	}
}

// ProfitResult contains the result of a profit calculation.
type ProfitResult struct {
	// GrossMarginEUR is the margin before fees (avg_price - msrp)
	GrossMarginEUR decimal.Decimal

	// GrossMarginPercent is the gross margin as a percentage
	GrossMarginPercent decimal.Decimal

	// NetMarginEUR is the margin after fees
	NetMarginEUR decimal.Decimal

	// NetMarginPercent is the net margin as a percentage
	NetMarginPercent decimal.Decimal

	// FeesEUR is the estimated fees in EUR
	FeesEUR decimal.Decimal

	// IsProfitable returns true if net margin is positive
	IsProfitable bool

	// ROI is the return on investment (net_margin / msrp * 100)
	ROI decimal.Decimal
}

// Calculate computes profitability metrics for a given average price and MSRP.
func (pc *ProfitCalculator) Calculate(avgPrice, msrp decimal.Decimal) *ProfitResult {
	result := &ProfitResult{}

	// Gross margin (before fees)
	result.GrossMarginEUR = avgPrice.Sub(msrp)

	// Calculate gross margin percentage
	if msrp.IsPositive() {
		result.GrossMarginPercent = result.GrossMarginEUR.Div(msrp).Mul(decimal.NewFromInt(100))
	}

	// Calculate fees (on the sale price)
	result.FeesEUR = avgPrice.Mul(pc.defaultFees)

	// Net margin (after fees)
	result.NetMarginEUR = result.GrossMarginEUR.Sub(result.FeesEUR)

	// Calculate net margin percentage
	if msrp.IsPositive() {
		result.NetMarginPercent = result.NetMarginEUR.Div(msrp).Mul(decimal.NewFromInt(100))
	}

	// Calculate ROI
	if msrp.IsPositive() {
		result.ROI = result.NetMarginEUR.Div(msrp).Mul(decimal.NewFromInt(100))
	}

	// Determine profitability
	result.IsProfitable = result.NetMarginEUR.IsPositive()

	return result
}

// CalculateForStats computes profitability metrics for ProductStats.
func (pc *ProfitCalculator) CalculateForStats(stats *models.ProductStats) *ProfitResult {
	if stats.MSRPEUR == nil || !stats.MSRPEUR.IsPositive() {
		// Cannot calculate without MSRP
		return &ProfitResult{
			GrossMarginEUR:     decimal.Zero,
			GrossMarginPercent: decimal.Zero,
			NetMarginEUR:       decimal.Zero,
			NetMarginPercent:   decimal.Zero,
			FeesEUR:            decimal.Zero,
			IsProfitable:       false,
			ROI:                decimal.Zero,
		}
	}

	return pc.Calculate(stats.AvgPrice, *stats.MSRPEUR)
}

// CalculateBatch computes profitability for multiple ProductStats.
func (pc *ProfitCalculator) CalculateBatch(stats []models.ProductStats) map[string]*ProfitResult {
	results := make(map[string]*ProfitResult, len(stats))
	for i := range stats {
		results[stats[i].NormalizedName] = pc.CalculateForStats(&stats[i])
	}
	return results
}

// EstimateProfit calculates estimated profit for a given purchase price and expected sale price.
func (pc *ProfitCalculator) EstimateProfit(purchasePrice, expectedSalePrice decimal.Decimal) *ProfitResult {
	return pc.Calculate(expectedSalePrice, purchasePrice)
}

// GetBreakEvenPrice calculates the minimum sale price needed to break even.
func (pc *ProfitCalculator) GetBreakEvenPrice(purchasePrice decimal.Decimal) decimal.Decimal {
	// breakeven = purchasePrice / (1 - fees)
	one := decimal.NewFromInt(1)
	return purchasePrice.Div(one.Sub(pc.defaultFees))
}

// GetMinimumProfitablePrice calculates the minimum sale price for a target profit margin.
func (pc *ProfitCalculator) GetMinimumProfitablePrice(purchasePrice decimal.Decimal, targetMarginPercent float64) decimal.Decimal {
	// targetSalePrice = purchasePrice * (1 + targetMargin) / (1 - fees)
	targetMargin := decimal.NewFromFloat(targetMarginPercent / 100)
	one := decimal.NewFromInt(1)

	return purchasePrice.Mul(one.Add(targetMargin)).Div(one.Sub(pc.defaultFees))
}

// FormatMarginEUR formats a margin in EUR with sign.
func FormatMarginEUR(margin decimal.Decimal) string {
	sign := ""
	if margin.IsPositive() {
		sign = "+"
	}
	return sign + margin.Round(2).String() + "€"
}

// FormatMarginPercent formats a margin percentage with sign.
func FormatMarginPercent(margin decimal.Decimal) string {
	sign := ""
	if margin.IsPositive() {
		sign = "+"
	}
	return sign + margin.Round(1).String() + "%"
}

// ProfitabilityLevel represents the profitability tier.
type ProfitabilityLevel string

const (
	ProfitabilityExcellent ProfitabilityLevel = "excellent" // > 30% ROI
	ProfitabilityGood      ProfitabilityLevel = "good"      // 15-30% ROI
	ProfitabilityMarginal  ProfitabilityLevel = "marginal"  // 0-15% ROI
	ProfitabilityLoss      ProfitabilityLevel = "loss"      // < 0% ROI
	ProfitabilityUnknown   ProfitabilityLevel = "unknown"   // No MSRP data
)

// GetProfitabilityLevel returns the profitability tier based on ROI.
func GetProfitabilityLevel(roi decimal.Decimal) ProfitabilityLevel {
	thirty := decimal.NewFromInt(30)
	fifteen := decimal.NewFromInt(15)
	zero := decimal.Zero

	switch {
	case roi.GreaterThan(thirty):
		return ProfitabilityExcellent
	case roi.GreaterThan(fifteen):
		return ProfitabilityGood
	case roi.GreaterThanOrEqual(zero):
		return ProfitabilityMarginal
	default:
		return ProfitabilityLoss
	}
}

// GetProfitabilityLevelForStats returns the profitability tier for ProductStats.
func GetProfitabilityLevelForStats(stats *models.ProductStats) ProfitabilityLevel {
	if stats.MarginPercent == nil {
		return ProfitabilityUnknown
	}
	return GetProfitabilityLevel(*stats.MarginPercent)
}

// ColorForProfitability returns a Discord embed color for the profitability level.
func ColorForProfitability(level ProfitabilityLevel) int {
	switch level {
	case ProfitabilityExcellent:
		return 0x00FF00 // Bright green
	case ProfitabilityGood:
		return 0x90EE90 // Light green
	case ProfitabilityMarginal:
		return 0xFFA500 // Orange
	case ProfitabilityLoss:
		return 0xFF0000 // Red
	default:
		return 0x808080 // Gray
	}
}
