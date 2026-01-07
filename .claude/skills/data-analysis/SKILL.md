---
name: data-analysis
description: Calculs de ROI, détection de trends, analyse de volumes. Pour identifier opportunités profitables.
allowed-tools: Read, Grep, Glob, Edit, Write, Bash
---

# Data Analysis Patterns (ROI & Opportunities)

## Core Concepts

### 1. Volume Analysis
Identifier les produits avec gros volume de ventes = demande forte

### 2. Arbitrage Detection
Trouver différences de prix entre plateformes

### 3. Spike Detection
Détecter hausses soudaines de prix = hype

### 4. Restock Monitoring
Alerter quand produit rentable revient en stock

## Structures de données

```go
package analyzer

type Opportunity struct {
    Type           string    // "volume", "arbitrage", "spike", "restock"
    ProductID      string
    ProductName    string
    BuyPrice       float64
    SellPrice      float64
    ROI            float64   // En pourcentage
    Confidence     float64   // 0.0 à 1.0
    Volume7d       int
    PlatformBuy    string
    PlatformSell   string
    DetectedAt     time.Time
}

type VolumeStats struct {
    ProductID    string
    SalesCount   int
    AvgPrice     float64
    StdDev       float64
    MinPrice     float64
    MaxPrice     float64
    Trend        string    // "up", "down", "stable"
}

type PricePoint struct {
    Platform   string
    Price      float64
    Timestamp  time.Time
    InStock    bool
}
```

## 1. Volume Analyzer

```go
type VolumeAnalyzer struct {
    salesRepo *SalesRepository
    minSales  int     // Minimum ventes pour considérer (ex: 10)
    minROI    float64 // ROI minimum (ex: 20.0 pour 20%)
}

func (va *VolumeAnalyzer) FindOpportunities(ctx context.Context) ([]Opportunity, error) {
    // 1. Récupérer stats des 7 derniers jours
    stats, err := va.salesRepo.GetVolumeStats(ctx, 7)
    if err != nil {
        return nil, err
    }

    opportunities := []Opportunity{}

    for _, stat := range stats {
        // Filtrer: volume suffisant
        if stat.SalesCount < va.minSales {
            continue
        }

        // Calcul du prix d'achat optimal (percentile 25)
        buyPrice := stat.AvgPrice - (stat.StdDev * 0.5)

        // Prix de vente réaliste (moyenne du marché)
        sellPrice := stat.AvgPrice

        // ROI = (sellPrice - buyPrice - fees) / buyPrice * 100
        fees := calculateFees(sellPrice)
        roi := ((sellPrice - buyPrice - fees) / buyPrice) * 100

        if roi < va.minROI {
            continue
        }

        // Score de confiance basé sur:
        // - Volume (plus = mieux)
        // - Stabilité prix (moins stddev = mieux)
        confidence := calculateConfidence(stat)

        opportunities = append(opportunities, Opportunity{
            Type:        "volume",
            ProductID:   stat.ProductID,
            ProductName: stat.ProductName,
            BuyPrice:    buyPrice,
            SellPrice:   sellPrice,
            ROI:         roi,
            Confidence:  confidence,
            Volume7d:    stat.SalesCount,
            DetectedAt:  time.Now(),
        })
    }

    // Trier par ROI * Confidence
    sort.Slice(opportunities, func(i, j int) bool {
        scoreI := opportunities[i].ROI * opportunities[i].Confidence
        scoreJ := opportunities[j].ROI * opportunities[j].Confidence
        return scoreI > scoreJ
    })

    return opportunities, nil
}

func calculateFees(sellPrice float64) float64 {
    // eBay: 12.9% + 0.30€
    ebayFee := sellPrice*0.129 + 0.30

    // PayPal: 2.9% + 0.35€
    paypalFee := sellPrice*0.029 + 0.35

    // Shipping estimé
    shipping := 2.0

    return ebayFee + paypalFee + shipping
}

func calculateConfidence(stat VolumeStats) float64 {
    // Volume score (0-1): normalisé sur max 100 ventes
    volumeScore := math.Min(float64(stat.SalesCount)/100.0, 1.0)

    // Stability score: inverse du coefficient de variation
    cv := stat.StdDev / stat.AvgPrice
    stabilityScore := 1.0 - math.Min(cv, 1.0)

    // Moyenne pondérée
    return (volumeScore * 0.6) + (stabilityScore * 0.4)
}
```

