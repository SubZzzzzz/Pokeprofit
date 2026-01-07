package discord

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/SubZzzzzz/pokeprofit/internal/analyzer"
	"github.com/SubZzzzzz/pokeprofit/internal/discord/embeds"
	"github.com/SubZzzzzz/pokeprofit/internal/logger"
	"github.com/SubZzzzzz/pokeprofit/internal/models"
	"github.com/SubZzzzzz/pokeprofit/internal/repository"
	"github.com/bwmarrin/discordgo"
)

// Bot represents the Discord bot.
type Bot struct {
	session       *discordgo.Session
	guildID       string
	analyzer      *analyzer.VolumeAnalyzer
	analysisRepo  *repository.AnalysisRepository
	productRepo   *repository.ProductRepository
	statsRepo     *repository.StatsRepository

	commands        []*discordgo.ApplicationCommand
	commandHandlers map[string]func(*discordgo.Session, *discordgo.InteractionCreate)
	buttonHandlers  map[string]func(*discordgo.Session, *discordgo.InteractionCreate, ButtonContext)

	rateLimiter *RateLimiter
	log         *logger.Logger

	mu            sync.RWMutex
	isRunning     bool
}

// ButtonContext contains parsed information from a button custom_id.
type ButtonContext struct {
	BaseID    string // e.g., "results_margin_percent" or "filter_display"
	Action    string // "prev", "next", "page"
	Page      int    // current page number
}

// Config holds Discord bot configuration.
type Config struct {
	Token   string
	GuildID string
}

// NewBot creates a new Discord bot instance.
func NewBot(cfg Config, analyzer *analyzer.VolumeAnalyzer, analysisRepo *repository.AnalysisRepository, productRepo *repository.ProductRepository, statsRepo *repository.StatsRepository) (*Bot, error) {
	session, err := discordgo.New("Bot " + cfg.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to create Discord session: %w", err)
	}

	log := logger.Default().WithComponent("discord")

	bot := &Bot{
		session:         session,
		guildID:         cfg.GuildID,
		analyzer:        analyzer,
		analysisRepo:    analysisRepo,
		productRepo:     productRepo,
		statsRepo:       statsRepo,
		commandHandlers: make(map[string]func(*discordgo.Session, *discordgo.InteractionCreate)),
		buttonHandlers:  make(map[string]func(*discordgo.Session, *discordgo.InteractionCreate, ButtonContext)),
		rateLimiter:     NewDefaultRateLimiter(),
		log:             log,
	}

	// Define slash commands
	bot.commands = []*discordgo.ApplicationCommand{
		{
			Name:        "analyze",
			Description: "Lance une analyse de volume des ventes Pokemon TCG sur eBay",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "query",
					Description: "Terme de recherche (ex: Pokemon Display 151)",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "category",
					Description: "Catégorie de produit à analyser",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Tous", Value: "all"},
						{Name: "Displays", Value: "display"},
						{Name: "ETB", Value: "etb"},
						{Name: "Coffrets", Value: "collection"},
						{Name: "Boosters", Value: "booster"},
					},
				},
			},
		},
		{
			Name:        "results",
			Description: "Affiche les résultats de la dernière analyse de volume",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "sort",
					Description: "Critère de tri",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Marge (%) - Recommandé", Value: "margin_percent"},
						{Name: "Volume de ventes", Value: "sales_count"},
						{Name: "Prix moyen", Value: "avg_price"},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "limit",
					Description: "Nombre de résultats (max 25)",
					Required:    false,
					MinValue:    floatPtr(1),
					MaxValue:    25,
				},
			},
		},
		{
			Name:        "filter",
			Description: "Filtre les résultats par catégorie de produit",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "category",
					Description: "Catégorie à afficher",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Displays (Boîtes 36 boosters)", Value: "display"},
						{Name: "ETB (Elite Trainer Box)", Value: "etb"},
						{Name: "Coffrets / Collections", Value: "collection"},
						{Name: "Boosters individuels", Value: "booster"},
						{Name: "Bundles (6 boosters)", Value: "bundle"},
						{Name: "Tins / Pokebox", Value: "tin"},
					},
				},
			},
		},
	}

	return bot, nil
}

