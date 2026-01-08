<!--
SYNC IMPACT REPORT
==================
Version Change: 3.1.0 → 4.0.0
Rationale: MAJOR - Refonte complète du projet. Nouveau scope (flip C2C vs restock retail),
nouveaux principes, nouvelle architecture modulaire.

Modified Principles:
- "Data-Driven" → supprimé (remplacé par approche modulaire)
- "Speed Matters" → "Speed First" (même concept, seuil 1min vs 30s)
- "ROI First" → "Low False Positives" (focus qualité alertes)
- "Simplicité" → "Discord-Native" (même esprit, plus précis)
- "Fiabilité" → "Scraping Resilient" (focus anti-bot)
- NEW: "Modular Architecture"

Removed Sections:
- Watchlist (remplacé par scanning automatique)
- Restock Monitor (plus de retail monitoring)
- Spike Detector (cartes singles via autre approche)
- Monétisation (déplacé hors scope MVP)

Added Sections:
- Listing Scanner
- Cross-Platform Arbitrage
- Bundle/Lot Analyzer (avec IA Vision)
- Facebook Groups Monitor
- Grading ROI Calculator
- Plateformes cibles et références prix

Templates Status:
✅ .specify/templates/plan-template.md - compatible (générique)
✅ .specify/templates/spec-template.md - compatible (générique)
✅ .specify/templates/tasks-template.md - compatible (générique)

Follow-up TODOs:
- None
-->

# PokéProfit Constitution

## Mission

Outil SaaS pour détecter les opportunités de flip/revente de cartes Pokémon TCG sur les marketplaces C2C (Vinted, LeBonCoin). L'outil scanne automatiquement les nouvelles annonces, identifie les deals rentables, et alerte l'utilisateur en temps réel.

## Core Principles

### I. Speed First

Les alertes MUST arriver en moins d'1 minute après la publication d'une annonce. La vitesse est l'avantage compétitif principal - premier arrivé, premier servi.

**MUST requirements:**

- Latence maximale de 60 secondes entre publication et alerte (p95)
- Polling agressif des plateformes (dans les limites du rate limiting)
- Architecture optimisée pour la latence (Go, concurrence native)
- Priorisation des annonces récentes dans le pipeline

**MUST NOT:**

- NEVER batching des alertes (envoyer immédiatement)
- NEVER délai artificiel pour "grouper" les notifications
- NEVER sacrifier la vitesse pour des features non-essentielles

**Rationale:** Sur Vinted/LBC, les bons deals partent en minutes. Une alerte en retard = opportunité perdue. La vitesse est THE differentiator.

### II. Low False Positives

Mieux vaut rater une opportunité que d'alerter sur un mauvais deal. Chaque alerte MUST représenter une vraie opportunité de profit.

**MUST requirements:**

- Seuil minimum de 30% de marge brute avant alerte
- Validation croisée des prix avec références (CardMarket, TCGPlayer)
- Scoring de confiance sur chaque deal détecté
- Filtrage des annonces suspectes (scams, fakes, erreurs de catégorie)

**MUST NOT:**

- NEVER alerter sous 30% de marge estimée
- NEVER alerter sans prix de référence validé
- NEVER flood l'utilisateur avec des alertes marginales

**Rationale:** Un utilisateur qui reçoit 50 alertes/jour dont 2 sont bonnes va désactiver l'outil. Qualité > Quantité. Confiance = rétention.

### III. Scraping Resilient

Les scrapers MUST gérer les protections anti-bot et rester opérationnels malgré les obstacles techniques.

**MUST requirements:**

- Retry logic avec backoff exponentiel (1s, 2s, 4s, 8s, 16s, max 60s)
- Rotation de proxies résidentiels pour distribuer les requêtes
- Rotation de User-Agents et fingerprints browser
- Fallback strategies (API officielle si dispo, mobile endpoints, etc.)
- Circuit breaker pour éviter les bans prolongés
- Health monitoring avec alertes si scraper down > 5 minutes

**MUST NOT:**

- NEVER requêtes sans proxy sur sites protégés
- NEVER ignorer les erreurs 429/403 (adapter immédiatement)
- NEVER continuer si détection de captcha sans stratégie

**Rationale:** Vinted et LBC ont des protections Cloudflare/DataDome. Un scraper qui tombe = revenus perdus. La résilience n'est pas optionnelle.

