package analyzer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/SubZzzzzz/pokeprofit/internal/logger"
	"github.com/SubZzzzzz/pokeprofit/internal/models"
	"github.com/SubZzzzzz/pokeprofit/internal/repository"
	"github.com/SubZzzzzz/pokeprofit/internal/scraper/ebay"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// AnalyzeOptions configures an analysis run.
type AnalyzeOptions struct {
	Query      string
	Category   string
	MaxPages   int
	OnProgress func(progress AnalysisProgress)
}

// AnalysisProgress reports analysis progress.
type AnalysisProgress struct {
	Phase           string  // "scraping", "normalizing", "saving", "complete", "failed"
	PagesScraped    int
	SalesFound      int
	ProductsMatched int
	PercentComplete float64
	Message         string
}

// AnalysisResult contains the final analysis output.
type AnalysisResult struct {
	AnalysisID    uuid.UUID
	ProductsCount int
	SalesCount    int
	Duration      time.Duration
	TopProducts   []models.ProductStats
}

// AnalysisStatus reports running analysis state.
type AnalysisStatus struct {
	IsRunning  bool
	AnalysisID uuid.UUID
	StartedAt  time.Time
	Progress   AnalysisProgress
}

// VolumeAnalyzer orchestrates the analysis workflow.
type VolumeAnalyzer struct {
	scraper    *ebay.EbayScraper
	normalizer *Normalizer
	productRepo *repository.ProductRepository
	saleRepo    *repository.SaleRepository
	analysisRepo *repository.AnalysisRepository
	log         *logger.Logger

	mu            sync.RWMutex
	currentStatus *AnalysisStatus
}

// NewVolumeAnalyzer creates a new VolumeAnalyzer.
func NewVolumeAnalyzer(
	scraper *ebay.EbayScraper,
	productRepo *repository.ProductRepository,
	saleRepo *repository.SaleRepository,
	analysisRepo *repository.AnalysisRepository,
) *VolumeAnalyzer {
	return &VolumeAnalyzer{
		scraper:      scraper,
		normalizer:   NewNormalizer(),
		productRepo:  productRepo,
		saleRepo:     saleRepo,
		analysisRepo: analysisRepo,
		log:          logger.Default().WithComponent("analyzer"),
	}
}