// Start starts the Discord bot.
func (b *Bot) Start(ctx context.Context) error {
	b.mu.Lock()
	if b.isRunning {
		b.mu.Unlock()
		return fmt.Errorf("bot is already running")
	}
	b.isRunning = true
	b.mu.Unlock()

	// Add handler for interactions
	b.session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		b.handleInteraction(s, i)
	})

	// Set intents
	b.session.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages

	// Open connection
	if err := b.session.Open(); err != nil {
		return fmt.Errorf("failed to open Discord connection: %w", err)
	}

	b.log.Info("Discord bot connected", "username", b.session.State.User.Username)

	// Register commands
	if err := b.registerCommands(); err != nil {
		b.session.Close()
		return fmt.Errorf("failed to register commands: %w", err)
	}

	b.log.Info("Slash commands registered", "count", len(b.commands))

	return nil
}

// Stop stops the Discord bot.
func (b *Bot) Stop() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.isRunning {
		return nil
	}

	b.isRunning = false

	// Stop rate limiter
	if b.rateLimiter != nil {
		b.rateLimiter.Stop()
	}

	// Unregister commands before closing
	if err := b.unregisterCommands(); err != nil {
		b.log.Warn("Failed to unregister commands", "error", err)
	}

	b.log.Info("Discord bot stopping")
	return b.session.Close()
}

// registerCommands registers slash commands with Discord.
func (b *Bot) registerCommands() error {
	for _, cmd := range b.commands {
		_, err := b.session.ApplicationCommandCreate(b.session.State.User.ID, b.guildID, cmd)
		if err != nil {
			return fmt.Errorf("failed to register command %s: %w", cmd.Name, err)
		}
		b.log.Debug("Command registered", "command", cmd.Name)
	}
	return nil
}

// unregisterCommands removes slash commands from Discord.
func (b *Bot) unregisterCommands() error {
	registeredCmds, err := b.session.ApplicationCommands(b.session.State.User.ID, b.guildID)
	if err != nil {
		return fmt.Errorf("failed to get registered commands: %w", err)
	}

	for _, cmd := range registeredCmds {
		if err := b.session.ApplicationCommandDelete(b.session.State.User.ID, b.guildID, cmd.ID); err != nil {
			b.log.Warn("Failed to delete command", "command", cmd.Name, "error", err)
		}
	}

	return nil
}

// handleInteraction handles incoming Discord interactions.
func (b *Bot) handleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		b.handleApplicationCommand(s, i)
	case discordgo.InteractionMessageComponent:
		b.handleMessageComponent(s, i)
	}
}

// handleApplicationCommand handles slash command interactions.
func (b *Bot) handleApplicationCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	commandName := i.ApplicationCommandData().Name
	userID := ""
	if i.Member != nil && i.Member.User != nil {
		userID = i.Member.User.ID
	} else if i.User != nil {
		userID = i.User.ID
	}

	// Check rate limit
	if userID != "" && !b.rateLimiter.Allow(userID, commandName) {
		waitTime := b.rateLimiter.TimeUntilAllowed(userID, commandName)
		b.log.Debug("Rate limited user", "user_id", userID, "command", commandName, "wait_time", waitTime)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Trop de requetes. Reessayez dans %s.", formatDuration(waitTime)),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	b.log.Debug("Handling command", "command", commandName, "user_id", userID)

	// Check if we have a registered handler
	if handler, ok := b.commandHandlers[commandName]; ok {
		handler(s, i)
		return
	}

	// Respond with error for unknown commands
	b.log.Warn("Unknown command received", "command", commandName)
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Unknown command",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

// formatDuration formats a duration for user display.
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%d secondes", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%d minutes", int(d.Minutes()))
	}
	return fmt.Sprintf("%d heures", int(d.Hours()))
}