### IV. Modular Architecture

Chaque module MUST être indépendant et activable séparément. Un utilisateur peut choisir exactement les features qu'il veut.

**MUST requirements:**

- Chaque module = package Go isolé avec interface claire
- Configuration par module (enable/disable, seuils, filtres)
- Pas de dépendances croisées entre modules métier
- Un module down ne MUST pas impacter les autres
- Feature flags pour activation granulaire

**MUST NOT:**

- NEVER couplage fort entre modules
- NEVER config globale qui force tous les modules
- NEVER déploiement monolithique obligatoire

**Rationale:** Différents utilisateurs ont différents besoins. Un flipper de lots n'a pas besoin du grading calculator. Modularité = flexibilité = plus de clients satisfaits.

### V. Discord-Native

Toutes les interactions utilisateur MUST passer par Discord. Pas de web UI pour le MVP.

**MUST requirements:**

- Alertes via messages Discord (embed rich avec images)
- Configuration via slash commands (`/config`, `/filters`, `/alerts`)
- Statut système visible via commandes (`/status`, `/health`)
- Support multi-serveurs (un bot, plusieurs guilds)

**MUST NOT:**

- NEVER créer de web dashboard pour le MVP
- NEVER forcer l'utilisateur hors de Discord
- NEVER alertes par email ou SMS

**Rationale:** Les flippers Pokemon sont déjà sur Discord. Zéro friction = meilleure adoption. Web UI = scope creep pour le MVP.

## Scope Fonctionnel

### Plateformes Cibles (Phase 1)

