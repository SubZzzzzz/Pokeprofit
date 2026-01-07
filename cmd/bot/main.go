package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SubZzzzzz/pokeprofit/internal/analyzer"
	"github.com/SubZzzzzz/pokeprofit/internal/config"
	"github.com/SubZzzzzz/pokeprofit/internal/database"
	"github.com/SubZzzzzz/pokeprofit/internal/discord"
	"github.com/SubZzzzzz/pokeprofit/internal/discord/commands"
	"github.com/SubZzzzzz/pokeprofit/internal/logger"
	"github.com/SubZzzzzz/pokeprofit/internal/repository"
	"github.com/SubZzzzzz/pokeprofit/internal/scraper/ebay"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if it exists
	_ = godotenv.Load()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Initialize logger
	logger.Init(logger.Config{Level: cfg.Log.Level})
	log := logger.Default().WithComponent("main")

	log.Info("Starting Pokemon TCG Volume Analyzer Bot")

	// Initialize database connection
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := database.Connect(ctx, cfg.Database.URL)
	if err != nil {
		log.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	log.Info("Database connection established")

	// Initialize repositories
	productRepo := repository.NewProductRepository(db)
	saleRepo := repository.NewSaleRepository(db)
	analysisRepo := repository.NewAnalysisRepository(db)
	statsRepo := repository.NewStatsRepository(db)

	// Clean up any stale running analyses
	staleCount, err := analysisRepo.MarkStaleAsFailed(context.Background(), 60)
	if err != nil {
		log.Warn("Failed to clean up stale analyses", "error", err)
	} else if staleCount > 0 {
		log.Info("Cleaned up stale analyses", "count", staleCount)
	}

	// Initialize scraper
	scraperCfg := ebay.ScraperConfig{
		RateLimitDelay: cfg.Scraper.RateLimitDelay,
		MaxRetries:     cfg.Scraper.MaxRetries,
		UserAgents:     cfg.Scraper.UserAgents,
	}
	scraper := ebay.NewEbayScraper(scraperCfg)

	// Initialize analyzer
	volumeAnalyzer := analyzer.NewVolumeAnalyzer(scraper, productRepo, saleRepo, analysisRepo)

	// Initialize Discord bot
	botCfg := discord.Config{
		Token:   cfg.Discord.Token,
		GuildID: cfg.Discord.GuildID,
	}
	bot, err := discord.NewBot(botCfg, volumeAnalyzer, analysisRepo, productRepo, statsRepo)
	if err != nil {
		log.Error("Failed to create Discord bot", "error", err)
		os.Exit(1)
	}

	// Register command handlers
	analyzeHandler := commands.NewAnalyzeHandler(bot)
	analyzeHandler.Register()

	resultsHandler := commands.NewResultsHandler(bot, statsRepo)
	resultsHandler.Register()

	filterHandler := commands.NewFilterCommandHandler(bot, statsRepo)
	filterHandler.Register()

	// Start the bot
	if err := bot.Start(context.Background()); err != nil {
		log.Error("Failed to start Discord bot", "error", err)
		os.Exit(1)
	}
	log.Info("Discord bot started successfully")

	// Wait for interrupt signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	log.Info("Bot is now running. Press Ctrl+C to stop")
	sig := <-stop
	log.Info("Received shutdown signal", "signal", sig.String())

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Graceful shutdown
	log.Info("Starting graceful shutdown...")

	// Create a channel to track shutdown completion
	done := make(chan struct{})

	go func() {
		// Stop the bot (unregisters commands and closes Discord connection)
		if err := bot.Stop(); err != nil {
			log.Error("Error stopping bot", "error", err)
		}

		// Mark any running analyses as failed
		if count, err := analysisRepo.MarkStaleAsFailed(context.Background(), 0); err != nil {
			log.Warn("Failed to mark running analyses as failed", "error", err)
		} else if count > 0 {
			log.Info("Marked running analyses as failed", "count", count)
		}

		// Close database connection
		if err := db.Close(); err != nil {
			log.Error("Error closing database connection", "error", err)
		}

		close(done)
	}()

	// Wait for shutdown to complete or timeout
	select {
	case <-done:
		log.Info("Graceful shutdown completed")
	case <-shutdownCtx.Done():
		log.Warn("Shutdown timed out, forcing exit")
	}

	log.Info("Bot stopped successfully")
}
