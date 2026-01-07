# Scraper Interface Contract: Volume Analyzer

**Date**: 2026-01-07
**Feature Branch**: `001-ebay-volume-analyzer`

## Overview

This document defines the internal Go interfaces for the eBay scraper component.

---

## Core Interfaces

### Scraper Interface

```go
package scraper

import (
    "context"
    "time"
)

// Scraper defines the contract for platform-specific scrapers
type Scraper interface {
    // Name returns the platform name (e.g., "ebay")
    Name() string

    // Scrape performs a scraping session and returns raw sale data
    Scrape(ctx context.Context, opts ScrapeOptions) (*ScrapeResult, error)

    // Health checks if the scraper can reach the target platform
    Health(ctx context.Context) error
}

// ScrapeOptions configures a scraping session
type ScrapeOptions struct {
    // Query is the search term (e.g., "Pokemon Display 151")
    Query string

    // Category filters by product type (optional)
    Category string

    // MaxPages limits the number of search result pages to scrape
    MaxPages int

    // Since filters sales to those completed after this time
    Since time.Time
}

// ScrapeResult contains the output of a scraping session
type ScrapeResult struct {
    // Sales contains the raw scraped sale records
    Sales []RawSale

    // PagesScraped is the number of pages processed
    PagesScraped int

    // Duration is the total scraping time
    Duration time.Duration

    // Errors contains non-fatal errors encountered during scraping
    Errors []error
}

// RawSale represents a single scraped sale before normalization
type RawSale struct {
    // Platform source (e.g., "ebay")
    Platform string

    // Title is the original listing title
    Title string

    // Price is the sale price in the original currency
    Price float64

    // Currency code (e.g., "EUR")
    Currency string

    // SoldAt is when the sale was completed
    SoldAt time.Time

    // URL is the unique listing URL
    URL string

    // Metadata contains platform-specific data
    Metadata map[string]string
}
```

---

### Normalizer Interface

```go
package normalizer

// Normalizer converts raw sale titles into normalized products
type Normalizer interface {
    // Normalize extracts product information from a raw sale title
    // Returns the matched product and a confidence score (0.0-1.0)
    Normalize(title string) (NormalizedProduct, float64)

    // AddPattern adds a new recognition pattern
    AddPattern(pattern ProductPattern) error
}

// NormalizedProduct represents a recognized product
type NormalizedProduct struct {
    // NormalizedName is the canonical product name
    NormalizedName string

    // Category is the product type
    Category string

    // SetName is the Pokemon set name (if identified)
    SetName string

    // SetCode is the short set identifier (if known)
    SetCode string
}

// ProductPattern defines a recognition pattern
type ProductPattern struct {
    // SetKeywords are terms that identify the set
    SetKeywords []string

    // TypeKeywords are terms that identify the product type
    TypeKeywords []string

    // ExcludeKeywords are terms that disqualify a match
    ExcludeKeywords []string

    // MSRP is the retail price for ROI calculation
    MSRP float64
}
```

---

### Repository Interfaces

```go
package repository

import (
    "context"

    "github.com/google/uuid"
)

// ProductRepository handles product data persistence
type ProductRepository interface {
    // FindByNormalizedName finds a product by its canonical name
    FindByNormalizedName(ctx context.Context, name string) (*Product, error)

    // FindOrCreate gets existing product or creates a new one
    FindOrCreate(ctx context.Context, product *Product) (*Product, error)

    // List returns products with optional filtering
    List(ctx context.Context, opts ProductListOptions) ([]Product, error)
}

// SaleRepository handles sale data persistence
type SaleRepository interface {
    // Create inserts a new sale (ignores if URL exists)
    Create(ctx context.Context, sale *Sale) error

    // BulkCreate inserts multiple sales efficiently
    BulkCreate(ctx context.Context, sales []Sale) (int, error)

    // FindByProductID returns sales for a product
    FindByProductID(ctx context.Context, productID uuid.UUID, opts SaleListOptions) ([]Sale, error)
}

// AnalysisRepository handles analysis session data
type AnalysisRepository interface {
    // Create starts a new analysis session
    Create(ctx context.Context) (*Analysis, error)

    // Update modifies an existing analysis
    Update(ctx context.Context, analysis *Analysis) error

    // GetLatest returns the most recent completed analysis
    GetLatest(ctx context.Context) (*Analysis, error)

    // GetByID retrieves an analysis by ID
    GetByID(ctx context.Context, id uuid.UUID) (*Analysis, error)
}

// StatsRepository handles aggregated statistics
type StatsRepository interface {
    // GetProductStats returns volume and profit stats for all products
    GetProductStats(ctx context.Context, opts StatsOptions) ([]ProductStats, error)

    // RefreshStats updates the materialized view
    RefreshStats(ctx context.Context) error
}

// Options structs

type ProductListOptions struct {
    Category string
    Limit    int
    Offset   int
}

type SaleListOptions struct {
    Since  time.Time
    Limit  int
    Offset int
}

type StatsOptions struct {
    Category     string
    SortBy       string // "sales_count", "margin_percent", "avg_price"
    SortOrder    string // "asc", "desc"
    MinSales     int
    Limit        int
    Offset       int
}
```

