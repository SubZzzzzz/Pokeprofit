# Implementation Plan: Volume Analyzer Phase 1 - eBay

**Branch**: `001-ebay-volume-analyzer` | **Date**: 2026-01-07 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-ebay-volume-analyzer/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Système d'analyse de volume de ventes Pokemon TCG sur eBay FR pour identifier les produits rentables. Le système scrape les ventes complétées eBay, calcule les statistiques (volume, prix moyen, marge vs MSRP) et expose les résultats via un bot Discord avec commandes slash (/analyze, /results, /filter).

## Technical Context

**Language/Version**: Go 1.21+
**Primary Dependencies**: colly (scraping HTML), discordgo (bot Discord), sqlx (database), go-redis (cache)
**Storage**: PostgreSQL (données de ventes, produits, analyses), Redis (cache, rate limiting)
**Testing**: go test avec table-driven tests
**Target Platform**: Linux server (VPS)
**Project Type**: single (backend monolithique avec bot Discord)
**Performance Goals**: Analyse de 100+ produits en <10 minutes, refresh quotidien des données
**Constraints**: Rate limiting eBay (1 req/s max), <512MB mémoire base
**Scale/Scope**: MVP pour 5-10 beta users, 100+ produits analysés par session

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Principle I: Data-Driven ✅
- **Compliance**: Le Volume Analyzer est fondé sur les données de ventes complétées eBay, pas des suppositions
- **Validation**: Calcul de volume, prix moyen, et marge basé sur données mesurables

### Principle II: Speed Matters ⚠️ (N/A pour Phase 1)
- **Note**: Ce principe concerne principalement les alertes (Phase 2+). Pour Phase 1, le refresh quotidien est suffisant
- **Future**: L'architecture doit permettre d'évoluer vers du temps réel

### Principle III: ROI First ✅
- **Compliance**: Chaque produit affiche son ROI potentiel (prix vente - MSRP)
- **Validation**: Tri par rentabilité, calcul de marge automatique

### Principle IV: Simplicité ✅
- **Compliance**: Commandes Discord simples (/analyze, /results, /filter)
- **Validation**: Données essentielles: nom, prix moyen, volume, marge - pas de dashboards complexes

### Principle V: Fiabilité ✅
- **Compliance**: Rate limiting intégré, retry logic, logs détaillés
- **Validation**: Gestion d'erreurs avec messages utilisateur clairs

### Discord-First UI ✅
- **Compliance**: Interface unique via bot Discord avec commandes slash
- **Validation**: Embeds riches pour présentation des résultats

### Technical Constraints ✅
- **Stack**: Go + PostgreSQL + Redis + colly + discordgo (conforme)
- **Rate Limiting**: 1 req/s max par domaine (implémenté via colly)
- **Performance**: 100 produits en <10 min (conforme objectif SC-002)

## Project Structure

### Documentation (this feature)

```text
specs/001-ebay-volume-analyzer/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
cmd/
├── bot/                 # Point d'entrée du bot Discord
│   └── main.go
└── analyzer/            # CLI pour lancer l'analyse manuellement (debug/cron)
    └── main.go

internal/
├── config/              # Configuration (env vars, settings)
│   └── config.go
├── database/            # Connexion et migrations PostgreSQL
│   ├── postgres.go
│   └── migrations/
├── cache/               # Client Redis
│   └── redis.go
├── scraper/             # Scrapers eBay
│   ├── ebay/
│   │   ├── scraper.go   # Scraper principal
│   │   ├── parser.go    # Parsing HTML des ventes
│   │   └── types.go     # Types spécifiques eBay
│   └── common/          # Utils partagés (rate limiting, retry)
│       ├── ratelimit.go
│       └── retry.go
├── analyzer/            # Logique d'analyse et calculs
│   ├── volume.go        # Calcul volume de ventes
│   ├── profit.go        # Calcul marge/ROI
│   └── normalizer.go    # Normalisation noms produits
├── repository/          # Accès données (patterns repository)
│   ├── product.go
│   ├── sale.go
│   └── analysis.go
├── discord/             # Bot Discord
│   ├── bot.go           # Setup et handlers
│   ├── commands/        # Slash commands
│   │   ├── analyze.go
│   │   ├── results.go
│   │   └── filter.go
│   └── embeds/          # Builders d'embeds Discord
│       └── results.go
└── models/              # Entités du domaine
    ├── product.go
    ├── sale.go
    ├── analysis.go
    └── msrp.go

pkg/                     # Code réutilisable (si besoin futur)
└── ...

tests/
├── integration/         # Tests d'intégration (DB, scraper mock)
└── mocks/               # Mocks pour tests unitaires
```

**Structure Decision**: Structure Go standard avec séparation `cmd/` (points d'entrée) et `internal/` (code privé). Pattern repository pour l'accès données, facilitant les tests et l'évolution future.

## Complexity Tracking

> **No violations - design is compliant with all Constitution principles.**

| Aspect | Decision | Rationale |
|--------|----------|-----------|
| Repository pattern | Used | Facilitates testing with mocks, aligns with Go best practices |
| Materialized view | Used for stats | Performance optimization for frequent reads, acceptable complexity |
| Single binary | Yes | Simpler deployment, all components in one process for MVP |

---

## Post-Design Constitution Check ✅

*Re-evaluated after Phase 1 design completion.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Data-Driven | ✅ Pass | All data from real eBay completed sales |
| II. Speed Matters | ⚠️ N/A | Phase 1 focus is batch analysis, not real-time alerts |
| III. ROI First | ✅ Pass | Margin calculation in every product result |
| IV. Simplicité | ✅ Pass | 3 simple Discord commands, no web UI |
| V. Fiabilité | ✅ Pass | Rate limiting, retries, error handling defined |
| Discord-First UI | ✅ Pass | Slash commands with embeds only |
| Technical Stack | ✅ Pass | Go, PostgreSQL, Redis, colly, discordgo |

**Gate Status**: PASSED - Ready for task generation.

---

## Generated Artifacts

| Artifact | Status | Path |
|----------|--------|------|
| plan.md | ✅ Complete | `specs/001-ebay-volume-analyzer/plan.md` |
| research.md | ✅ Complete | `specs/001-ebay-volume-analyzer/research.md` |
| data-model.md | ✅ Complete | `specs/001-ebay-volume-analyzer/data-model.md` |
| quickstart.md | ✅ Complete | `specs/001-ebay-volume-analyzer/quickstart.md` |
| discord-commands.md | ✅ Complete | `specs/001-ebay-volume-analyzer/contracts/discord-commands.md` |
| scraper-interface.md | ✅ Complete | `specs/001-ebay-volume-analyzer/contracts/scraper-interface.md` |

---

## Next Step

Run `/speckit.tasks` to generate the implementation task list.
