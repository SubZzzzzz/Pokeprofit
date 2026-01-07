# Research: Volume Analyzer Phase 1 - eBay

**Date**: 2026-01-07
**Feature Branch**: `001-ebay-volume-analyzer`

## Summary

Research completed for eBay Volume Analyzer implementation. All technical decisions are aligned with project skills and constitution.

---

## 1. eBay Scraping Strategy

### Decision: HTML scraping with colly (ventes terminées)

**Rationale**:
- eBay completed listings are server-rendered HTML, no JS required
- colly is the prescribed stack for static HTML sites (per constitution)
- Simpler and more reliable than chromedp for this use case

**URL Structure for eBay FR Completed Sales**:
```
https://www.ebay.fr/sch/i.html?_nkw={search_term}&_sacat=0&LH_Complete=1&LH_Sold=1&_sop=13
```
- `LH_Complete=1` : Show completed listings
- `LH_Sold=1` : Show sold items only
- `_sop=13` : Sort by end date (most recent first)
- `_sacat=183454` : Pokemon TCG category (optional filter)

**Key HTML Selectors** (subject to verification):
- Product container: `.s-item`
- Title: `.s-item__title`
- Price: `.s-item__price`
- Sold date: `.s-item__title--tag` or `.POSITIVE`
- Item link: `.s-item__link`

**Alternatives Rejected**:
- eBay API: Requires approval, limited completed sales data
- chromedp: Overkill for static HTML pages, slower

---

## 2. Colly Configuration

### Decision: Use colly with rate limiting and retry logic

**Configuration** (from project skill):
```go
c := colly.NewCollector(
    colly.AllowedDomains("www.ebay.fr"),
    colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"),
)

c.Limit(&colly.LimitRule{
    DomainGlob:  "*ebay.*",
    Parallelism: 1,        // 1 concurrent request
    Delay:       2 * time.Second,  // 2s between requests
    RandomDelay: 1 * time.Second,  // +0-1s random
})
```

**Anti-Ban Measures**:
1. Rate limiting: Max 1 req/s (constitution constraint)
2. User-Agent rotation: Pool of realistic browser UAs
3. Retry with exponential backoff: 2s, 4s, 8s, 16s (constitution requirement)
4. Respect robots.txt
5. Random delays between requests

**Error Handling**:
```go
c.OnError(func(r *colly.Response, err error) {
    log.Printf("Error scraping %s: %v", r.Request.URL, err)
    // Implement retry logic
})
```

**Alternatives Rejected**:
- chromedp: Not needed for eBay HTML pages
- No rate limiting: Would get banned quickly

---

## 3. Discord Bot Architecture

### Decision: discordgo with slash commands and embeds

**Structure** (from project skill):
```go
var commands = []*discordgo.ApplicationCommand{
    {Name: "analyze", Description: "Lance une analyse de volume eBay"},
    {Name: "results", Description: "Affiche les résultats de la dernière analyse"},
    {Name: "filter", Description: "Filtre les résultats par catégorie"},
}
```

**Deferred Response Pattern** (for long operations):
```go
func handleAnalyze(s *discordgo.Session, i *discordgo.InteractionCreate) {
    // Acknowledge immediately (prevents timeout)
    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
    })

    // Run analysis in background
    go func() {
        results := runAnalysis()

        // Edit the deferred response with results
        s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
            Embeds: &[]*discordgo.MessageEmbed{formatResults(results)},
        })
    }()
}
```

**Embed Format** (ROI First principle):
```go
embed := &discordgo.MessageEmbed{
    Title: "📊 Top Produits Pokemon TCG",
    Color: 0x00FF00,
    Fields: []*discordgo.MessageEmbedField{
        {Name: "Produit", Value: "Display SV 151", Inline: true},
        {Name: "Prix Moyen", Value: "179€", Inline: true},
        {Name: "ROI vs MSRP", Value: "+49%", Inline: true},
        {Name: "Volume (30j)", Value: "127 ventes", Inline: true},
    },
}
```

**Alternatives Rejected**:
- Web interface: Constitution mandates Discord-First UI
- Traditional commands: Slash commands are modern standard

---

## 4. Product Name Normalization

### Decision: Keyword-based normalization with reference dictionary

**Strategy**:
1. **Text normalization**: Lowercase, remove accents, strip special chars
2. **Keyword extraction**: Identify set name + product type
3. **Reference matching**: Match against known product dictionary

**Implementation**:
```go
type ProductNormalizer struct {
    setPatterns    map[string]*regexp.Regexp  // "151", "prismatic", etc.
    productTypes   map[string][]string        // "display" -> ["display", "boite 36", "box"]
    referenceDict  map[string]Product         // Known products with MSRP
}

func (n *ProductNormalizer) Normalize(title string) (Product, float64) {
    normalized := normalizeText(title)  // lowercase, no accents

    set := n.extractSet(normalized)
    productType := n.extractType(normalized)

    if product, confidence := n.matchReference(set, productType); confidence > 0.8 {
        return product, confidence
    }
    return Product{Name: title, Category: productType}, 0.5
}
```

**Pokemon TCG Product Categories**:
| Category | Keywords FR | Keywords EN | MSRP Range |
|----------|-------------|-------------|------------|
| Booster | booster, paquet | booster pack | 4-5€ |
| Display | display, boite 36 | booster box | 140-160€ |
| ETB | coffret dresseur, etb | elite trainer box | 45-55€ |
| Collection | coffret, collection | collection box | 25-50€ |
| Bundle | pack 6 boosters | bundle | 25-30€ |
| Tin | pokebox, tin | tin | 20-25€ |

**Set Name Patterns** (2024-2025):
- Écarlate et Violet (Scarlet & Violet)
- 151
- Destinées de Paldea (Paldean Fates)
- Évolutions Prismatiques (Prismatic Evolutions)
- Masques du Crépuscule (Twilight Masquerade)

**Alternatives Rejected**:
- ML-based clustering: Too complex for MVP, no training data
- Exact string matching: Too brittle for varied eBay titles
- No normalization: Would create duplicate products

---

## 5. Database Schema

### Decision: PostgreSQL with sqlx (from project skill)

**Schema adapted for Volume Analyzer**:

See `data-model.md` for detailed schema.

**Key Design Choices**:
- UUID primary keys (standard for this project)
- `normalized_name` for product grouping
- `msrp_eur` stored with products for ROI calculation
- Indexes on frequently queried columns

**Alternatives Rejected**:
- MongoDB: Project uses PostgreSQL per constitution
- Separate MSRP table: Simpler to embed in products

---

## 6. MSRP Reference Data

### Decision: Static reference table with manual updates

**Rationale**:
- MSRP doesn't change often (new sets released quarterly)
- Manual maintenance is acceptable for MVP
- Can be automated later by scraping Pokemon Center

**Initial Data Sources**:
- Pokemon Center FR official prices
- Major retailers (FNAC, Micromania)

**Structure**:
```go
var MSRPReference = map[string]float64{
    "display_sv_151":       159.99,
    "display_paldean_fates": 159.99,
    "etb_prismatic_evo":    54.99,
    // ...
}
```

---

## Unresolved Items

None - all technical decisions made.

---

## References

- Project Skills: `.claude/skills/go-scraping/SKILL.md`
- Project Skills: `.claude/skills/discord-bot/SKILL.md`
- Project Skills: `.claude/skills/database/SKILL.md`
- Constitution: `.specify/memory/constitution.md`
