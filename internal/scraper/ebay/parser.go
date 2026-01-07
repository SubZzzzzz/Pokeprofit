package ebay

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Parser handles parsing of eBay HTML pages.
type Parser struct {
	priceRegex *regexp.Regexp
	dateRegex  *regexp.Regexp
}

// NewParser creates a new eBay HTML parser.
func NewParser() *Parser {
	return &Parser{
		priceRegex: regexp.MustCompile(`(\d+[.,]?\d*)\s*(?:EUR|€)`),
		dateRegex:  regexp.MustCompile(`(\d{1,2})\s+(janv?\.?|févr?\.?|mars|avr\.?|mai|juin|juil\.?|août|sept\.?|oct\.?|nov\.?|déc\.?)\s*(\d{4})?`),
	}
}

// ParseSearchResults parses an eBay search results page and returns raw sales.
func (p *Parser) ParseSearchResults(doc *goquery.Document) ([]RawSale, error) {
	var sales []RawSale

	doc.Find(".s-item").Each(func(i int, s *goquery.Selection) {
		// Skip the first item which is often a placeholder
		if i == 0 && s.Find(".s-item__title").Text() == "Shop on eBay" {
			return
		}

		sale, err := p.parseItem(s)
		if err != nil {
			return // Skip items that fail to parse
		}

		if sale.Title != "" && sale.Price > 0 {
			sales = append(sales, sale)
		}
	})

	return sales, nil
}

func (p *Parser) parseItem(s *goquery.Selection) (RawSale, error) {
	sale := RawSale{
		Platform: "ebay",
		Currency: "EUR",
		Metadata: make(map[string]string),
	}

	// Parse title
	titleSel := s.Find(".s-item__title")
	sale.Title = strings.TrimSpace(titleSel.Text())

	// Parse price
	priceSel := s.Find(".s-item__price")
	priceText := priceSel.Text()
	sale.Price = p.parsePrice(priceText)

	// Parse URL
	linkSel := s.Find(".s-item__link")
	if href, exists := linkSel.Attr("href"); exists {
		// Clean the URL (remove tracking parameters)
		sale.URL = p.cleanURL(href)
	}

	// Parse sold date
	soldDateSel := s.Find(".s-item__title--tag, .POSITIVE")
	soldDateText := soldDateSel.Text()
	sale.SoldAt = p.parseSoldDate(soldDateText)

	// If no sold date found, use current time as fallback
	if sale.SoldAt.IsZero() {
		sale.SoldAt = time.Now()
	}

	// Parse additional metadata
	if condition := s.Find(".SECONDARY_INFO").Text(); condition != "" {
		sale.Metadata["condition"] = strings.TrimSpace(condition)
	}

	if shipping := s.Find(".s-item__shipping").Text(); shipping != "" {
		sale.Metadata["shipping"] = strings.TrimSpace(shipping)
	}

	return sale, nil
}

func (p *Parser) parsePrice(priceText string) float64 {
	// Remove spaces and normalize
	priceText = strings.TrimSpace(priceText)
	priceText = strings.ReplaceAll(priceText, "\u00a0", " ")

	// Extract price using regex
	matches := p.priceRegex.FindStringSubmatch(priceText)
	if len(matches) < 2 {
		// Try alternate format
		priceText = strings.ReplaceAll(priceText, "€", "")
		priceText = strings.ReplaceAll(priceText, "EUR", "")
		priceText = strings.TrimSpace(priceText)
	} else {
		priceText = matches[1]
	}

	// Normalize decimal separator
	priceText = strings.ReplaceAll(priceText, ",", ".")
	priceText = strings.ReplaceAll(priceText, " ", "")

	// Handle range prices (take the first/lower price)
	if strings.Contains(priceText, "à") {
		parts := strings.Split(priceText, "à")
		priceText = strings.TrimSpace(parts[0])
	}

	price, _ := strconv.ParseFloat(priceText, 64)
	return price
}

func (p *Parser) parseSoldDate(dateText string) time.Time {
	dateText = strings.ToLower(strings.TrimSpace(dateText))

	// Check for "Vendu le" or "Sold" prefix
	dateText = strings.TrimPrefix(dateText, "vendu le ")
	dateText = strings.TrimPrefix(dateText, "sold ")

	matches := p.dateRegex.FindStringSubmatch(dateText)
	if len(matches) < 3 {
		return time.Time{}
	}

	day, _ := strconv.Atoi(matches[1])
	month := p.parseMonth(matches[2])
	year := time.Now().Year()

	if len(matches) >= 4 && matches[3] != "" {
		year, _ = strconv.Atoi(matches[3])
	}

	// If the parsed date is in the future, assume last year
	parsedDate := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
	if parsedDate.After(time.Now()) {
		parsedDate = parsedDate.AddDate(-1, 0, 0)
	}

	return parsedDate
}

func (p *Parser) parseMonth(monthStr string) time.Month {
	monthStr = strings.ToLower(strings.TrimSuffix(monthStr, "."))

	monthMap := map[string]time.Month{
		"janv":  time.January,
		"jan":   time.January,
		"févr":  time.February,
		"fev":   time.February,
		"mars":  time.March,
		"avr":   time.April,
		"mai":   time.May,
		"juin":  time.June,
		"juil":  time.July,
		"août":  time.August,
		"aout":  time.August,
		"sept":  time.September,
		"oct":   time.October,
		"nov":   time.November,
		"déc":   time.December,
		"dec":   time.December,
	}

	if month, ok := monthMap[monthStr]; ok {
		return month
	}

	return time.January
}

func (p *Parser) cleanURL(url string) string {
	// Remove tracking parameters
	if idx := strings.Index(url, "?"); idx != -1 {
		baseURL := url[:idx]
		// Keep only the item ID part
		return baseURL
	}
	return url
}

// HasNextPage checks if there's a next page link.
func (p *Parser) HasNextPage(doc *goquery.Document) bool {
	return doc.Find(".pagination__next").Length() > 0
}

// GetNextPageURL extracts the next page URL.
func (p *Parser) GetNextPageURL(doc *goquery.Document) string {
	nextLink := doc.Find(".pagination__next")
	if href, exists := nextLink.Attr("href"); exists {
		return href
	}
	return ""
}

// GetResultCount extracts the total result count from the page.
func (p *Parser) GetResultCount(doc *goquery.Document) int {
	countText := doc.Find(".srp-controls__count-heading").Text()

	// Extract number from text like "123 résultats"
	numRegex := regexp.MustCompile(`(\d+(?:\s*\d+)*)\s*résultats?`)
	matches := numRegex.FindStringSubmatch(countText)
	if len(matches) >= 2 {
		numStr := strings.ReplaceAll(matches[1], " ", "")
		count, _ := strconv.Atoi(numStr)
		return count
	}

	return 0
}