---

### Analyzer Interface

```go
package analyzer

import (
    "context"
)

// VolumeAnalyzer orchestrates the analysis workflow
type VolumeAnalyzer interface {
    // Run executes a complete analysis session
    Run(ctx context.Context, opts AnalyzeOptions) (*AnalysisResult, error)

    // GetStatus returns the current analysis status (if running)
    GetStatus(ctx context.Context) (*AnalysisStatus, error)
}

// AnalyzeOptions configures an analysis run
type AnalyzeOptions struct {
    // Query is the search term
    Query string

    // Category filters products
    Category string

    // OnProgress is called with progress updates
    OnProgress func(progress AnalysisProgress)
}

// AnalysisProgress reports analysis progress
type AnalysisProgress struct {
    Phase          string // "scraping", "normalizing", "saving", "calculating"
    PagesScraped   int
    SalesFound     int
    ProductsMatched int
    PercentComplete float64
}

// AnalysisResult contains the final analysis output
type AnalysisResult struct {
    AnalysisID    uuid.UUID
    ProductsCount int
    SalesCount    int
    Duration      time.Duration
    TopProducts   []ProductStats
}

// AnalysisStatus reports running analysis state
type AnalysisStatus struct {
    IsRunning   bool
    AnalysisID  uuid.UUID
    StartedAt   time.Time
    Progress    AnalysisProgress
}
```

---

## Error Types

```go
package errors

import "errors"

var (
    // ErrNoData indicates no analysis data is available
    ErrNoData = errors.New("no analysis data available")

    // ErrAnalysisRunning indicates an analysis is already in progress
    ErrAnalysisRunning = errors.New("analysis already running")

    // ErrScrapeFailed indicates the scraper could not reach the target
    ErrScrapeFailed = errors.New("scraping failed")

    // ErrRateLimited indicates too many requests
    ErrRateLimited = errors.New("rate limited by target platform")

    // ErrProductNotFound indicates the product doesn't exist
    ErrProductNotFound = errors.New("product not found")

    // ErrInvalidCategory indicates an unknown product category
    ErrInvalidCategory = errors.New("invalid product category")
)
```

---

## Data Flow

```
┌─────────────────┐
│  Discord /analyze
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ VolumeAnalyzer  │
│   .Run()        │
└────────┬────────┘
         │
         ▼
┌─────────────────┐     ┌─────────────────┐
│  EbayScraper    │────▶│   RawSale[]     │
│   .Scrape()     │     │  (unprocessed)  │
└─────────────────┘     └────────┬────────┘
                                 │
                                 ▼
                        ┌─────────────────┐
                        │  Normalizer     │
                        │  .Normalize()   │
                        └────────┬────────┘
                                 │
                                 ▼
                        ┌─────────────────┐
                        │  Repository     │
                        │  .BulkCreate()  │
                        └────────┬────────┘
                                 │
                                 ▼
                        ┌─────────────────┐
                        │ StatsRepository │
                        │ .RefreshStats() │
                        └────────┬────────┘
                                 │
                                 ▼
                        ┌─────────────────┐
                        │  Discord Embed  │
                        │  (results)      │
                        └─────────────────┘
```
