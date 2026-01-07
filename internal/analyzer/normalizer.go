package analyzer

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/SubZzzzzz/pokeprofit/internal/models"
	"github.com/shopspring/decimal"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// NormalizedProduct represents a recognized product from a raw sale title.
type NormalizedProduct struct {
	NormalizedName string
	Category       models.ProductCategory
	SetName        string
	SetCode        string
	MSRP           *decimal.Decimal
	Confidence     float64
}

// ProductPattern defines a recognition pattern for normalizing products.
type ProductPattern struct {
	SetKeywords     []string
	TypeKeywords    []string
	ExcludeKeywords []string
	NormalizedName  string
	SetName         string
	SetCode         string
	Category        models.ProductCategory
	MSRP            float64
}

// Normalizer converts raw sale titles into normalized products.
type Normalizer struct {
	patterns      []ProductPattern
	setPatterns   map[string]*regexp.Regexp
	typePatterns  map[models.ProductCategory][]string
	excludeWords  []string
	transformer   transform.Transformer
}

// NewNormalizer creates a new product normalizer with default patterns.
func NewNormalizer() *Normalizer {
	n := &Normalizer{
		transformer: transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC),
	}

	n.initializePatterns()
	return n
}

func (n *Normalizer) initializePatterns() {
	// Set patterns - regexp for identifying Pokemon TCG sets
	n.setPatterns = map[string]*regexp.Regexp{
		"sv-151":           regexp.MustCompile(`(?i)(151|ecarlate.*violet.*151|scarlet.*violet.*151)`),
		"sv-paldean-fates": regexp.MustCompile(`(?i)(paldea|paldean|destin[eé]es?\s*de\s*paldea|paldean\s*fates)`),
		"sv-prismatic-evo": regexp.MustCompile(`(?i)([eé]volutions?\s*prismatiques?|prismatic\s*evolutions?)`),
		"sv-twilight":      regexp.MustCompile(`(?i)(masques?\s*du\s*cr[eé]puscule|twilight\s*masquerade?)`),
		"sv-temporal":      regexp.MustCompile(`(?i)(forces?\s*temporelles?|temporal\s*forces?)`),
		"sv-obsidian":      regexp.MustCompile(`(?i)(flammes?\s*obsidiennes?|obsidian\s*flames?)`),
		"sv-paradox":       regexp.MustCompile(`(?i)(faille\s*paradoxe?|paradox\s*rift)`),
		"sv-base":          regexp.MustCompile(`(?i)([eé]carlate.*violet|scarlet.*violet)`),
	}

	// Type patterns - keywords for identifying product types
	n.typePatterns = map[models.ProductCategory][]string{
		models.CategoryDisplay: {
			"display", "boite 36", "boîte 36", "box 36", "booster box",
			"36 boosters", "36 packs", "coffret 36",
		},
		models.CategoryETB: {
			"etb", "elite trainer", "coffret dresseur", "trainer box",
			"coffret d'entrainement", "coffret entrainement",
		},
		models.CategoryCollection: {
			"coffret", "collection", "premium collection", "ultra premium",
			"upc", "special collection", "coffret premium",
		},
		models.CategoryBundle: {
			"bundle", "pack 6", "6 boosters", "pack boosters",
		},
		models.CategoryTin: {
			"tin", "pokebox", "poke box", "metal box", "boite metal",
		},
		models.CategoryBooster: {
			"booster", "pack", "pochette", "sachet",
		},
		models.CategorySingle: {
			"carte", "card", "holo", "reverse", "full art", "alt art",
			"secret rare", "illustration rare", "special art",
		},
	}

	// Words to exclude (likely not Pokemon TCG products)
	n.excludeWords = []string{
		"lot de", "bundle lot", "fake", "proxy", "custom",
		"yugioh", "yu-gi-oh", "magic", "mtg", "one piece",
		"digimon", "dragon ball", "weiss schwarz",
	}

	// Pre-defined product patterns for known products
	n.patterns = []ProductPattern{
		// Displays
		{
			SetKeywords:    []string{"151"},
			TypeKeywords:   []string{"display", "boite 36", "36 boosters"},
			NormalizedName: "Display Écarlate et Violet 151",
			SetName:        "Écarlate et Violet 151",
			SetCode:        "sv-151",
			Category:       models.CategoryDisplay,
			MSRP:           159.99,
		},
		{
			SetKeywords:    []string{"paldea", "paldean", "destinées"},
			TypeKeywords:   []string{"display", "boite 36", "36 boosters"},
			NormalizedName: "Display Destinées de Paldea",
			SetName:        "Destinées de Paldea",
			SetCode:        "sv-paldean-fates",
			Category:       models.CategoryDisplay,
			MSRP:           159.99,
		},
		{
			SetKeywords:    []string{"prismatique", "prismatic"},
			TypeKeywords:   []string{"display", "boite 36", "36 boosters"},
			NormalizedName: "Display Évolutions Prismatiques",
			SetName:        "Évolutions Prismatiques",
			SetCode:        "sv-prismatic-evo",
			Category:       models.CategoryDisplay,
			MSRP:           159.99,
		},
		{
			SetKeywords:    []string{"crépuscule", "twilight", "masque"},
			TypeKeywords:   []string{"display", "boite 36", "36 boosters"},
			NormalizedName: "Display Masques du Crépuscule",
			SetName:        "Masques du Crépuscule",
			SetCode:        "sv-twilight",
			Category:       models.CategoryDisplay,
			MSRP:           159.99,
		},
		{
			SetKeywords:    []string{"temporelles", "temporal"},
			TypeKeywords:   []string{"display", "boite 36", "36 boosters"},
			NormalizedName: "Display Forces Temporelles",
			SetName:        "Forces Temporelles",
			SetCode:        "sv-temporal",
			Category:       models.CategoryDisplay,
			MSRP:           159.99,
		},
		// ETBs
		{
			SetKeywords:    []string{"151"},
			TypeKeywords:   []string{"etb", "elite trainer", "coffret dresseur"},
			NormalizedName: "ETB Écarlate et Violet 151",
			SetName:        "Écarlate et Violet 151",
			SetCode:        "sv-151",
			Category:       models.CategoryETB,
			MSRP:           54.99,
		},
		{
			SetKeywords:    []string{"paldea", "paldean", "destinées"},
			TypeKeywords:   []string{"etb", "elite trainer", "coffret dresseur"},
			NormalizedName: "ETB Destinées de Paldea",
			SetName:        "Destinées de Paldea",
			SetCode:        "sv-paldean-fates",
			Category:       models.CategoryETB,
			MSRP:           54.99,
		},
		{
			SetKeywords:    []string{"prismatique", "prismatic"},
			TypeKeywords:   []string{"etb", "elite trainer", "coffret dresseur"},
			NormalizedName: "ETB Évolutions Prismatiques",
			SetName:        "Évolutions Prismatiques",
			SetCode:        "sv-prismatic-evo",
			Category:       models.CategoryETB,
			MSRP:           54.99,
		},
		// Ultra Premium Collections
		{
			SetKeywords:    []string{"dracaufeu", "charizard"},
			TypeKeywords:   []string{"ultra premium", "upc"},
			NormalizedName: "Coffret Dracaufeu Ultra Premium",
			SetName:        "Écarlate et Violet",
			SetCode:        "sv-charizard-upc",
			Category:       models.CategoryCollection,
			MSRP:           119.99,
		},
		{
			SetKeywords:    []string{"mew", "151"},
			TypeKeywords:   []string{"ultra premium", "upc"},
			NormalizedName: "Coffret Mew Ultra Premium",
			SetName:        "Écarlate et Violet 151",
			SetCode:        "sv-151-upc",
			Category:       models.CategoryCollection,
			MSRP:           119.99,
		},
	}
}

