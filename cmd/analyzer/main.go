package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/SubZzzzzz/pokeprofit/internal/analyzer"
	"github.com/SubZzzzzz/pokeprofit/internal/config"
	"github.com/SubZzzzzz/pokeprofit/internal/database"
	"github.com/SubZzzzzz/pokeprofit/internal/models"
	"github.com/SubZzzzzz/pokeprofit/internal/repository"
	"github.com/SubZzzzzz/pokeprofit/internal/scraper/ebay"
	"github.com/joho/godotenv"
)

func main() {
	// Parse command line flags
	query := flag.String("query", "Pokemon TCG", "Search query for eBay")
	category := flag.String("category", "", "Product category filter (display, etb, booster, etc.)")
	maxPages := flag.Int("max-pages", 10, "Maximum number of pages to scrape")
	timeout := flag.Duration("timeout", 15*time.Minute, "Timeout for analysis")
	verbose := flag.Bool("verbose", false, "Enable verbose output")
	flag.Parse()

	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		// For CLI mode, we can use defaults for non-Discord config
		cfg = config.LoadWithDefaults()
		if cfg.Database.URL == "" {
			log.Fatalf("DATABASE_URL is required")
		}
	}

	log.Printf("Pokemon TCG Volume Analyzer CLI")
	log.Printf("Query: %s", *query)
	if *category != "" {
		log.Printf("Category: %s", *category)
	}
	log.Printf("Max pages: %d", *maxPages)

	// Initialize database connection
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	db, err := database.Connect(ctx, cfg.Database.URL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Database connection established")

	// Initialize repositories
	productRepo := repository.NewProductRepository(db)
	saleRepo := repository.NewSaleRepository(db)
	analysisRepo := repository.NewAnalysisRepository(db)
	statsRepo := repository.NewStatsRepository(db)

	// Initialize scraper
	scraperCfg := ebay.ScraperConfig{
		RateLimitDelay: cfg.Scraper.RateLimitDelay,
		MaxRetries:     cfg.Scraper.MaxRetries,
		UserAgents:     cfg.Scraper.UserAgents,
	}
	scraper := ebay.NewEbayScraper(scraperCfg)

	// Initialize analyzer
	volumeAnalyzer := analyzer.NewVolumeAnalyzer(scraper, productRepo, saleRepo, analysisRepo)

	// Run analysis
	log.Println("Starting analysis...")
	startTime := time.Now()

	opts := analyzer.AnalyzeOptions{
		Query:    *query,
		Category: *category,
		MaxPages: *maxPages,
		OnProgress: func(progress analyzer.AnalysisProgress) {
			if *verbose {
				fmt.Printf("[%s] Pages: %d, Sales: %d, Products: %d - %s\n",
					progress.Phase,
					progress.PagesScraped,
					progress.SalesFound,
					progress.ProductsMatched,
					progress.Message,
				)
			}
		},
	}

	result, err := volumeAnalyzer.Run(ctx, opts)
	if err != nil {
		log.Fatalf("Analysis failed: %v", err)
	}

	log.Printf("Analysis complete!")
	log.Printf("  Duration: %s", result.Duration.Round(time.Second))
	log.Printf("  Products found: %d", result.ProductsCount)
	log.Printf("  Sales recorded: %d", result.SalesCount)
	log.Printf("  Analysis ID: %s", result.AnalysisID)

	// Display top results
	fmt.Println("\n--- Top Products by Volume ---")
	topByVolume, err := statsRepo.GetTopProductsByVolume(ctx, 10)
	if err != nil {
		log.Printf("Failed to get top products: %v", err)
	} else {
		printProductStats(topByVolume)
	}

	fmt.Println("\n--- Top Products by Margin ---")
	topByMargin, err := statsRepo.GetTopProductsByMargin(ctx, 10)
	if err != nil {
		log.Printf("Failed to get top products: %v", err)
	} else {
		printProductStats(topByMargin)
	}

	elapsed := time.Since(startTime)
	log.Printf("\nTotal execution time: %s", elapsed.Round(time.Second))

	os.Exit(0)
}

func printProductStats(stats []models.ProductStats) {
	if len(stats) == 0 {
		fmt.Println("  No products found")
		return
	}

	for i, stat := range stats {
		fmt.Printf("  %d. %s\n", i+1, stat.NormalizedName)
		fmt.Printf("     Category: %s\n", stat.Category.DisplayName())
		fmt.Printf("     Sales (30d): %d\n", stat.SalesCount30d)
		fmt.Printf("     Avg Price: %s\n", stat.FormatAvgPrice())
		fmt.Printf("     MSRP: %s\n", stat.FormatMSRP())
		fmt.Printf("     Margin: %s\n", stat.FormatMarginPercent())
		fmt.Println()
	}
}
