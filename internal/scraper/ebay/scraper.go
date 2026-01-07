package ebay

import (
	"context"
	"fmt"
	"math/rand"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/SubZzzzzz/pokeprofit/internal/logger"
	"github.com/SubZzzzzz/pokeprofit/internal/scraper/common"
	"github.com/gocolly/colly/v2"
)

// EbayScraper implements the Scraper interface for eBay FR.
type EbayScraper struct {
	collector   *colly.Collector
	parser      *Parser
	rateLimiter *common.RateLimiter
	retryer     *common.Retryer
	userAgents  []string
	log         *logger.Logger
}

// ScraperConfig holds configuration for the eBay scraper.
type ScraperConfig struct {
	RateLimitDelay time.Duration
	MaxRetries     int
	UserAgents     []string
}

// DefaultScraperConfig returns default scraper configuration.
func DefaultScraperConfig() ScraperConfig {
	return ScraperConfig{
		RateLimitDelay: 2 * time.Second,
		MaxRetries:     3,
		UserAgents: []string{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		},
	}
}

// NewEbayScraper creates a new eBay scraper with the given configuration.
func NewEbayScraper(cfg ScraperConfig) *EbayScraper {
	c := colly.NewCollector(
		colly.AllowedDomains("www.ebay.fr", "ebay.fr"),
	)

	// Configure rate limiting via colly
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*ebay.*",
		Parallelism: 1,
		Delay:       cfg.RateLimitDelay,
		RandomDelay: 1 * time.Second,
	})

	// Set default user agent
	if len(cfg.UserAgents) > 0 {
		c.UserAgent = cfg.UserAgents[0]
	} else {
		c.UserAgent = DefaultUserAgent
	}

	retryConfig := common.RetryConfig{
		MaxRetries:    cfg.MaxRetries,
		InitialDelay:  2 * time.Second,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
		Jitter:        true,
	}

	return &EbayScraper{
		collector:   c,
		parser:      NewParser(),
		rateLimiter: common.NewRateLimiter(cfg.RateLimitDelay),
		retryer:     common.NewRetryer(retryConfig),
		userAgents:  cfg.UserAgents,
		log:         logger.Default().WithComponent("scraper.ebay"),
	}
}

// NewDefaultEbayScraper creates a new eBay scraper with default configuration.
func NewDefaultEbayScraper() *EbayScraper {
	return NewEbayScraper(DefaultScraperConfig())
}

// Name returns the scraper platform name.
func (s *EbayScraper) Name() string {
	return "ebay"
}

// HealthStatus contains the result of a health check.
type HealthStatus struct {
	Healthy      bool          `json:"healthy"`
	Platform     string        `json:"platform"`
	ResponseTime time.Duration `json:"response_time"`
	StatusCode   int           `json:"status_code,omitempty"`
	Error        string        `json:"error,omitempty"`
	CheckedAt    time.Time     `json:"checked_at"`
}

// Health checks if the scraper can reach eBay.
func (s *EbayScraper) Health() error {
	status := s.HealthCheck(context.Background())
	if !status.Healthy {
		return fmt.Errorf("health check failed: %s", status.Error)
	}
	return nil
}

// HealthCheck performs a detailed health check and returns status information.
func (s *EbayScraper) HealthCheck(ctx context.Context) *HealthStatus {
	start := time.Now()
	status := &HealthStatus{
		Platform:  "ebay.fr",
		CheckedAt: start,
	}

	c := s.collector.Clone()

	c.OnError(func(r *colly.Response, err error) {
		status.Error = err.Error()
	})

	c.OnResponse(func(r *colly.Response) {
		status.StatusCode = r.StatusCode
		if r.StatusCode != 200 {
			status.Error = fmt.Sprintf("unexpected status code: %d", r.StatusCode)
		}
	})

	err := c.Visit(EbayFRBaseURL)
	status.ResponseTime = time.Since(start)

	if err != nil {
		status.Error = err.Error()
		status.Healthy = false
		s.log.Warn("Health check failed", "platform", status.Platform, "error", status.Error)
		return status
	}

	if status.Error != "" {
		status.Healthy = false
		s.log.Warn("Health check failed", "platform", status.Platform, "error", status.Error)
		return status
	}

	status.Healthy = true
	s.log.Debug("Health check passed", "platform", status.Platform, "response_time", status.ResponseTime)
	return status
}

// Scrape performs a scraping session and returns raw sale data.
func (s *EbayScraper) Scrape(opts ScrapeOptions) (*ScrapeResult, error) {
	return s.ScrapeWithContext(context.Background(), opts)
}

