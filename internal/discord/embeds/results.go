package embeds

import (
	"fmt"
	"strings"
	"time"

	"github.com/SubZzzzzz/pokeprofit/internal/analyzer"
	"github.com/SubZzzzzz/pokeprofit/internal/models"
	"github.com/bwmarrin/discordgo"
)

// Embed color constants
const (
	ColorSuccess  = 0x00FF00 // Green
	ColorError    = 0xFF0000 // Red
	ColorInfo     = 0x3498DB // Blue
	ColorFiltered = 0x9B59B6 // Purple
	ColorWarning  = 0xFFA500 // Orange
	ColorProgress = 0xFFFF00 // Yellow
)

// AnalysisStartedEmbed creates an embed for when analysis starts.
func AnalysisStartedEmbed() *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       "🔄 Analyse en cours...",
		Description: "L'analyse des ventes eBay a démarré. Cela peut prendre quelques minutes.",
		Color:       ColorProgress,
		Timestamp:   time.Now().Format(time.RFC3339),
	}
}

// AnalysisProgressEmbed creates an embed for analysis progress updates.
func AnalysisProgressEmbed(pagesScraped, salesFound int, message string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title: "🔄 Analyse en cours...",
		Color: ColorProgress,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "📑 Pages analysées",
				Value:  fmt.Sprintf("%d", pagesScraped),
				Inline: true,
			},
			{
				Name:   "💰 Ventes trouvées",
				Value:  fmt.Sprintf("%d", salesFound),
				Inline: true,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: message,
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

// AnalysisCompleteEmbed creates an embed for completed analysis.
func AnalysisCompleteEmbed(productsCount, salesCount int, duration time.Duration) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title: "✅ Analyse Terminée",
		Color: ColorSuccess,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "📊 Résultats",
				Value:  fmt.Sprintf("%d produits analysés", productsCount),
				Inline: true,
			},
			{
				Name:   "💰 Ventes",
				Value:  fmt.Sprintf("%d ventes sur 30j", salesCount),
				Inline: true,
			},
			{
				Name:   "⏱️ Durée",
				Value:  formatDuration(duration),
				Inline: true,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Utilisez /results pour voir le classement",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

// AnalysisFailedEmbed creates an embed for failed analysis.
func AnalysisFailedEmbed(errorMsg, suggestion string) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title:       "❌ Erreur d'Analyse",
		Description: errorMsg,
		Color:       ColorError,
		Timestamp:   time.Now().Format(time.RFC3339),
	}

	if suggestion != "" {
		embed.Fields = []*discordgo.MessageEmbedField{
			{
				Name:  "💡 Cause possible",
				Value: suggestion,
			},
		}
	}

	return embed
}

// ResultsEmbed creates an embed for displaying analysis results.
func ResultsEmbed(stats []models.ProductStats, sortBy string, lastAnalysisDate *time.Time) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title:       "📊 Top Produits Pokemon TCG",
		Description: fmt.Sprintf("Classement par %s (30 derniers jours)", getSortLabel(sortBy)),
		Color:       ColorInfo,
		Fields:      make([]*discordgo.MessageEmbedField, 0),
	}

	if len(stats) == 0 {
		embed.Description = "Aucun résultat disponible. Lancez une analyse avec /analyze."
		return embed
	}

	medals := []string{"🥇", "🥈", "🥉"}

	for i, stat := range stats {
		rank := ""
		if i < 3 {
			rank = medals[i] + " "
		} else {
			rank = fmt.Sprintf("%d. ", i+1)
		}

		// Build the field value
		var lines []string

		// Price info with MSRP
		msrpStr := stat.FormatMSRP()
		lines = append(lines, fmt.Sprintf("💰 Prix: %s (MSRP: %s)", stat.FormatAvgPrice(), msrpStr))

		// ROI/Margin info
		roiStr := stat.FormatMarginPercent()
		marginEurStr := formatMarginEUR(&stat)
		if marginEurStr != "" {
			lines = append(lines, fmt.Sprintf("📈 ROI: %s (%s)", roiStr, marginEurStr))
		} else {
			lines = append(lines, fmt.Sprintf("📈 ROI: %s", roiStr))
		}

		// Profitability indicator
		profitLevel := analyzer.GetProfitabilityLevelForStats(&stat)
		profitIndicator := getProfitabilityIndicator(profitLevel)
		if profitIndicator != "" {
			lines = append(lines, profitIndicator)
		}

		// Volume info
		lines = append(lines, fmt.Sprintf("📦 Volume: %d ventes", stat.SalesCount30d))

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   rank + stat.NormalizedName,
			Value:  strings.Join(lines, "\n"),
			Inline: false,
		})

		// Discord limits to 25 fields
		if len(embed.Fields) >= 10 {
			break
		}
	}

	if lastAnalysisDate != nil {
		embed.Footer = &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Dernière mise à jour: %s", lastAnalysisDate.Format("02/01/2006 15:04")),
		}
	}

	return embed
}

