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

// ResultsHandler handles the /results command.
type ResultsHandler struct {
	bot       *discord.Bot
	statsRepo *repository.StatsRepository
}

// NewResultsHandler creates a new results command handler.
func NewResultsHandler(bot *discord.Bot, statsRepo *repository.StatsRepository) *ResultsHandler {
	return &ResultsHandler{
		bot:       bot,
		statsRepo: statsRepo,
	}
}

// Handle processes the /results command.
func (h *ResultsHandler) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Parse options
	options := i.ApplicationCommandData().Options
	sortBy := "margin_percent" // Default sort
	limit := 10                // Default limit

	for _, opt := range options {
		switch opt.Name {
		case "sort":
			sortBy = opt.StringValue()
		case "limit":
			limit = int(opt.IntValue())
			if limit > 25 {
				limit = 25
			}
			if limit < 1 {
				limit = 1
			}
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get latest analysis date
	latestAnalysis, err := h.bot.GetAnalysisRepo().GetLatest(ctx)
	if err != nil {
		log.Printf("No analysis found: %v", err)
		discord.RespondEmbed(s, i, embeds.NoDataEmbed())
		return
	}

	// Get total count for pagination
	totalStats, err := h.statsRepo.GetProductStats(ctx, repository.StatsOptions{
		SortBy:    sortBy,
		SortOrder: "desc",
		MinSales:  1,
		Limit:     100, // Get a higher count for pagination calculation
	})
	if err != nil {
		log.Printf("Failed to get product stats: %v", err)
		discord.RespondEmbed(s, i, embeds.ErrorEmbed("Erreur", "Impossible de récupérer les statistiques."))
		return
	}

	if len(totalStats) == 0 {
		discord.RespondEmbed(s, i, embeds.NoDataEmbed())
		return
	}

	// Get first page of results (respecting user-specified limit for page size)
	pageSize := limit
	if pageSize > 10 {
		pageSize = 10 // Cap at 10 per page for embed field limits
	}
	stats := totalStats
	if len(stats) > pageSize {
		stats = stats[:pageSize]
	}

	// Build and send response
	var lastDate *time.Time
	if latestAnalysis.CompletedAt != nil {
		lastDate = latestAnalysis.CompletedAt
	}

	embed := embeds.ResultsEmbed(stats, sortBy, lastDate)

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
		response.Data.Components = embeds.PaginationButtons(1, totalPages, "results_"+sortBy)
	}

	if err := s.InteractionRespond(i.Interaction, response); err != nil {
		log.Printf("Failed to respond to results command: %v", err)
	}
}

// Register registers the results command handler with the bot.
func (h *ResultsHandler) Register() {
	h.bot.RegisterCommandHandler("results", h.Handle)
}