**Scan (sources d'annonces):**
- Vinted FR
- LeBonCoin FR

**Référence Prix:**
- CardMarket (prix marché EU)
- TCGPlayer (prix marché US, conversion)

### Module 1: Listing Scanner (CORE)

**But:** Scanner les nouvelles annonces et détecter les deals via keywords, fautes d'orthographe, et price crashes.

**Stratégies de détection:**

1. **Keyword Matching**
   - Noms de sets (151, Écarlate et Violet, etc.)
   - Noms de cartes populaires (Dracaufeu, Pikachu, etc.)
   - Termes de valeur (PSA, BGS, sealed, display, ETB)

2. **Typo Detection**
   - Fautes courantes (Dracofeu, Pikatchou, etc.)
   - Erreurs de set (152 au lieu de 151)
   - Mauvaise catégorisation (jeux vidéo au lieu de cartes)

3. **Price Crash Detection**
   - Prix < 50% du prix marché → alerte haute priorité
   - Prix < 70% du prix marché → alerte normale

**Output:** Alerte Discord avec lien, prix, marge estimée, confiance.

### Module 2: Cross-Platform Arbitrage

**But:** Détecter les différences de prix entre Vinted et LBC pour le même produit.

**Fonctionnement:**
- Matching de produits similaires entre plateformes
- Calcul de marge nette (prix vente - prix achat - frais)
- Alerte si arbitrage > seuil configuré

**Complexité:** Moyenne (matching produits approximatif)

### Module 3: Bundle/Lot Analyzer

**But:** Analyser les lots de cartes pour estimer leur valeur réelle vs prix demandé.

**Composants:**

1. **Analyse Texte**
   - Extraction des cartes mentionnées dans la description
   - Parsing des listes (quantités, sets, conditions)

2. **Analyse Vision (Claude API)**
   - Upload des photos de lots vers Claude Vision
   - Identification des cartes visibles
   - Estimation de valeur basée sur les cartes détectées

**Output:** Valeur estimée du lot, marge potentielle, liste des cartes identifiées.

### Module 4: Facebook Groups Monitor

**But:** Scanner les groupes Facebook de vente Pokemon pour deals.

**Complexité:** Haute (auth Facebook, scraping difficile)

**Phase:** 2+ (pas MVP)

### Module 5: Grading ROI Calculator

**But:** Calculer si une carte vaut le coût du grading (PSA/CGC).

**Inputs:**
- Prix actuel de la carte raw
- Prix moyen gradé (PSA 9, PSA 10)
- Coût du grading + shipping

**Output:** ROI estimé par grade, recommandation go/no-go.

**Phase:** 2+ (pas MVP)

## Contraintes Techniques

### Stack Imposé

| Composant | Technologie | Justification |
|-----------|-------------|---------------|
| Backend | Go | Performance, concurrence native |
| Scraping (protégé) | Chromedp | Sites avec JS/Cloudflare |
| Scraping (simple) | Colly | Sites HTML statiques |
| Database | PostgreSQL | Données relationnelles |
| Cache/Queue | Redis | Rate limiting, job queue |
| Notifications | discordgo | Bot Discord natif |
| IA Vision | Claude API | Analyse images de lots |

### Contraintes Scraping

**Rate Limits:**
- Vinted: Max 1 req/2s par proxy
- LeBonCoin: Max 1 req/3s par proxy
- CardMarket: Max 1 req/s (API ou scrape)

**Proxy Requirements:**
- Pool minimum: 50 proxies résidentiels
- Rotation: round-robin avec health check
- Géolocalisation: FR prioritaire

**Anti-Detection:**
- User-Agent rotation (pool de 20+ UA récents)
- Headers réalistes (Accept-Language, etc.)
- Delays randomisés (±20% du rate limit)
- Session management (cookies, tokens)

### Contraintes Performance

| Métrique | Cible | Critique |
|----------|-------|----------|
| Latence alerte | < 60s (p95) | < 120s |
| Scan throughput | 1000 annonces/min | 500/min |
| Uptime scrapers | 99% | 95% |
| False positive rate | < 10% | < 20% |

## Interface Discord

### Alertes (Push Automatique)

```
🔥 DEAL DÉTECTÉ - Vinted

📦 Lot 50 cartes Pokemon 151
💰 Prix: 25€
📊 Valeur estimée: 80€+
📈 Marge: +220% (~55€)
🎯 Confiance: 85%

🔗 [Voir l'annonce](lien)
⏰ Publié il y a 45 secondes
```

### Commandes Slash

| Commande | Description |
|----------|-------------|
| `/status` | État des scrapers et stats |
| `/config module <name> <on/off>` | Activer/désactiver un module |
| `/filters set <param> <value>` | Configurer les filtres |
| `/alerts pause <duration>` | Pause temporaire des alertes |
| `/stats [period]` | Statistiques de deals |

## MVP Scope (Phase 1)

**In Scope:**
- Listing Scanner (Vinted + LBC)
- Prix de référence CardMarket
- Alertes Discord basiques
- Configuration minimale via commands

**Out of Scope (Phase 2+):**
- Cross-Platform Arbitrage
- Bundle Analyzer avec Vision
- Facebook Groups Monitor
- Grading ROI Calculator
- Web dashboard
- Monétisation/paiements

## Métriques de Succès

### Phase 1 (MVP)

- [ ] 2 plateformes scannées (Vinted, LBC)
- [ ] Latence < 60s sur 95% des alertes
- [ ] < 10% false positives
- [ ] 5 beta users actifs
- [ ] Au moins 5 deals actionnés avec profit par beta user

### Phase 2

- [ ] Bundle Analyzer opérationnel
- [ ] Cross-Platform Arbitrage actif
- [ ] 20+ beta users
- [ ] Taux de conversion deal→achat > 20%

## Ce que le projet N'EST PAS

**MUST NOT implémenter:**

- Bot d'achat automatique (alertes seulement, décision humaine)
- Marketplace intégré (on détecte, on n'achète/vend pas)
- Gestion d'inventaire personnel
- Prédiction IA des prix futurs
- Scraping de retailers (focus C2C uniquement)
- Support US/international (France only pour MVP)

## Governance

### Amendment Process

1. Proposition documentée avec justification
2. Validation contre les 5 principes
3. Impact assessment sur modules existants
4. Mise à jour constitution + propagation templates
5. Commit avec changelog

### Version Management

- **MAJOR (X.0.0):** Changement de scope, principes, ou suppression de module core
- **MINOR (0.X.0):** Ajout module, nouveau principe, expansion significative
- **PATCH (0.0.X):** Clarifications, corrections, ajustements mineurs

### Compliance

- Toute feature MUST respecter les 5 principes
- PR reviews MUST vérifier: vitesse, qualité alertes, résilience
- Modules MUST être testables indépendamment

**Version**: 4.0.0 | **Ratified**: 2026-01-08 | **Last Amended**: 2026-01-08