// Run executes a complete analysis session.
func (va *VolumeAnalyzer) Run(ctx context.Context, opts AnalyzeOptions) (*AnalysisResult, error) {
	// Check if already running
	if va.IsRunning() {
		return nil, fmt.Errorf("analysis already running")
	}

	// Set defaults
	if opts.MaxPages <= 0 {
		opts.MaxPages = 10
	}

	// Create analysis record
	analysis := models.NewAnalysis()
	if opts.Query != "" {
		analysis.SetSearchQuery(opts.Query)
	}

	if err := va.analysisRepo.Create(ctx, analysis); err != nil {
		return nil, fmt.Errorf("failed to create analysis record: %w", err)
	}

	// Set running status
	va.setStatus(&AnalysisStatus{
		IsRunning:  true,
		AnalysisID: analysis.ID,
		StartedAt:  analysis.StartedAt,
	})
	defer va.clearStatus()

	startTime := time.Now()

	// Report progress: starting scrape
	va.reportProgress(opts.OnProgress, AnalysisProgress{
		Phase:           "scraping",
		Message:         "Starting eBay scrape...",
		PercentComplete: 0.1,
	})

	// Step 1: Scrape eBay
	scrapeOpts := ebay.ScrapeOptions{
		Query:    opts.Query,
		Category: opts.Category,
		MaxPages: opts.MaxPages,
		Since:    time.Now().AddDate(0, 0, -30),
	}

	scrapeResult, err := va.scraper.ScrapeWithContext(ctx, scrapeOpts)
	if err != nil {
		va.failAnalysis(ctx, analysis, err)
		return nil, fmt.Errorf("scraping failed: %w", err)
	}

	va.reportProgress(opts.OnProgress, AnalysisProgress{
		Phase:           "scraping",
		PagesScraped:    scrapeResult.PagesScraped,
		SalesFound:      scrapeResult.SaleCount(),
		Message:         fmt.Sprintf("Scraped %d pages, found %d sales", scrapeResult.PagesScraped, scrapeResult.SaleCount()),
		PercentComplete: 0.3,
	})

	if scrapeResult.SaleCount() == 0 {
		analysis.Complete(0, 0)
		if err := va.analysisRepo.Update(ctx, analysis); err != nil {
			va.log.Error("Failed to update analysis", "error", err)
		}
		va.log.Info("Analysis completed with no sales", "analysis_id", analysis.ID)
		return &AnalysisResult{
			AnalysisID:    analysis.ID,
			ProductsCount: 0,
			SalesCount:    0,
			Duration:      time.Since(startTime),
		}, nil
	}

	// Step 2: Normalize and save products/sales
	va.reportProgress(opts.OnProgress, AnalysisProgress{
		Phase:           "normalizing",
		SalesFound:      scrapeResult.SaleCount(),
		Message:         "Normalizing product names...",
		PercentComplete: 0.4,
	})

	productsCount, salesCount, err := va.processSales(ctx, analysis.ID, scrapeResult.Sales, opts.OnProgress)
	if err != nil {
		va.failAnalysis(ctx, analysis, err)
		return nil, fmt.Errorf("processing sales failed: %w", err)
	}

	// Step 3: Complete analysis
	analysis.Complete(productsCount, salesCount)
	if err := va.analysisRepo.Update(ctx, analysis); err != nil {
		va.log.Error("Failed to update analysis", "error", err)
	}
	va.log.Info("Analysis completed", "analysis_id", analysis.ID, "products", productsCount, "sales", salesCount, "duration", time.Since(startTime))

	va.reportProgress(opts.OnProgress, AnalysisProgress{
		Phase:           "complete",
		PagesScraped:    scrapeResult.PagesScraped,
		SalesFound:      salesCount,
		ProductsMatched: productsCount,
		Message:         fmt.Sprintf("Analysis complete: %d products, %d sales", productsCount, salesCount),
		PercentComplete: 1.0,
	})

	return &AnalysisResult{
		AnalysisID:    analysis.ID,
		ProductsCount: productsCount,
		SalesCount:    salesCount,
		Duration:      time.Since(startTime),
	}, nil
}

// processSales normalizes raw sales and saves them to the database.
func (va *VolumeAnalyzer) processSales(ctx context.Context, analysisID uuid.UUID, rawSales []ebay.RawSale, onProgress func(AnalysisProgress)) (int, int, error) {
	productMap := make(map[string]*models.Product)
	var salesToCreate []models.Sale
	var processedCount int

	for i, rawSale := range rawSales {
		select {
		case <-ctx.Done():
			return 0, 0, ctx.Err()
		default:
		}

		// Normalize the product name
		normalized, confidence := va.normalizer.Normalize(rawSale.Title)
		if confidence < 0.3 {
			// Skip low confidence matches
			continue
		}

		// Find or create product
		product, ok := productMap[normalized.NormalizedName]
		if !ok {
			// Check database
			dbProduct, err := va.productRepo.FindByNormalizedName(ctx, normalized.NormalizedName)
			if err != nil {
				// Create new product
				newProduct := models.NewProduct(normalized.NormalizedName, normalized.Category)
				if normalized.SetName != "" {
					newProduct.SetSetInfo(normalized.SetName, normalized.SetCode)
				}
				if normalized.MSRP != nil {
					newProduct.SetMSRP(*normalized.MSRP)
				}

				dbProduct, err = va.productRepo.FindOrCreate(ctx, newProduct)
				if err != nil {
					va.log.Warn("Failed to create product", "name", normalized.NormalizedName, "error", err)
					continue
				}
			}
			product = dbProduct
			productMap[normalized.NormalizedName] = product
		}

		// Create sale record
		sale := models.NewSale(
			product.ID,
			rawSale.Title,
			decimal.NewFromFloat(rawSale.Price),
			rawSale.SoldAt,
		)
		sale.SetAnalysisID(analysisID)
		if rawSale.URL != "" {
			sale.SetURL(rawSale.URL)
		}

		salesToCreate = append(salesToCreate, *sale)
		processedCount++

		// Report progress periodically
		if i%50 == 0 {
			va.reportProgress(onProgress, AnalysisProgress{
				Phase:           "saving",
				SalesFound:      len(rawSales),
				ProductsMatched: len(productMap),
				Message:         fmt.Sprintf("Processing sale %d/%d", i+1, len(rawSales)),
				PercentComplete: 0.4 + (float64(i)/float64(len(rawSales)))*0.5,
			})
		}
	}

	// Bulk insert sales
	if len(salesToCreate) > 0 {
		inserted, err := va.saleRepo.BulkCreate(ctx, salesToCreate)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to bulk create sales: %w", err)
		}
		va.log.Debug("Sales inserted", "inserted", inserted, "attempted", len(salesToCreate))
	}

	return len(productMap), len(salesToCreate), nil
}

