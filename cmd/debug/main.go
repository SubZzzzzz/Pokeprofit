package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/Danny-Dasilva/CycleTLS/cycletls"
	"github.com/PuerkitoBio/goquery"
)

// browserProfile represents a complete browser fingerprint profile
// NOTE: This is duplicated from internal/scraper/ebay/scraper.go for debug purposes
type browserProfile struct {
	UserAgent     string
	SecChUa       string
	Platform      string
	ChromeVersion string
}

// browserProfiles contains realistic browser profiles with matching headers
var browserProfiles = []browserProfile{
	// Chrome 120 - Windows
	{
		UserAgent:     "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		SecChUa:       `"Not_A Brand";v="8", "Chromium";v="120", "Google Chrome";v="120"`,
		Platform:      `"Windows"`,
		ChromeVersion: "120",
	},
	// Chrome 121 - Windows
	{
		UserAgent:     "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36",
		SecChUa:       `"Not A(Brand";v="99", "Google Chrome";v="121", "Chromium";v="121"`,
		Platform:      `"Windows"`,
		ChromeVersion: "121",
	},
	// Chrome 122 - Windows
	{
		UserAgent:     "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
		SecChUa:       `"Chromium";v="122", "Not(A:Brand";v="24", "Google Chrome";v="122"`,
		Platform:      `"Windows"`,
		ChromeVersion: "122",
	},
}

// getChromeHeaders returns realistic Chrome headers based on the profile
func getChromeHeaders(profile browserProfile) map[string]string {
	return map[string]string{
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8",
		"Accept-Language":           "fr-FR,fr;q=0.9,en-US;q=0.8,en;q=0.7",
		"Accept-Encoding":           "gzip, deflate, br",
		"Cache-Control":             "max-age=0",
		"Sec-Ch-Ua":                 profile.SecChUa,
		"Sec-Ch-Ua-Mobile":          "?0",
		"Sec-Ch-Ua-Platform":        profile.Platform,
		"Sec-Fetch-Dest":            "document",
		"Sec-Fetch-Mode":            "navigate",
		"Sec-Fetch-Site":            "none",
		"Sec-Fetch-User":            "?1",
		"Upgrade-Insecure-Requests": "1",
	}
}

func main() {
	// Seed random for profile selection
	rand.Seed(time.Now().UnixNano())

	url := "https://www.ebay.fr/sch/i.html?_nkw=Pokemon+ETB&LH_Complete=1&LH_Sold=1&_sacat=261044"

	// Select random profile
	profile := browserProfiles[rand.Intn(len(browserProfiles))]

	log.Printf("Fetching with CycleTLS: %s", url)
	log.Printf("Using browser profile: Chrome %s on %s", profile.ChromeVersion, profile.Platform)

	client := cycletls.Init()

	// Use Chrome TLS fingerprint with matching headers
	resp, err := client.Do(url, cycletls.Options{
		Body:      "",
		Ja3:       "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-17513,29-23-24,0",
		UserAgent: profile.UserAgent,
		Headers:   getChromeHeaders(profile),
	}, "GET")

	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}

	log.Printf("Status: %d", resp.Status)

	// Save HTML for inspection
	os.WriteFile("debug_output.html", []byte(resp.Body), 0644)
	log.Println("HTML saved to debug_output.html")

	// Parse with goquery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(resp.Body))
	if err != nil {
		log.Fatalf("Parse failed: %v", err)
	}

	// Check title
	title := doc.Find("title").Text()
	log.Printf("Page title: %s", title)

	// Try different selectors
	selectors := []string{
		".s-item",
		"li[data-viewport].s-card",
		"[data-viewport]",
		".srp-river-results li",
	}

	fmt.Println("\n=== Selector Results ===")
	for _, sel := range selectors {
		count := doc.Find(sel).Length()
		fmt.Printf("%-30s : %d elements\n", sel, count)
	}

	// Show first few item titles if found (using new selectors)
	fmt.Println("\n=== First items found (new s-card structure) ===")
	itemCount := 0
	doc.Find("li[data-viewport].s-card").Each(func(i int, s *goquery.Selection) {
		// Get title from new structure
		title := strings.TrimSpace(s.Find(".s-card__title span").Text())

		// Skip placeholder items
		if title == "Shop on eBay" || title == "" {
			return
		}

		// Clean title - remove "La page s'ouvre dans une nouvelle" suffix
		if idx := strings.Index(title, "La page s'ouvre"); idx > 0 {
			title = strings.TrimSpace(title[:idx])
		}

		// Get price from new structure
		price := strings.TrimSpace(s.Find("span.s-card__price").Text())

		// Get sold date
		var soldDate string
		s.Find("span.su-styled-text.positive, span.positive").Each(func(j int, span *goquery.Selection) {
			text := strings.TrimSpace(span.Text())
			if strings.Contains(strings.ToLower(text), "vendu") {
				soldDate = text
			}
		})

		// Get URL
		var url string
		s.Find("a.s-card__link").Each(func(j int, a *goquery.Selection) {
			if href, exists := a.Attr("href"); exists {
				if strings.Contains(href, "/itm/") {
					url = href
					if len(url) > 80 {
						url = url[:80] + "..."
					}
				}
			}
		})

		if itemCount < 10 {
			fmt.Printf("%d. Title: %s\n   Price: %s | Sold: %s\n   URL: %s\n\n",
				itemCount+1, title, price, soldDate, url)
		}
		itemCount++
	})
	fmt.Printf("\nTotal valid items found: %d\n", itemCount)
}
