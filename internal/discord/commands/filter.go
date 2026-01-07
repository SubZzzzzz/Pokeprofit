package commands

import (
	"context"
	"log"
	"time"

	"github.com/SubZzzzzz/pokeprofit/internal/discord"
	"github.com/SubZzzzzz/pokeprofit/internal/discord/embeds"
	"github.com/SubZzzzzz/pokeprofit/internal/repository"
	"github.com/bwmarrin/discordgo"
)

// FilterCommandHandler handles the /filter command.
type FilterCommandHandler struct {
	bot       *discord.Bot
	statsRepo *repository.StatsRepository
}

// NewFilterCommandHandler creates a new filter command handler.
func NewFilterCommandHandler(bot *discord.Bot, statsRepo *repository.StatsRepository) *FilterCommandHandler {
	return &FilterCommandHandler{
		bot:       bot,
		statsRepo: statsRepo,
	}
}

// Handle processes the /filter command.
func (h *FilterCommandHandler) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Parse options
	options := i.ApplicationCommandData().Options
	var category string

	for _, opt := range options {
		if opt.Name == "category" {
			category = opt.StringValue()
			break
		}
	}

	if category == "" {
		discord.RespondError(s, i, "Erreur", "La catégorie est requise.")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get latest analysis to verify data exists
	latestAnalysis, err := h.bot.GetAnalysisRepo().GetLatest(ctx)
	if err != nil {
		log.Printf("No analysis found: %v", err)
		discord.RespondEmbed(s, i, embeds.NoDataEmbed())
		return
	}

	// Count total items for pagination
	totalStats, err := h.statsRepo.GetProductStats(ctx, repository.StatsOptions{
		Category:  category,
		SortBy:    "margin_percent",
		SortOrder: "desc",
		MinSales:  1,
		Limit:     100, // Get a higher count for pagination calculation
	})
	if err != nil {
		log.Printf("Failed to count filtered stats: %v", err)
		discord.RespondEmbed(s, i, embeds.ErrorEmbed("Erreur", "Impossible de récupérer les statistiques."))
		return
	}

	if len(totalStats) == 0 {
		discord.RespondEmbed(s, i, embeds.ErrorEmbed(
			"Aucun résultat",
			"Aucun produit trouvé pour cette catégorie.",
		))
		return
	}

	// Get first page of results
	pageSize := 10
	stats := totalStats
	if len(stats) > pageSize {
		stats = stats[:pageSize]
	}

	// Build and send response
	var lastDate *time.Time
	if latestAnalysis.CompletedAt != nil {
		lastDate = latestAnalysis.CompletedAt
	}

	// Get category from first result for display
	productCategory := stats[0].Category
	embed := embeds.FilteredResultsEmbed(stats, productCategory, lastDate)

	// Calculate total pages
	totalItems := len(totalStats)
	totalPages := (totalItems + pageSize - 1) / pageSize

	// Create response with pagination buttons if needed
	response := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	}

	// Add pagination buttons if there are multiple pages
	if totalPages > 1 {
		response.Data.Components = embeds.PaginationButtons(1, totalPages, "filter_"+category)
	}

	if err := s.InteractionRespond(i.Interaction, response); err != nil {
		log.Printf("Failed to respond to filter command: %v", err)
	}
}

// Register registers the filter command handler with the bot.
func (h *FilterCommandHandler) Register() {
	h.bot.RegisterCommandHandler("filter", h.Handle)
}