// GetStatus returns the current analysis status (if running).
func (va *VolumeAnalyzer) GetStatus(ctx context.Context) (*AnalysisStatus, error) {
	va.mu.RLock()
	defer va.mu.RUnlock()

	if va.currentStatus != nil {
		return va.currentStatus, nil
	}

	// Check database for running analysis
	running, err := va.analysisRepo.GetRunning(ctx)
	if err != nil {
		return nil, err
	}
	if running == nil {
		return &AnalysisStatus{IsRunning: false}, nil
	}

	return &AnalysisStatus{
		IsRunning:  true,
		AnalysisID: running.ID,
		StartedAt:  running.StartedAt,
	}, nil
}

// IsRunning returns true if an analysis is currently running.
func (va *VolumeAnalyzer) IsRunning() bool {
	va.mu.RLock()
	defer va.mu.RUnlock()
	return va.currentStatus != nil && va.currentStatus.IsRunning
}

// setStatus sets the current analysis status.
func (va *VolumeAnalyzer) setStatus(status *AnalysisStatus) {
	va.mu.Lock()
	defer va.mu.Unlock()
	va.currentStatus = status
}

// clearStatus clears the current analysis status.
func (va *VolumeAnalyzer) clearStatus() {
	va.mu.Lock()
	defer va.mu.Unlock()
	va.currentStatus = nil
}

// updateProgress updates the progress in the current status.
func (va *VolumeAnalyzer) updateProgress(progress AnalysisProgress) {
	va.mu.Lock()
	defer va.mu.Unlock()
	if va.currentStatus != nil {
		va.currentStatus.Progress = progress
	}
}

// reportProgress reports progress via callback and updates internal status.
func (va *VolumeAnalyzer) reportProgress(onProgress func(AnalysisProgress), progress AnalysisProgress) {
	va.updateProgress(progress)
	if onProgress != nil {
		onProgress(progress)
	}
}

// failAnalysis marks an analysis as failed.
func (va *VolumeAnalyzer) failAnalysis(ctx context.Context, analysis *models.Analysis, err error) {
	analysis.Fail(err)
	if updateErr := va.analysisRepo.Update(ctx, analysis); updateErr != nil {
		va.log.Error("Failed to update failed analysis", "error", updateErr)
	}
	va.log.Error("Analysis failed", "analysis_id", analysis.ID, "error", err)
}

// GetNormalizer returns the normalizer for external use.
func (va *VolumeAnalyzer) GetNormalizer() *Normalizer {
	return va.normalizer
}