// handleMessageComponent handles button and select menu interactions.
func (b *Bot) handleMessageComponent(s *discordgo.Session, i *discordgo.InteractionCreate) {
	customID := i.MessageComponentData().CustomID

	// Parse the custom_id format: baseID_action_page
	// e.g., "results_margin_percent_next_1" or "filter_display_prev_2"
	ctx := parseButtonCustomID(customID)
	if ctx == nil {
		b.log.Warn("Failed to parse button custom_id", "custom_id", customID)
		return
	}

	// Determine the handler type based on the baseID prefix
	handlerKey := ""
	if strings.HasPrefix(ctx.BaseID, "results_") {
		handlerKey = "results"
	} else if strings.HasPrefix(ctx.BaseID, "filter_") {
		handlerKey = "filter"
	}

	// Check if we have a registered handler
	if handler, ok := b.buttonHandlers[handlerKey]; ok {
		handler(s, i, *ctx)
		return
	}

	// Default pagination handler
	b.handlePaginationButton(s, i, *ctx)
}

// parseButtonCustomID parses a button custom_id into its components.
// Format: baseID_action_page (e.g., "results_margin_percent_next_1")
func parseButtonCustomID(customID string) *ButtonContext {
	parts := strings.Split(customID, "_")
	if len(parts) < 3 {
		return nil
	}

	// The last part is the page number
	page, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		return nil
	}

	// The second-to-last part is the action
	action := parts[len(parts)-2]
	if action != "prev" && action != "next" && action != "page" {
		return nil
	}

	// Everything before is the baseID
	baseID := strings.Join(parts[:len(parts)-2], "_")

	return &ButtonContext{
		BaseID: baseID,
		Action: action,
		Page:   page,
	}
}

// RegisterCommandHandler registers a handler for a slash command.
func (b *Bot) RegisterCommandHandler(name string, handler func(*discordgo.Session, *discordgo.InteractionCreate)) {
	b.commandHandlers[name] = handler
}

// RegisterButtonHandler registers a handler for button interactions.
func (b *Bot) RegisterButtonHandler(name string, handler func(*discordgo.Session, *discordgo.InteractionCreate, ButtonContext)) {
	b.buttonHandlers[name] = handler
}

// GetSession returns the Discord session.
func (b *Bot) GetSession() *discordgo.Session {
	return b.session
}

// GetAnalyzer returns the volume analyzer.
func (b *Bot) GetAnalyzer() *analyzer.VolumeAnalyzer {
	return b.analyzer
}

// GetAnalysisRepo returns the analysis repository.
func (b *Bot) GetAnalysisRepo() *repository.AnalysisRepository {
	return b.analysisRepo
}

// GetStatsRepo returns the stats repository.
func (b *Bot) GetStatsRepo() *repository.StatsRepository {
	return b.statsRepo
}

// IsRunning returns true if the bot is running.
func (b *Bot) IsRunning() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.isRunning
}

// Helper function for float pointer
func floatPtr(f float64) *float64 {
	return &f
}

// Respond sends a response to an interaction.
func Respond(s *discordgo.Session, i *discordgo.InteractionCreate, content string) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
}

// RespondDeferred sends a deferred response to an interaction.
func RespondDeferred(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
}

// RespondEmbed sends an embed response to an interaction.
func RespondEmbed(s *discordgo.Session, i *discordgo.InteractionCreate, embed *discordgo.MessageEmbed) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

// EditResponse edits the original response to an interaction.
func EditResponse(s *discordgo.Session, i *discordgo.InteractionCreate, content string) error {
	_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})
	return err
}

// EditResponseEmbed edits the original response with an embed.
func EditResponseEmbed(s *discordgo.Session, i *discordgo.InteractionCreate, embed *discordgo.MessageEmbed) error {
	_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	})
	return err
}

