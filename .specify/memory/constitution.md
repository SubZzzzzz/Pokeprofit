<!--
SYNC IMPACT REPORT
==================
Version Change: 1.0.1 → 2.0.0
Rationale: MAJOR - Retrait du Module 1 (Volume Analyzer) et remplacement par
Data Tracker simplifié. Changement de philosophie: automatisation complète
vers données + analyse manuelle.

Modified Principles:
- Principe I "Data-Driven" → Reformulé: focus sur collecte de données, pas calculs automatisés

Added Sections:
- Module 1 remplacé: "Data Tracker" (liste des ventes/restocks observées)

Removed Sections:
- Volume Analyzer (scrapers eBay/Vinted avec calculs ROI automatisés)
- Scorer automatique (Volume × Marge)
- Calculs automatisés de rentabilité

Templates Status:
✅ .specify/templates/plan-template.md - compatible (Constitution Check générique)
✅ .specify/templates/spec-template.md - compatible (user stories génériques)
✅ .specify/templates/tasks-template.md - compatible (structure de phases générique)
⚠ .specify/templates/commands/*.md - aucun fichier présent

Follow-up TODOs:
- Mettre à jour les métriques de succès Phase 1 (adapter aux nouvelles fonctions)
- Valider avec utilisateur si d'autres modules doivent être simplifiés de manière similaire
-->

# PokéProfit Constitution

## Core Principles

### I. Data-Driven

Toutes les décisions sont basées sur des données réelles du marché. On collecte les données observables (ventes, restocks), l'analyse de rentabilité reste à la discrétion de l'utilisateur.

**MUST requirements:**

- Tracker les ventes complétées (eBay FR, Vinted) pour exposer les tendances
- Tracker les restocks et disponibilités sur les retailers
- Présenter les données de manière claire et exploitable
- Permettre à l'utilisateur de faire ses propres analyses avec les données fournies

**Rationale:** Automatiser la collecte de données est fiable et scalable. L'analyse de rentabilité dépend de critères personnels (coûts d'envoi, temps disponible, objectifs). L'utilisateur est le mieux placé pour décider ce qui est rentable pour lui.

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
- Fournir les données nécessaires pour que l'utilisateur calcule son ROI
- Rejeter les features "nice-to-have" qui ne contribuent pas au ROI
- Mesurer le succès en valeur apportée à l'utilisateur

**MUST NOT:**

- NEVER implémenter des features purement esthétiques
- NEVER créer des dashboards complexes sans valeur actionnable
- NEVER ajouter de la complexité qui n'améliore pas les profits

**Rationale:** Les utilisateurs paient pour gagner de l'argent, pas pour des interfaces jolies. Chaque euro de développement doit générer des euros de profit utilisateur.

### IV. Simplicité

L'utilisateur veut des données claires et exploitables. Pas de dashboards complexes inutiles.

**MUST requirements:**

- Messages Discord concis avec les informations essentielles
- Commandes simples et intuitives (`/sales`, `/restocks`, `/alerts`)
- Pas de configuration complexe requise pour obtenir de la valeur
- Données brutes accessibles pour analyse personnelle

**MUST NOT:**

- NEVER créer des interfaces nécessitant une formation
- NEVER cacher l'information essentielle derrière des clics multiples
- NEVER utiliser du jargon technique face à l'utilisateur

**Rationale:** L'utilisateur est un revendeur occupé. Il a besoin d'informations claires pour prendre ses propres décisions.

### V. Fiabilité

Les scrapers doivent être robustes. Une donnée manquée = information incomplète = mauvaise décision potentielle.

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

**Rationale:** La fiabilité est la base de la confiance. Si l'outil manque des données, l'utilisateur perd confiance. Un système fiable = utilisateurs qui restent et paient.

## Scope Fonctionnel

### Module 1: Data Tracker (CORE)

**But:** Collecter et exposer les données de marché pour analyse manuelle par l'utilisateur

**Composants:**

- Scraper eBay FR (ventes complétées) - expose: produit, prix vendu, date
- Scraper Vinted (ventes complétées) - expose: produit, prix vendu, date
- Database stockant l'historique des ventes observées
- Discord bot command: `/sales [produit]` pour voir l'historique des ventes
- Discord bot command: `/trending` pour voir les produits avec le plus de ventes récentes

**Principe:** On collecte les données, l'utilisateur analyse. Flexibilité maximale, complexité minimale.

**Pourquoi ce changement:** Le Volume Analyzer automatisé était trop rigide - les critères de rentabilité varient selon chaque revendeur (frais, localisation, temps). Exposer les données brutes permet à chacun d'appliquer ses propres critères.

### Module 2: Restock Monitor

**But:** Alerter quand les produits sont disponibles chez les retailers

**Composants:**

- Moniteurs pour retailers FR: Pokemon Center, FNAC, Micromania, Amazon, Cultura
- Système d'alertes Discord avec: lien direct, prix retail, stock disponible
- Configuration utilisateur: quels produits surveiller

**Principe:** Alerte = Action immédiate possible (toutes les infos nécessaires présentes)

### Module 3: Arbitrage Finder

**But:** Détecter les différences de prix entre plateformes

**Composants:**

- Comparateur de prix: CardMarket vs eBay vs Vinted
- Calculateur de profit net après frais (commissions, shipping)
- Alertes quand opportunité > seuil défini (configurable par tier)

**Principe:** Arbitrage = profit quasi sans risque si bien exécuté

### Module 4: Spike Detector

**But:** Détecter les hausses de prix anormales sur les cartes (singles)

**Composants:**

- Tracker de prix CardMarket pour cartes populaires
- Détecteur de variations > X% en Y heures
- Système d'alerte avec contexte (cause probable du spike)

**Principe:** Information = pouvoir (vendre avant les autres, ou acheter avant que ça monte)

### Module 5: Monétisation

**But:** Générer des revenus récurrents

**Composants:**

- Système de tiers: Free (limité), Pro (15€/mois), Business (35€/mois)
- Intégration Stripe pour paiements
- Feature gating par tier
- Gestion des abonnements et renouvellements

**Tiers:**

- **Free:** Accès limité aux données récentes (7 jours), 3 alertes restock/jour
- **Pro (15€/mois):** Historique complet + Restock Monitor illimité
- **Business (35€/mois):** Tout Pro + Arbitrage Finder + Spike Detector + alertes prioritaires

## Contraintes Techniques

### Stack Imposé

- **Backend:** Go (Golang) - performance et concurrence native pour scrapers
- **Database:** PostgreSQL - données relationnelles (produits, ventes, utilisateurs)
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
- Refresh des données sales tracker toutes les 24h minimum
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

Les monitors tournent en background et envoient des alertes automatiquement quand un événement est détecté:

- **Restock Monitor:** Alerte automatique quand un produit surveillé est de nouveau en stock
- **Spike Detector:** Alerte automatique quand une carte voit son prix augmenter significativement
- **Arbitrage Finder:** Alerte automatique quand une opportunité d'arbitrage est détectée

**2. Commandes Interactives (Pull)**

L'utilisateur peut interroger le système à la demande via des slash commands:

- `/sales [produit]` - Voir l'historique des ventes pour un produit donné
- `/trending` - Voir les produits avec le plus de ventes récentes
- `/alerts` - Gérer ses préférences d'alertes
- `/watchlist` - Gérer sa liste de produits à surveiller
- `/stats` - Voir ses statistiques personnelles

### Pourquoi Discord-First

**MUST requirements:**

- Toute fonctionnalité MUST être accessible via Discord (notifications ou commandes)
- Les monitors MUST fonctionner de manière autonome sans intervention utilisateur
- Les commandes MUST permettre d'interroger/configurer le système à la demande
- L'utilisateur MUST pouvoir choisir quelles alertes automatiques il reçoit

**Rationale:**

- Les revendeurs Pokemon sont déjà sur Discord (communautés, groupes d'échange)
- Pas de friction: pas d'app à installer, pas de compte à créer
- Notifications push natives (mobile + desktop)
- Réactivité maximale: alertes reçues instantanément là où l'utilisateur est déjà

**MUST NOT:**

- NEVER créer une interface web comme UI principale (peut être ajouté plus tard pour config avancée uniquement)
- NEVER forcer l'utilisateur à checker manuellement - les alertes importantes arrivent automatiquement

## Contraintes Business

### Budget

- **Initial:** Quelques milliers d'euros maximum
- **Infrastructure:** Budget VPS + proxies + storage < 100€/mois initial
- **Scaling:** Budget croît avec MRR (max 20% du MRR en infra)

### Timeline

- **Phase 1 MVP (Data Tracker):** 2-3 semaines
- **Phase 2 (Restock Monitor):** 3-4 semaines
- **Phase 1+2 total:** 6-8 semaines pour MVP complet
- **Phase 3-4:** 2-3 mois additionnels
- **Phase 5 (Monétisation):** 2-3 mois

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
- Service de prédiction IA des prix futurs (data-driven seulement)
- Calculateur de rentabilité automatisé (l'utilisateur fait ses propres analyses)

**Rationale:** Rester focus sur la mission core = données fiables + alertes rapides. L'utilisateur garde le contrôle sur l'analyse et les décisions.

## Métriques de Succès

### Phase 1 (MVP - Data Tracker)

- Tracker 50+ produits avec historique de ventes
- Données mises à jour quotidiennement
- 5 beta users utilisent l'outil activement
- Commande `/sales` répond en < 2 secondes

### Phase 2 (Restock Monitor)

- Latence alerte < 30 secondes (p95)
- 0 faux positifs par semaine (alertes stock erronées)
- 10+ beta users utilisent les alertes
- Utilisateurs déclarent avoir profité d'au moins 1 restock grâce aux alertes

### Phase 3 (Monétisation)

- 50 utilisateurs payants à 6 mois du lancement
- MRR > 500€
- Churn rate < 10% mensuel
- Net Promoter Score > 40

### Long terme (12 mois)

- 200+ utilisateurs payants
- MRR > 3000€
- Données fiables maintenues (< 5% erreurs rapportées)
- Feature requests alignées avec principes de simplicité

## Roadmap Simplifiée

**Phase 1: Data Tracker** (2-3 semaines)
→ Exposer les données de marché

- Scrapers eBay FR + Vinted (ventes complétées)
- Database schema pour stocker historique
- Discord bot `/sales` et `/trending` commands
- **Deliverable:** Historique des ventes accessible via Discord

**Phase 2: Restock Monitor** (3-4 semaines)
→ Savoir QUAND acheter

- Scrapers retailers FR (Pokemon Center, FNAC, Micromania, Amazon, Cultura)
- Système d'alertes Discord
- Watchlist utilisateur
- **Deliverable:** Alertes temps réel pour restocks

**Phase 3: Arbitrage Finder** (4-6 semaines)
→ Nouvelles opportunités de profit

- Comparateur de prix multi-plateformes
- Calculateur profit net
- Alertes arbitrage
- **Deliverable:** Opportunités d'arbitrage quotidiennes

**Phase 4: Spike Detector** (4-6 semaines)
→ Extension aux singles

- Tracker prix CardMarket
- Détecteur de variations anormales
- Alertes spikes avec contexte
- **Deliverable:** Alertes sur hausses de prix significatives

**Phase 5: Monétisation + Scale** (8-12 semaines)
→ Revenus récurrents

- Système de paiement Stripe
- Feature gating par tier
- Analytics utilisateur
- **Deliverable:** SaaS complet avec abonnements payants

## Ton et Communication

### Discord (Interface Principale)

**MUST:**

- Messages concis (< 280 caractères idéalement)
- Emojis pour lisibilité (📊 data, 🔔 alerte, 📦 restock)
- Données chiffrées précises (prix en €, quantités, dates)
- Call-to-action clair (lien direct vers produit)

**Exemple d'alerte restock:**

```
🔔 RESTOCK ALERTE
📦 Coffret Dracaufeu Ultra Premium
💰 Prix: 119.99€ @ FNAC
🔗 [Acheter maintenant](lien)
⏰ Stock limité détecté
```

**Exemple de données sales:**

```
📊 VENTES: Coffret Dracaufeu UPC
Dernières 7 jours:
- eBay: 15 ventes, 145€-185€ (moy: 168€)
- Vinted: 8 ventes, 135€-160€ (moy: 148€)
```

### Communication Générale

**MUST:**

- Pas de bullshit: données réelles, pas de promesses exagérées
- Transparence: Si des données sont manquantes, l'indiquer
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
- Metrics MUST être trackées selon "Métriques de Succès"

**Review Cadence:**

- Constitution review: tous les 3 mois ou après lancement de phase majeure
- Metrics review: mensuel
- Principles compliance: chaque PR

### Development Guidance

Voir `.specify/templates/plan-template.md` pour guidance d'implémentation. Toute feature doit passer par le workflow: Spec → Plan → Tasks → Implementation.

**Constitution supersedes all other practices.** En cas de conflit entre ce document et d'autres guidelines, la Constitution prévaut.

**Version**: 2.0.0 | **Ratified**: 2026-01-07 | **Last Amended**: 2026-01-07
