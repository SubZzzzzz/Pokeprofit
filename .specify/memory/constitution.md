<!--
SYNC IMPACT REPORT
==================
Version Change: 3.0.0 → 3.1.0
Rationale: MINOR - Ajout du prix de revente estimé dans la watchlist
pour calcul ROI dans les alertes restock.

Modified Sections:
- Watchlist: ajout champ "prix revente estimé" par produit
- Alertes restock: affichent maintenant le profit estimé

Added Features:
- Prix revente configurable par produit dans watchlist
- Calcul ROI automatique dans alertes (retail vs revente)

Templates Status:
✅ .specify/templates/plan-template.md - compatible
✅ .specify/templates/spec-template.md - compatible
✅ .specify/templates/tasks-template.md - compatible

Follow-up TODOs:
- None
-->

# PokéProfit Constitution

## Core Principles

### I. Data-Driven

Toutes les décisions sont basées sur des données réelles. L'utilisateur identifie les produits rentables via sa propre recherche, l'outil surveille leur disponibilité.

**MUST requirements:**

- Surveiller les retailers pour détecter les restocks en temps réel
- Fournir des données précises (prix, disponibilité, lien direct)
- Permettre à l'utilisateur de gérer sa watchlist de produits

**Rationale:** Le matching automatique entre listings marketplace est trop complexe et peu fiable. L'utilisateur connaît mieux que quiconque les produits qu'il veut surveiller. On se concentre sur ce qu'on fait bien : détecter les restocks rapidement.

### II. Speed Matters

Dans le reselling, la vitesse est critique. Les alertes doivent arriver en secondes, pas en minutes. Premier arrivé = premier servi.

**MUST requirements:**

- Alertes envoyées en moins de 30 secondes après détection d'un restock
- Architecture conçue pour la performance (Go, concurrence native)
- Monitoring en temps réel des retailers (pas de polling lent)
- Support de 1000+ produits monitorés simultanément

**Rationale:** Les produits Pokemon à forte demande se vendent en minutes. Une alerte en retard = opportunité perdue = argent perdu pour l'utilisateur.

### III. ROI First

Chaque feature doit aider l'utilisateur à gagner de l'argent. Si une feature n'améliore pas le ROI, elle n'a pas sa place.

**MUST requirements:**

- Prioriser les features qui augmentent directement le profit utilisateur
- Rejeter les features "nice-to-have" qui ne contribuent pas au ROI
- Mesurer le succès en valeur apportée à l'utilisateur

**MUST NOT:**

- NEVER implémenter des features purement esthétiques
- NEVER créer des dashboards complexes sans valeur actionnable
- NEVER ajouter de la complexité qui n'améliore pas les profits

**Rationale:** Les utilisateurs paient pour gagner de l'argent, pas pour des interfaces jolies. Chaque euro de développement doit générer des euros de profit utilisateur.

### IV. Simplicité

L'utilisateur veut des alertes claires et actionnables. Pas de complexité inutile.

**MUST requirements:**

- Messages Discord concis avec les informations essentielles
- Commandes simples et intuitives (`/watch`, `/unwatch`, `/alerts`)
- Watchlist facile à gérer
- Pas de configuration complexe requise

**MUST NOT:**

- NEVER créer des interfaces nécessitant une formation
- NEVER cacher l'information essentielle derrière des clics multiples
- NEVER utiliser du jargon technique face à l'utilisateur

**Rationale:** L'utilisateur est un revendeur occupé. Il a besoin d'alertes claires pour agir rapidement.

### V. Fiabilité

Les scrapers doivent être robustes. Une alerte manquée = argent perdu pour l'utilisateur = perte de confiance.

**MUST requirements:**

- Scrapers avec retry logic et backoff exponentiel
- Proxies rotatifs pour éviter les bans
- Logs détaillés pour debug et monitoring
- Alertes de santé système (scraper down, API error, etc.)
- Tests d'intégration pour valider les scrapers régulièrement

**MUST NOT:**

- NEVER déployer un scraper sans tests de robustesse
- NEVER ignorer les erreurs silencieusement
- NEVER laisser un scraper cassé sans alerte système

**Rationale:** La fiabilité est la base de la confiance. Si l'outil rate des opportunités, l'utilisateur le désinstalle. Un système fiable = utilisateurs qui restent et paient.

## Scope Fonctionnel

### Watchlist (Source de données)

**But:** Liste des produits à surveiller, maintenue par l'utilisateur avec prix de revente estimé