// ScrapeWithContext performs a scraping session with context support.
func (s *EbayScraper) ScrapeWithContext(ctx context.Context, opts ScrapeOptions) (*ScrapeResult, error) {
	if opts.Query == "" {
		return nil, fmt.Errorf("query is required")
	}

	if opts.MaxPages <= 0 {
		opts.MaxPages = DefaultMaxPages
	}

	if opts.Since.IsZero() {
		opts.Since = time.Now().AddDate(0, 0, -30)
	}

	result := NewScrapeResult()
	startTime := time.Now()

	// Create a new collector for this scrape session
	c := s.collector.Clone()

	// Rotate user agent
	c.UserAgent = s.getRandomUserAgent()

	var scrapeErr error

	// Handle errors
	c.OnError(func(r *colly.Response, err error) {
		s.log.Error("Scrape error", "url", r.Request.URL.String(), "error", err)
		result.AddError(fmt.Errorf("request failed: %w", err))
	})

	// Handle responses
	c.OnResponse(func(r *colly.Response) {
		if r.StatusCode != 200 {
			result.AddError(fmt.Errorf("unexpected status code: %d", r.StatusCode))
		}
	})

	// Parse HTML content
	c.OnHTML("body", func(e *colly.HTMLElement) {
		doc := e.DOM.Parent()
		goqueryDoc := goquery.NewDocumentFromNode(doc.Get(0))

		sales, err := s.parser.ParseSearchResults(goqueryDoc)
		if err != nil {
			result.AddError(fmt.Errorf("parse error: %w", err))
			return
		}

		// Filter sales by date
		for _, sale := range sales {
			if !sale.SoldAt.Before(opts.Since) {
				result.AddSale(sale)
			}
		}
	})

	// Scrape pages
	for page := 1; page <= opts.MaxPages; page++ {
		select {
		case <-ctx.Done():
			scrapeErr = ctx.Err()
			break
		default:
		}

		if scrapeErr != nil {
			break
		}

		// Wait for rate limiter
		if err := s.rateLimiter.Wait(ctx); err != nil {
			scrapeErr = err
			break
		}

		pageURL := s.buildSearchURL(opts.Query, opts.Category, page)

		// Use retry logic for each page
		err := s.retryer.Do(ctx, func() error {
			return c.Visit(pageURL)
		})

		if err != nil {
			result.AddError(fmt.Errorf("failed to scrape page %d: %w", page, err))
			// Continue to next page on error
			continue
		}

		result.PagesScraped = page

		// Check if we got any results on this page
		// If the page is empty, stop scraping
		if result.SaleCount() == 0 && page > 1 {
			break
		}
	}

	result.Duration = time.Since(startTime)

	if scrapeErr != nil {
		return result, scrapeErr
	}

	return result, nil
}

// buildSearchURL constructs the eBay search URL for completed listings.
func (s *EbayScraper) buildSearchURL(query, category string, page int) string {
	baseURL := EbayFRBaseURL + EbaySearchPath

	params := url.Values{}
	params.Set(ParamKeyword, query)
	params.Set(ParamCompleted, "1")
	params.Set(ParamSold, "1")
	params.Set(ParamSort, SortEndDateRecent)
	params.Set(ParamItemsPerPage, "100")

	if category != "" {
		catID := s.getCategoryID(category)
		if catID != "" {
			params.Set(ParamCategory, catID)
		}
	} else {
		// Default to Pokemon TCG category
		params.Set(ParamCategory, EbayPokemonTCGCatID)
	}

	if page > 1 {
		params.Set(ParamPageNumber, fmt.Sprintf("%d", page))
	}

	return baseURL + "?" + params.Encode()
}

// getCategoryID maps internal category names to eBay category IDs.
func (s *EbayScraper) getCategoryID(category string) string {
	categoryMap := map[string]string{
		"all":        EbayPokemonTCGCatID,
		"display":    EbayPokemonTCGCatID,
		"etb":        EbayPokemonTCGCatID,
		"collection": EbayPokemonTCGCatID,
		"booster":    EbayPokemonTCGCatID,
		"bundle":     EbayPokemonTCGCatID,
		"tin":        EbayPokemonTCGCatID,
		"single":     "183454", // Pokemon TCG Singles
	}

	if id, ok := categoryMap[strings.ToLower(category)]; ok {
		return id
	}
	return EbayPokemonTCGCatID
}

// getRandomUserAgent returns a random user agent from the configured list.
func (s *EbayScraper) getRandomUserAgent() string {
	if len(s.userAgents) == 0 {
		return DefaultUserAgent
	}
	return s.userAgents[rand.Intn(len(s.userAgents))]
}

// SetUserAgents updates the user agent list.
func (s *EbayScraper) SetUserAgents(agents []string) {
	s.userAgents = agents
}

// GetRateLimiter returns the rate limiter for external use.
func (s *EbayScraper) GetRateLimiter() *common.RateLimiter {
	return s.rateLimiter
}