## 2. Arbitrage Finder

```go
type ArbitrageFinder struct {
    priceRepo *PriceRepository
    minSpread float64 // Spread minimum (ex: 5.0€)
}

func (af *ArbitrageFinder) FindArbitrage(ctx context.Context) ([]Opportunity, error) {
    // Récupérer prix actuels sur toutes plateformes
    prices, err := af.priceRepo.GetLatestPrices(ctx)
    if err != nil {
        return nil, err
    }

    // Grouper par produit
    byProduct := groupByProduct(prices)

    opportunities := []Opportunity{}

    for productID, pricePoints := range byProduct {
        // Trouver min et max
        minPrice := findMin(pricePoints)
        maxPrice := findMax(pricePoints)

        spread := maxPrice.Price - minPrice.Price
        fees := calculateFees(maxPrice.Price)
        profit := spread - fees

        if profit < af.minSpread {
            continue
        }

        roi := (profit / minPrice.Price) * 100

        opportunities = append(opportunities, Opportunity{
            Type:         "arbitrage",
            ProductID:    productID,
            BuyPrice:     minPrice.Price,
            SellPrice:    maxPrice.Price,
            ROI:          roi,
            PlatformBuy:  minPrice.Platform,
            PlatformSell: maxPrice.Platform,
            Confidence:   0.8, // Arbitrage = assez safe
            DetectedAt:   time.Now(),
        })
    }

    return opportunities, nil
}
```

## 3. Spike Detector

```go
type SpikeDetector struct {
    priceRepo     *PriceRepository
    lookbackDays  int     // Période de référence (ex: 30 jours)
    spikeThreshold float64 // % hausse pour qualifier (ex: 50.0)
}

func (sd *SpikeDetector) DetectSpikes(ctx context.Context) ([]Opportunity, error) {
    // Pour chaque produit, comparer prix actuel vs historique
    products, err := sd.priceRepo.GetActiveProducts(ctx)
    if err != nil {
        return nil, err
    }

    opportunities := []Opportunity{}

    for _, product := range products {
        // Prix moyen historique
        avgHistorical, err := sd.priceRepo.GetAvgPrice(ctx, product.ID, sd.lookbackDays)
        if err != nil {
            continue
        }

        // Prix actuel
        currentPrice, err := sd.priceRepo.GetCurrentPrice(ctx, product.ID)
        if err != nil {
            continue
        }

        // Calcul du spike
        increase := ((currentPrice - avgHistorical) / avgHistorical) * 100

        if increase < sd.spikeThreshold {
            continue
        }

        opportunities = append(opportunities, Opportunity{
            Type:        "spike",
            ProductID:   product.ID,
            ProductName: product.Name,
            BuyPrice:    avgHistorical,
            SellPrice:   currentPrice,
            ROI:         increase,
            Confidence:  0.6, // Spikes = plus risqué
            DetectedAt:  time.Now(),
        })
    }

    return opportunities, nil
}
```

## 4. Scoring Global

```go
func ScoreOpportunity(opp Opportunity) float64 {
    // Facteurs:
    // 1. ROI (poids 40%)
    roiScore := math.Min(opp.ROI/100.0, 1.0) * 0.4

    // 2. Confidence (poids 30%)
    confidenceScore := opp.Confidence * 0.3

    // 3. Volume (poids 30%)
    volumeScore := math.Min(float64(opp.Volume7d)/50.0, 1.0) * 0.3

    return roiScore + confidenceScore + volumeScore
}
```

## Best Practices

1. **Fenêtres glissantes** : Toujours analyser sur période récente (7-30j)
2. **Outlier removal** : Enlever ventes anormales (prix trop bas/haut)
3. **Seasonal adjustment** : Tenir compte de saisonnalité (ex: Noël)
4. **Backtesting** : Valider algorithmes sur données historiques
5. **A/B testing** : Tester différents seuils et paramètres
