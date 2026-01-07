# Discord Commands Contract: Volume Analyzer

**Date**: 2026-01-07
**Feature Branch**: `001-ebay-volume-analyzer`

## Overview

This document defines the Discord slash commands for the Volume Analyzer feature.

---

## Commands

### /analyze

Launches a new eBay volume analysis.

**Definition**:
```yaml
name: analyze
description: Lance une analyse de volume des ventes Pokemon TCG sur eBay
options:
  - name: query
    type: STRING
    description: Terme de recherche (ex: "Pokemon Display 151")
    required: false
  - name: category
    type: STRING
    description: Catégorie de produit à analyser
    required: false
    choices:
      - name: Tous
        value: all
      - name: Displays
        value: display
      - name: ETB
        value: etb
      - name: Coffrets
        value: collection
      - name: Boosters
        value: booster
```

**Response Flow**:
1. **Immediate** (deferred): "🔄 Analyse en cours... Cela peut prendre quelques minutes."
2. **On Progress** (ephemeral edit): "🔄 Analyse: {x} ventes trouvées sur {y} pages..."
3. **On Complete** (final edit): Embed with results summary

**Success Response**:
```yaml
embed:
  title: "✅ Analyse Terminée"
  color: 0x00FF00  # Green
  fields:
    - name: "📊 Résultats"
      value: "{products_count} produits analysés"
      inline: true
    - name: "💰 Ventes"
      value: "{sales_count} ventes sur 30j"
      inline: true
    - name: "⏱️ Durée"
      value: "{duration}"
      inline: true
  footer:
    text: "Utilisez /results pour voir le classement"
```

**Error Response**:
```yaml
embed:
  title: "❌ Erreur d'Analyse"
  color: 0xFF0000  # Red
  description: "{error_message}"
  fields:
    - name: "Cause possible"
      value: "{suggestion}"
```

---

### /results

Displays the analysis results sorted by profitability.

**Definition**:
```yaml
name: results
description: Affiche les résultats de la dernière analyse de volume
options:
  - name: sort
    type: STRING
    description: Critère de tri
    required: false
    choices:
      - name: Marge (%) - Recommandé
        value: margin_percent
      - name: Volume de ventes
        value: sales_count
      - name: Prix moyen
        value: avg_price
  - name: limit
    type: INTEGER
    description: Nombre de résultats (max 25)
    required: false
    min_value: 1
    max_value: 25
```

**Response**:
```yaml
embed:
  title: "📊 Top Produits Pokemon TCG"
  color: 0x3498DB  # Blue
  description: "Classement par {sort_criteria} (30 derniers jours)"
  fields:
    # Repeated for each product (max 10 inline fields)
    - name: "🥇 {product_name}"
      value: |
        💰 Prix: {avg_price}€ (MSRP: {msrp}€)
        📈 ROI: +{margin_percent}%
        📦 Volume: {sales_count} ventes
      inline: false
    # ... more products
  footer:
    text: "Dernière mise à jour: {last_analysis_date}"
```

**Pagination** (if > 10 results):
```yaml
components:
  - type: ACTION_ROW
    components:
      - type: BUTTON
        style: PRIMARY
        label: "◀️ Précédent"
        custom_id: "results_prev_{page}"
        disabled: {is_first_page}
      - type: BUTTON
        style: SECONDARY
        label: "Page {current}/{total}"
        custom_id: "results_page"
        disabled: true
      - type: BUTTON
        style: PRIMARY
        label: "Suivant ▶️"
        custom_id: "results_next_{page}"
        disabled: {is_last_page}
```

---

### /filter

Filters the results by product category.

**Definition**:
```yaml
name: filter
description: Filtre les résultats par catégorie de produit
options:
  - name: category
    type: STRING
    description: Catégorie à afficher
    required: true
    choices:
      - name: Displays (Boîtes 36 boosters)
        value: display
      - name: ETB (Elite Trainer Box)
        value: etb
      - name: Coffrets / Collections
        value: collection
      - name: Boosters individuels
        value: booster
      - name: Bundles (6 boosters)
        value: bundle
      - name: Tins / Pokebox
        value: tin
```

**Response**:
Same format as `/results` but filtered to the selected category.

```yaml
embed:
  title: "📊 Top {category_name}"
  color: 0x9B59B6  # Purple
  # ... same fields structure as /results
```

---

## Error Handling

### Common Error Messages

| Error Code | User Message | Suggestion |
|------------|--------------|------------|
| `NO_DATA` | "Aucune donnée disponible" | "Lancez d'abord une analyse avec /analyze" |
| `SCRAPE_FAILED` | "Impossible de collecter les données" | "eBay peut être temporairement inaccessible. Réessayez dans quelques minutes." |
| `RATE_LIMITED` | "Trop de requêtes" | "Attendez quelques secondes avant de réessayer." |
| `ANALYSIS_RUNNING` | "Une analyse est déjà en cours" | "Attendez la fin de l'analyse actuelle." |

---

## Embed Color Codes

| Status | Color | Hex |
|--------|-------|-----|
| Success | Green | `0x00FF00` |
| Error | Red | `0xFF0000` |
| Info/Results | Blue | `0x3498DB` |
| Filtered | Purple | `0x9B59B6` |
| Warning | Orange | `0xFFA500` |
| In Progress | Yellow | `0xFFFF00` |

---

## Rate Limits

- `/analyze`: 1 per user per 5 minutes (long-running operation)
- `/results`: 5 per user per minute
- `/filter`: 10 per user per minute

---

## Permissions

All commands require:
- User must be in the guild where the bot is installed
- No specific role requirements for MVP (all users can use)

Future (paid tiers):
- Free tier: Limited to 3 /analyze per day
- Pro tier: Unlimited /analyze
