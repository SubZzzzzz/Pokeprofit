# Pokemon TCG Profit Tool

## Description
Outil SaaS pour identifier et profiter des opportunités de revente Pokemon TCG.

## Stack
- Go (backend, scrapers, bot)
- PostgreSQL (database)
- Redis (cache)
- Discord (notifications)

## Modules
1. Volume Analyzer - Scrape eBay/Vinted pour trouver produits rentables
2. Restock Monitor - Alertes stock retailers
3. Arbitrage Finder - Différences de prix cross-platform
4. Spike Detector - Hausses de prix cartes

## Conventions
- Fichiers Go: snake_case
- Packages: lowercase
- Errors: wrap avec fmt.Errorf
- Tests: _test.go dans le même package