// RespondError sends an error response to an interaction.
func RespondError(s *discordgo.Session, i *discordgo.InteractionCreate, title, message string) error {
	embed := &discordgo.MessageEmbed{
		Title:       "❌ " + title,
		Description: message,
		Color:       0xFF0000, // Red
	}
	return RespondEmbed(s, i, embed)
}

// handlePaginationButton handles pagination button clicks for results and filter commands.
func (b *Bot) handlePaginationButton(s *discordgo.Session, i *discordgo.InteractionCreate, ctx ButtonContext) {
	// Calculate new page
	newPage := ctx.Page
	switch ctx.Action {
	case "prev":
		newPage = ctx.Page - 1
	case "next":
		newPage = ctx.Page + 1
	case "page":
		// Page indicator button, just acknowledge
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredMessageUpdate,
		})
		return
	}

	if newPage < 1 {
		newPage = 1
	}

	// Determine if this is a results or filter pagination
	isFilter := strings.HasPrefix(ctx.BaseID, "filter_")

	// Parse the sort/category from baseID
	var sortBy, category string
	if isFilter {
		// baseID format: "filter_category"
		category = strings.TrimPrefix(ctx.BaseID, "filter_")
	} else {
		// baseID format: "results_sortBy"
		sortBy = strings.TrimPrefix(ctx.BaseID, "results_")
		if sortBy == "" {
			sortBy = "margin_percent"
		}
	}

	// Fetch the data
	dbCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pageSize := 10
	offset := (newPage - 1) * pageSize

	// Build stats options
	opts := repository.StatsOptions{
		SortBy:    sortBy,
		SortOrder: "desc",
		MinSales:  1,
		Limit:     100, // Get enough for pagination calculation
	}
	if isFilter {
		opts.Category = category
		opts.SortBy = "margin_percent"
	}

	// Get all stats for pagination calculation
	allStats, err := b.statsRepo.GetProductStats(dbCtx, opts)
	if err != nil {
		b.log.Error("Failed to get product stats for pagination", "error", err)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content: "Erreur lors de la récupération des données.",
			},
		})
		return
	}

	totalItems := len(allStats)
	totalPages := (totalItems + pageSize - 1) / pageSize

	// Adjust page if out of bounds
	if newPage > totalPages {
		newPage = totalPages
	}

	// Get the page slice
	start := offset
	end := offset + pageSize
	if start >= len(allStats) {
		start = 0
		end = pageSize
		newPage = 1
	}
	if end > len(allStats) {
		end = len(allStats)
	}
	pageStats := allStats[start:end]

	// Get last analysis date
	var lastDate *time.Time
	latestAnalysis, err := b.analysisRepo.GetLatest(dbCtx)
	if err == nil && latestAnalysis.CompletedAt != nil {
		lastDate = latestAnalysis.CompletedAt
	}

	// Build the embed
	var embed *discordgo.MessageEmbed
	if isFilter {
		productCategory := models.ProductCategory(category)
		embed = embeds.FilteredResultsEmbed(pageStats, productCategory, lastDate)
	} else {
		embed = embeds.ResultsEmbed(pageStats, sortBy, lastDate)
	}

	// Add page indicator to footer
	if embed.Footer == nil {
		embed.Footer = &discordgo.MessageEmbedFooter{}
	}
	embed.Footer.Text = fmt.Sprintf("Page %d/%d | %s", newPage, totalPages, embed.Footer.Text)

	// Build the response
	response := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: embeds.PaginationButtons(newPage, totalPages, ctx.BaseID),
		},
	}

	if err := s.InteractionRespond(i.Interaction, response); err != nil {
		b.log.Error("Failed to update pagination", "error", err, "page", newPage)
	}
}

// GetRateLimiter returns the rate limiter for external use.
func (b *Bot) GetRateLimiter() *RateLimiter {
	return b.rateLimiter
}