// Normalize extracts product information from a raw sale title.
// Returns the matched product and a confidence score (0.0-1.0).
func (n *Normalizer) Normalize(title string) (NormalizedProduct, float64) {
	// Normalize the title
	normalized := n.normalizeText(title)

	// Check for excluded keywords
	if n.containsExcluded(normalized) {
		return NormalizedProduct{}, 0.0
	}

	// Try to match against known patterns first
	if product, confidence := n.matchKnownPattern(normalized); confidence > 0 {
		return product, confidence
	}

	// Fall back to generic matching
	return n.genericMatch(normalized, title)
}

// normalizeText prepares text for matching.
func (n *Normalizer) normalizeText(text string) string {
	// Convert to lowercase
	text = strings.ToLower(text)

	// Remove accents
	result, _, _ := transform.String(n.transformer, text)

	// Remove special characters except spaces and alphanumerics
	result = regexp.MustCompile(`[^a-z0-9\s]`).ReplaceAllString(result, " ")

	// Normalize multiple spaces
	result = regexp.MustCompile(`\s+`).ReplaceAllString(result, " ")

	return strings.TrimSpace(result)
}

// containsExcluded checks if the text contains any excluded keywords.
func (n *Normalizer) containsExcluded(text string) bool {
	for _, word := range n.excludeWords {
		if strings.Contains(text, word) {
			return true
		}
	}
	return false
}

