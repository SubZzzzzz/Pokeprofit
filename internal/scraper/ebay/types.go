package ebay

import (
	"time"
)

// RawSale represents a single scraped sale before normalization.
type RawSale struct {
	Platform string            `json:"platform"`
	Title    string            `json:"title"`
	Price    float64           `json:"price"`
	Currency string            `json:"currency"`
	SoldAt   time.Time         `json:"sold_at"`
	URL      string            `json:"url"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// ScrapeOptions configures a scraping session.
type ScrapeOptions struct {
	Query    string
	Category string
	MaxPages int
	Since    time.Time
}

// DefaultScrapeOptions returns default scraping options.
func DefaultScrapeOptions() ScrapeOptions {
	return ScrapeOptions{
		MaxPages: 10,
		Since:    time.Now().AddDate(0, 0, -30), // 30 days ago
	}
}

// ScrapeResult contains the output of a scraping session.
type ScrapeResult struct {
	Sales        []RawSale
	PagesScraped int
	Duration     time.Duration
	Errors       []error
}

// NewScrapeResult creates a new empty ScrapeResult.
func NewScrapeResult() *ScrapeResult {
	return &ScrapeResult{
		Sales:  make([]RawSale, 0),
		Errors: make([]error, 0),
	}
}

// AddSale adds a sale to the result.
func (r *ScrapeResult) AddSale(sale RawSale) {
	r.Sales = append(r.Sales, sale)
}

// AddError adds an error to the result.
func (r *ScrapeResult) AddError(err error) {
	r.Errors = append(r.Errors, err)
}

// SaleCount returns the number of sales scraped.
func (r *ScrapeResult) SaleCount() int {
	return len(r.Sales)
}

// HasErrors returns true if errors occurred during scraping.
func (r *ScrapeResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// Scraper defines the contract for platform-specific scrapers.
type Scraper interface {
	Name() string
	Scrape(opts ScrapeOptions) (*ScrapeResult, error)
	Health() error
}

// eBay-specific constants
const (
	EbayFRBaseURL        = "https://www.ebay.fr"
	EbaySearchPath       = "/sch/i.html"
	EbayPokemonTCGCatID  = "183454"
	DefaultRateLimitMs   = 2000
	DefaultMaxPages      = 10
	DefaultUserAgent     = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
)

// eBay URL query parameters
const (
	ParamKeyword      = "_nkw"       // Search keyword
	ParamCategory     = "_sacat"     // Category ID
	ParamCompleted    = "LH_Complete" // Show completed listings
	ParamSold         = "LH_Sold"    // Show sold items only
	ParamSort         = "_sop"       // Sort order
	ParamItemsPerPage = "_ipg"       // Items per page
	ParamPageNumber   = "_pgn"       // Page number
)

// eBay sort options
const (
	SortEndDateRecent = "13" // Sort by end date (most recent first)
	SortPriceLowest   = "15" // Sort by price (lowest first)
	SortPriceHighest  = "16" // Sort by price (highest first)
)
