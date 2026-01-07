package commands

import (
	"context"
	"log"
	"time"

	"github.com/SubZzzzzz/pokeprofit/internal/analyzer"
	"github.com/SubZzzzzz/pokeprofit/internal/discord"
	"github.com/SubZzzzzz/pokeprofit/internal/discord/embeds"
	"github.com/bwmarrin/discordgo"
)

// AnalyzeHandler handles the /analyze command.
type AnalyzeHandler struct {
	bot *discord.Bot
}

// NewAnalyzeHandler creates a new analyze command handler.
func NewAnalyzeHandler(bot *discord.Bot) *AnalyzeHandler {
	return &AnalyzeHandler{bot: bot}
}

// Handle processes the /analyze command.
func (h *AnalyzeHandler) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Parse options
	options := i.ApplicationCommandData().Options
	var query, category string

	for _, opt := range options {
		switch opt.Name {
		case "query":
			query = opt.StringValue()
		case "category":
			category = opt.StringValue()
		}
	}

	// Set default query if not provided
	if query == "" {
		query = "Pokemon TCG"
	}

	// Check if analysis is already running
	if h.bot.GetAnalyzer().IsRunning() {
		discord.RespondEmbed(s, i, embeds.AnalysisRunningEmbed())
		return
	}

	// Send deferred response (analysis takes time)
	if err := discord.RespondDeferred(s, i); err != nil {
		log.Printf("Failed to send deferred response: %v", err)
		return
	}

	// Run analysis in background
	go h.runAnalysis(s, i, query, category)
}

func (h *AnalyzeHandler) runAnalysis(s *discordgo.Session, i *discordgo.InteractionCreate, query, category string) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	// Update with initial progress
	discord.EditResponseEmbed(s, i, embeds.AnalysisStartedEmbed())

	opts := analyzer.AnalyzeOptions{
		Query:    query,
		Category: category,
		MaxPages: 10,
		OnProgress: func(progress analyzer.AnalysisProgress) {
			// Update progress periodically
			if progress.Phase == "scraping" || progress.Phase == "saving" {
				embed := embeds.AnalysisProgressEmbed(
					progress.PagesScraped,
					progress.SalesFound,
					progress.Message,
				)
				if err := discord.EditResponseEmbed(s, i, embed); err != nil {
					log.Printf("Failed to update progress: %v", err)
				}
			}
		},
	}

	result, err := h.bot.GetAnalyzer().Run(ctx, opts)
	if err != nil {
		log.Printf("Analysis failed: %v", err)

		suggestion := getSuggestionForError(err)
		embed := embeds.AnalysisFailedEmbed(err.Error(), suggestion)
		if editErr := discord.EditResponseEmbed(s, i, embed); editErr != nil {
			log.Printf("Failed to send error response: %v", editErr)
		}
		return
	}

	// Send success response
	embed := embeds.AnalysisCompleteEmbed(result.ProductsCount, result.SalesCount, result.Duration)
	if err := discord.EditResponseEmbed(s, i, embed); err != nil {
		log.Printf("Failed to send success response: %v", err)
	}
}

func getSuggestionForError(err error) string {
	errMsg := err.Error()

	switch {
	case contains(errMsg, "rate limit"):
		return "eBay limite le nombre de requêtes. Réessayez dans quelques minutes."
	case contains(errMsg, "timeout"), contains(errMsg, "context deadline"):
		return "L'analyse a pris trop de temps. Essayez avec moins de pages."
	case contains(errMsg, "connection"), contains(errMsg, "network"):
		return "Problème de connexion. Vérifiez votre connexion internet."
	case contains(errMsg, "database"):
		return "Problème de base de données. Contactez l'administrateur."
	default:
		return "Réessayez dans quelques minutes. Si le problème persiste, contactez l'administrateur."
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Register registers the analyze command handler with the bot.
func (h *AnalyzeHandler) Register() {
	h.bot.RegisterCommandHandler("analyze", h.Handle)
}