// matchKnownPattern tries to match against pre-defined patterns.
func (n *Normalizer) matchKnownPattern(normalized string) (NormalizedProduct, float64) {
	var bestMatch NormalizedProduct
	var bestScore float64

	for _, pattern := range n.patterns {
		score := n.calculatePatternScore(normalized, pattern)
		if score > bestScore {
			bestScore = score
			msrp := decimal.NewFromFloat(pattern.MSRP)
			bestMatch = NormalizedProduct{
				NormalizedName: pattern.NormalizedName,
				Category:       pattern.Category,
				SetName:        pattern.SetName,
				SetCode:        pattern.SetCode,
				MSRP:           &msrp,
				Confidence:     score,
			}
		}
	}

	return bestMatch, bestScore
}

// calculatePatternScore calculates how well a title matches a pattern.
func (n *Normalizer) calculatePatternScore(normalized string, pattern ProductPattern) float64 {
	var setMatches, typeMatches int

	// Check set keywords
	for _, keyword := range pattern.SetKeywords {
		if strings.Contains(normalized, strings.ToLower(keyword)) {
			setMatches++
		}
	}

	// Check type keywords
	for _, keyword := range pattern.TypeKeywords {
		if strings.Contains(normalized, strings.ToLower(keyword)) {
			typeMatches++
		}
	}

	// Check exclude keywords
	for _, keyword := range pattern.ExcludeKeywords {
		if strings.Contains(normalized, strings.ToLower(keyword)) {
			return 0.0 // Immediate disqualification
		}
	}

	// Need at least one match from each category
	if setMatches == 0 || typeMatches == 0 {
		return 0.0
	}

	// Calculate score based on number of matches
	setScore := float64(setMatches) / float64(len(pattern.SetKeywords))
	typeScore := float64(typeMatches) / float64(len(pattern.TypeKeywords))

	// Combined score (weight type slightly higher as it's more specific)
	return (setScore*0.4 + typeScore*0.6) * 0.9 // Max 0.9 for pattern matches
}

// genericMatch attempts to identify product type and set without a specific pattern.
func (n *Normalizer) genericMatch(normalized, originalTitle string) (NormalizedProduct, float64) {
	product := NormalizedProduct{
		NormalizedName: originalTitle,
		Confidence:     0.3, // Low confidence for generic matches
	}

	// Identify category
	for category, keywords := range n.typePatterns {
		for _, keyword := range keywords {
			if strings.Contains(normalized, strings.ToLower(keyword)) {
				product.Category = category
				product.Confidence = 0.5 // Slightly higher if we identified category
				break
			}
		}
		if product.Category != "" {
			break
		}
	}

	// Default to single if no category identified
	if product.Category == "" {
		product.Category = models.CategorySingle
	}

	// Try to identify set
	for setCode, pattern := range n.setPatterns {
		if pattern.MatchString(normalized) {
			product.SetCode = setCode
			product.Confidence += 0.2 // Bonus for identifying set
			break
		}
	}

	// Generate normalized name based on identified components
	product.NormalizedName = n.generateNormalizedName(product.Category, product.SetCode, originalTitle)

	return product, product.Confidence
}

// generateNormalizedName creates a canonical name from components.
func (n *Normalizer) generateNormalizedName(category models.ProductCategory, setCode, originalTitle string) string {
	if setCode == "" {
		// Can't generate a meaningful name without set info
		// Truncate original title to reasonable length
		if len(originalTitle) > 50 {
			return originalTitle[:50] + "..."
		}
		return originalTitle
	}

	setNames := map[string]string{
		"sv-151":           "Écarlate et Violet 151",
		"sv-paldean-fates": "Destinées de Paldea",
		"sv-prismatic-evo": "Évolutions Prismatiques",
		"sv-twilight":      "Masques du Crépuscule",
		"sv-temporal":      "Forces Temporelles",
		"sv-obsidian":      "Flammes Obsidiennes",
		"sv-paradox":       "Faille Paradoxe",
		"sv-base":          "Écarlate et Violet",
	}

	setName := setNames[setCode]
	if setName == "" {
		setName = setCode
	}

	categoryPrefixes := map[models.ProductCategory]string{
		models.CategoryDisplay:    "Display",
		models.CategoryETB:        "ETB",
		models.CategoryCollection: "Coffret",
		models.CategoryBundle:     "Bundle",
		models.CategoryTin:        "Tin",
		models.CategoryBooster:    "Booster",
		models.CategorySingle:     "Carte",
	}

	prefix := categoryPrefixes[category]
	if prefix == "" {
		prefix = "Produit"
	}

	return prefix + " " + setName
}

// AddPattern adds a new recognition pattern.
func (n *Normalizer) AddPattern(pattern ProductPattern) {
	n.patterns = append(n.patterns, pattern)
}

// GetPatternCount returns the number of registered patterns.
func (n *Normalizer) GetPatternCount() int {
	return len(n.patterns)
}
