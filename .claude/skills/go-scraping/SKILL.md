---
name: go-scraping
description: Patterns et best practices pour scraper en Go avec colly et chromedp. Utiliser pour créer des scrapers eBay, Vinted, CardMarket, ou tout site e-commerce.
allowed-tools: Read, Grep, Glob, Edit, Write, Bash
---

# Go Scraping Patterns

## Stack
- **colly** : Sites statiques (HTML simple)
- **chromedp/rod** : Sites JS-heavy (React, Vue, etc.)
- **goquery** : Parsing HTML

## Structure d'un scraper

```go
package scraper

type Scraper interface {
    Scrape(ctx context.Context) ([]Sale, error)
    Name() string
}

type Sale struct {
    Platform    string
    Title       string
    Price       float64
    SoldAt      time.Time
    URL         string
}
```

## Pattern Colly (sites simples)

```go
func NewEbayScraper() *EbayScraper {
    c := colly.NewCollector(
        colly.AllowedDomains("www.ebay.fr"),
        colly.UserAgent("Mozilla/5.0..."),
    )

    c.Limit(&colly.LimitRule{
        DomainGlob:  "*ebay.*",
        Parallelism: 2,
        Delay:       2 * time.Second,
    })

    return &EbayScraper{collector: c}
}
```

## Pattern Chromedp (sites JS)

```go
func scrapeDynamic(url string) ([]Sale, error) {
    ctx, cancel := chromedp.NewContext(context.Background())
    defer cancel()

    var html string
    err := chromedp.Run(ctx,
        chromedp.Navigate(url),
        chromedp.WaitVisible(".product-list"),
        chromedp.OuterHTML("body", &html),
    )
    return parseHTML(html)
}
```

## Anti-ban essentiels

- Délai entre requêtes (2-5s)
- Rotation User-Agent
- Proxies rotatifs
- Retry avec backoff exponentiel
- Respecter robots.txt