**Fonctionnement:**

- L'utilisateur ajoute/retire des produits via commandes Discord
- Chaque produit = nom + URLs retailers + prix de revente estimé
- Le prix de revente permet de calculer le ROI dans les alertes
- L'utilisateur identifie les produits rentables via sa propre recherche (eBay sold, groupes Discord, expérience)
- L'IA peut aider à la recherche mais la décision reste humaine

**Données par produit:**

- Nom du produit
- URLs des retailers à surveiller
- Prix de revente estimé (défini par l'utilisateur)
- Date d'ajout

**Commandes:**

- `/watch [nom] [prix_revente] [url1] [url2]...` - Ajouter un produit avec prix revente estimé
- `/unwatch [nom]` - Retirer un produit
- `/watchlist` - Voir sa liste de produits surveillés avec prix revente
- `/setprice [nom] [prix]` - Modifier le prix revente d'un produit existant

**Principe:** L'utilisateur sait ce qui est rentable. L'outil surveille et calcule le ROI, l'humain décide.

### Module 1: Restock Monitor (CORE)

**But:** Alerter quand les produits de la watchlist sont disponibles, avec calcul du profit estimé

**Composants:**

- Scrapers pour retailers FR: Pokemon Center, FNAC, Micromania, Amazon, Cultura
- Système d'alertes Discord avec: lien direct, prix retail, prix revente, profit estimé, stock disponible
- Calcul automatique du ROI basé sur le prix revente de la watchlist
- Polling intelligent avec détection de changements

**Données dans l'alerte:**

- Nom du produit
- Prix retail (scrappé du retailer)
- Prix revente estimé (de la watchlist)
- Profit estimé en € et en %
- Lien direct vers le produit
- Indicateur de stock (limité/disponible)

**Principe:** Alerte = Décision immédiate possible (toutes les infos ROI présentes)

### Module 2: Arbitrage Finder

**But:** Détecter les différences de prix entre plateformes

**Composants:**

- Comparateur de prix: CardMarket vs eBay vs Vinted
- Calculateur de profit net après frais (commissions, shipping)
- Alertes quand opportunité > seuil défini (configurable par tier)

**Principe:** Arbitrage = profit quasi sans risque si bien exécuté

### Module 3: Spike Detector

**But:** Détecter les hausses de prix anormales sur les cartes (singles)

**Composants:**

- Tracker de prix CardMarket pour cartes populaires
- Détecteur de variations > X% en Y heures
- Système d'alerte avec contexte (cause probable du spike)

**Principe:** Information = pouvoir (vendre avant les autres, ou acheter avant que ça monte)

### Module 4: Monétisation

**But:** Générer des revenus récurrents

**Composants:**

- Système de tiers: Free (limité), Pro (15€/mois), Business (35€/mois)
- Intégration Stripe pour paiements
- Feature gating par tier
- Gestion des abonnements et renouvellements

**Tiers:**

- **Free:** 5 produits max dans watchlist, 3 alertes/jour
- **Pro (15€/mois):** Watchlist illimitée, alertes illimitées, alertes prioritaires
- **Business (35€/mois):** Tout Pro + Arbitrage Finder + Spike Detector

## Contraintes Techniques

### Stack Imposé

- **Backend:** Go (Golang) - performance et concurrence native pour scrapers
- **Database:** PostgreSQL - données relationnelles (produits, watchlists, utilisateurs)
- **Cache:** Redis - sessions utilisateur, rate limiting, cache de données fréquentes
- **Bot:** Discord via discordgo library
- **Scraping:** colly (sites HTML statiques), chromedp/rod (sites JavaScript)

**Justification:** Go offre les performances nécessaires pour monitorer 1000+ produits avec latence < 30s. PostgreSQL + Redis assurent fiabilité et rapidité.

### Contraintes Scraping

**MUST requirements:**

- Respecter les rate limits pour éviter les bans (1 requête/seconde max par retailer)
- Utiliser des proxies rotatifs pour distribuer la charge
- Implémenter retry logic avec backoff exponentiel (2s, 4s, 8s, 16s)
- Logs détaillés pour debug (timestamp, URL, status code, erreur)
- User-agents rotatifs et headers réalistes

**MUST NOT:**

- NEVER faire plus de 1 req/s par domaine
- NEVER ignorer les robots.txt
- NEVER scraper sans retry logic

### Contraintes Performance

**MUST requirements:**

- Alertes envoyées < 30 secondes après détection
- Support 1000+ produits monitorés simultanément
- API Discord répondant en < 500ms
- Database queries optimisées (indexes, no N+1)

**Benchmarks:**

- Latency p95 < 30s pour alertes restock
- Throughput: 1000 produits scannés en < 5 minutes
- Memory usage < 512MB (base) + 1MB per 100 produits

## Interface Utilisateur (Discord-First)

Discord est l'interface utilisateur principale et UNIQUE du projet. Pas de web app, pas de mobile app - tout passe par Discord.

### Deux modes d'interaction

**1. Notifications Automatiques (Push)**

Les monitors tournent en background et envoient des alertes automatiquement:

- **Restock Monitor:** Alerte automatique quand un produit de la watchlist est de nouveau en stock
- **Spike Detector:** Alerte automatique quand une carte voit son prix augmenter significativement
- **Arbitrage Finder:** Alerte automatique quand une opportunité d'arbitrage est détectée

**2. Commandes Interactives (Pull)**

L'utilisateur gère sa watchlist et interroge le système via des slash commands:

- `/watch [nom] [prix_revente] [urls...]` - Ajouter un produit avec prix revente estimé
- `/unwatch [nom]` - Retirer un produit de la watchlist
- `/watchlist` - Voir tous les produits surveillés avec prix revente
- `/setprice [nom] [prix]` - Modifier le prix revente d'un produit
- `/alerts` - Gérer ses préférences d'alertes
- `/status` - Voir le statut des monitors

### Pourquoi Discord-First

**MUST requirements:**

- Toute fonctionnalité MUST être accessible via Discord (notifications ou commandes)
- Les monitors MUST fonctionner de manière autonome sans intervention utilisateur
- Les commandes MUST permettre de gérer la watchlist et configurer les alertes
- L'utilisateur MUST pouvoir choisir quelles alertes il reçoit

**Rationale:**

- Les revendeurs Pokemon sont déjà sur Discord (communautés, groupes d'échange)
- Pas de friction: pas d'app à installer, pas de compte à créer
- Notifications push natives (mobile + desktop)
- Réactivité maximale: alertes reçues instantanément là où l'utilisateur est déjà

**MUST NOT:**

- NEVER créer une interface web comme UI principale
- NEVER forcer l'utilisateur à checker manuellement

## Contraintes Business

### Budget

- **Initial:** Quelques milliers d'euros maximum
- **Infrastructure:** Budget VPS + proxies + storage < 100€/mois initial
- **Scaling:** Budget croît avec MRR (max 20% du MRR en infra)

### Timeline

- **Phase 1 MVP (Restock Monitor):** 3-4 semaines
- **Phase 2 (Arbitrage + Spike):** 2-3 mois
- **Phase 3 (Monétisation):** 2-3 mois

### Validation

**MUST requirements:**

- L'outil doit d'abord être utile au créateur lui-même (dogfooding)
- Validation avec 5-10 beta users avant monétisation
- Valeur prouvée sur données réelles avant scaling

### Croissance

- **Canal principal:** Communautés Discord Pokemon FR (organiques)
- **Stratégie:** Bouche-à-oreille via beta users satisfaits
- **Marketing:** Pas de budget ads initial, focus qualité produit

## Ce que le projet N'EST PAS

**MUST NOT implémenter:**

- Bot d'achat automatique (juste des alertes pour décision humaine)
- Outil de gestion de stock/inventaire
- Marketplace pour acheter/vendre directement
- Outil pour cartes gradées (PSA, BGS, etc.) - focus sealed products uniquement
- Outil US-first (focus France/Europe)
- Service de prédiction IA des prix futurs
- Scraper de marketplaces (eBay, Vinted) pour analyse de ventes
- Système de matching automatique de produits

**Rationale:** Rester focus sur la mission core = alertes rapides pour restocks. L'utilisateur garde le contrôle sur la sélection des produits à surveiller.

## Métriques de Succès

### Phase 1 (MVP - Restock Monitor)

- 5 retailers FR monitorés (Pokemon Center, FNAC, Micromania, Amazon, Cultura)
- Latence alerte < 30 secondes (p95)
- 0 faux positifs par semaine (alertes stock erronées)
- 5 beta users avec watchlist active
- Au moins 1 restock capté et actionné par beta user

### Phase 2 (Arbitrage + Spike)

- Arbitrage Finder détecte 5+ opportunités/semaine
- Spike Detector alerte sur variations > 20%
- 10+ beta users actifs

### Phase 3 (Monétisation)

- 50 utilisateurs payants à 6 mois du lancement
- MRR > 500€
- Churn rate < 10% mensuel
- Net Promoter Score > 40

### Long terme (12 mois)

- 200+ utilisateurs payants
- MRR > 3000€
- Fiabilité maintenue (< 1% alertes manquées)

## Roadmap Simplifiée

**Phase 1: Restock Monitor** (3-4 semaines)
→ Alertes de disponibilité

- Système de watchlist (add/remove/list)
- Scrapers retailers FR (Pokemon Center, FNAC, Micromania, Amazon, Cultura)
- Système d'alertes Discord
- **Deliverable:** Alertes temps réel pour restocks de produits surveillés

**Phase 2: Arbitrage Finder** (4-6 semaines)
→ Opportunités de profit

- Comparateur de prix multi-plateformes
- Calculateur profit net
- Alertes arbitrage
- **Deliverable:** Opportunités d'arbitrage quotidiennes

**Phase 3: Spike Detector** (4-6 semaines)
→ Extension aux singles

- Tracker prix CardMarket
- Détecteur de variations anormales
- Alertes spikes avec contexte
- **Deliverable:** Alertes sur hausses de prix significatives

**Phase 4: Monétisation + Scale** (8-12 semaines)
→ Revenus récurrents

- Système de paiement Stripe
- Feature gating par tier
- Analytics utilisateur
- **Deliverable:** SaaS complet avec abonnements payants

## Ton et Communication

### Discord (Interface Principale)

**MUST:**

- Messages concis (< 280 caractères idéalement)
- Emojis pour lisibilité (🔔 alerte, 📦 restock, ✅ en stock)
- Données précises (prix en €, lien direct)
- Call-to-action clair

**Exemple d'alerte restock:**

```
🔔 RESTOCK ALERTE
📦 Coffret Dracaufeu Ultra Premium
💰 Prix retail: 119.99€ @ FNAC
📈 Prix revente: 180€ (ton estimation)
✨ Profit estimé: +60€ (+50%)
🔗 [Acheter maintenant](lien)
⏰ Stock limité détecté
```

**Exemple watchlist:**

```
📋 Ta Watchlist (3 produits)

1. Coffret Dracaufeu UPC
   💰 Revente: 180€
   → FNAC, Pokemon Center, Amazon

2. ETB Ecarlate et Violet
   💰 Revente: 65€
   → Micromania, Cultura

3. Display 151 JAP
   💰 Revente: 140€
   → Pokemon Center
```

### Communication Générale

**MUST:**

- Pas de bullshit: alertes réelles, pas de faux positifs
- Transparence: Si un scraper est down, l'indiquer
- Communautaire: Écouter feedback beta users, itérer rapidement
- Français par défaut (marché FR/EU)

**MUST NOT:**

- NEVER promettre des gains garantis
- NEVER cacher les limites de l'outil
- NEVER ignorer les bugs rapportés par utilisateurs

## Governance

### Amendment Process

1. Proposition d'amendement documentée avec justification
2. Validation contre les 5 principes fondamentaux
3. Review d'impact sur modules existants
4. Mise à jour de ce document
5. Propagation aux templates et documentation

### Version Management

**Semantic Versioning:**

- **MAJOR (X.0.0):** Changement de principe fondamental ou retrait de module core
- **MINOR (0.X.0):** Ajout de nouveau principe, module, ou contrainte significative
- **PATCH (0.0.X):** Clarifications, corrections, ajustements mineurs

### Compliance

**MUST requirements:**

- Toute nouvelle feature MUST être validée contre les 5 principes
- Toute PR MUST vérifier alignement avec Simplicité
- Code reviews MUST valider la simplicité (principe IV)
- Déploiements MUST valider la fiabilité (principe V)

**Review Cadence:**

- Constitution review: tous les 3 mois ou après lancement de phase majeure
- Metrics review: mensuel
- Principles compliance: chaque PR

### Development Guidance

Voir `.specify/templates/plan-template.md` pour guidance d'implémentation. Toute feature doit passer par le workflow: Spec → Plan → Tasks → Implementation.

**Constitution supersedes all other practices.** En cas de conflit entre ce document et d'autres guidelines, la Constitution prévaut.

**Version**: 3.1.0 | **Ratified**: 2026-01-07 | **Last Amended**: 2026-01-07