// formatMarginEUR formats the margin in EUR for display.
func formatMarginEUR(stat *models.ProductStats) string {
	if stat.MarginEUR == nil {
		return ""
	}
	sign := ""
	if stat.MarginEUR.IsPositive() {
		sign = "+"
	}
	return sign + stat.MarginEUR.Round(2).String() + "€"
}

// getProfitabilityIndicator returns a visual indicator for profitability level.
func getProfitabilityIndicator(level analyzer.ProfitabilityLevel) string {
	switch level {
	case analyzer.ProfitabilityExcellent:
		return "🟢 Excellent profit"
	case analyzer.ProfitabilityGood:
		return "🟡 Bon profit"
	case analyzer.ProfitabilityMarginal:
		return "🟠 Profit marginal"
	case analyzer.ProfitabilityLoss:
		return "🔴 Perte"
	default:
		return ""
	}
}

// FilteredResultsEmbed creates an embed for filtered results.
func FilteredResultsEmbed(stats []models.ProductStats, category models.ProductCategory, lastAnalysisDate *time.Time) *discordgo.MessageEmbed {
	embed := ResultsEmbed(stats, "margin_percent", lastAnalysisDate)
	embed.Title = fmt.Sprintf("📊 Top %s", category.DisplayName())
	embed.Color = ColorFiltered

	return embed
}

// NoDataEmbed creates an embed for when no data is available.
func NoDataEmbed() *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       "📭 Aucune donnée disponible",
		Description: "Aucune analyse n'a été effectuée récemment.",
		Color:       ColorWarning,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "💡 Suggestion",
				Value: "Lancez d'abord une analyse avec `/analyze`",
			},
		},
	}
}

// AnalysisRunningEmbed creates an embed for when an analysis is already running.
func AnalysisRunningEmbed() *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       "⏳ Analyse en cours",
		Description: "Une analyse est déjà en cours. Veuillez attendre qu'elle se termine.",
		Color:       ColorWarning,
	}
}

// RateLimitedEmbed creates an embed for when the user is rate limited.
func RateLimitedEmbed() *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       "⚠️ Trop de requêtes",
		Description: "Vous avez envoyé trop de requêtes. Veuillez attendre quelques secondes.",
		Color:       ColorWarning,
	}
}

// PaginationButtons creates pagination buttons for results.
func PaginationButtons(currentPage, totalPages int, baseID string) []discordgo.MessageComponent {
	if totalPages <= 1 {
		return nil
	}

	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "◀️ Précédent",
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("%s_prev_%d", baseID, currentPage),
					Disabled: currentPage <= 1,
				},
				discordgo.Button{
					Label:    fmt.Sprintf("Page %d/%d", currentPage, totalPages),
					Style:    discordgo.SecondaryButton,
					CustomID: fmt.Sprintf("%s_page_%d", baseID, currentPage),
					Disabled: true,
				},
				discordgo.Button{
					Label:    "Suivant ▶️",
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("%s_next_%d", baseID, currentPage),
					Disabled: currentPage >= totalPages,
				},
			},
		},
	}
}

// Helper functions

func getSortLabel(sortBy string) string {
	switch sortBy {
	case "margin_percent":
		return "marge (%)"
	case "sales_count":
		return "volume de ventes"
	case "avg_price":
		return "prix moyen"
	default:
		return "marge (%)"
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
}

// ProductDetailEmbed creates a detailed embed for a single product.
func ProductDetailEmbed(stat models.ProductStats) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title: stat.NormalizedName,
		Color: ColorInfo,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "📦 Catégorie",
				Value:  stat.Category.DisplayName(),
				Inline: true,
			},
			{
				Name:   "💰 Prix moyen",
				Value:  stat.FormatAvgPrice(),
				Inline: true,
			},
			{
				Name:   "🏷️ MSRP",
				Value:  stat.FormatMSRP(),
				Inline: true,
			},
			{
				Name:   "📈 ROI",
				Value:  stat.FormatMarginPercent(),
				Inline: true,
			},
			{
				Name:   "📊 Volume (30j)",
				Value:  fmt.Sprintf("%d ventes", stat.SalesCount30d),
				Inline: true,
			},
			{
				Name:   "💵 Fourchette de prix",
				Value:  fmt.Sprintf("%s - %s", stat.MinPrice.Round(2).String()+"€", stat.MaxPrice.Round(2).String()+"€"),
				Inline: true,
			},
		},
	}

	if stat.SetName != nil && *stat.SetName != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "🎴 Set",
			Value:  *stat.SetName,
			Inline: true,
		})
	}

	if stat.LastSaleAt != nil {
		embed.Footer = &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Dernière vente: %s", stat.LastSaleAt.Format("02/01/2006")),
		}
	}

	return embed
}

// ErrorEmbed creates a generic error embed.
func ErrorEmbed(title, message string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       "❌ " + title,
		Description: message,
		Color:       ColorError,
		Timestamp:   time.Now().Format(time.RFC3339),
	}
}

// SuccessEmbed creates a generic success embed.
func SuccessEmbed(title, message string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       "✅ " + title,
		Description: message,
		Color:       ColorSuccess,
		Timestamp:   time.Now().Format(time.RFC3339),
	}
}
