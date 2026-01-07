<!--
SYNC IMPACT REPORT
==================
Version Change: 0.0.0 → 1.0.0
Rationale: Initial constitution creation for PokéProfit project

Modified Principles:
- N/A (initial creation)

Added Sections:
- Core Principles (5 principles: Data-Driven, Speed Matters, ROI First, Simplicité, Fiabilité)
- Scope Fonctionnel (5 modules)
- Contraintes Techniques
- Contraintes Business
- Ce que le projet N'EST PAS
- Métriques de Succès
- Roadmap Simplifiée
- Ton et Communication
- Governance

Removed Sections:
- N/A (initial creation)

Templates Status:
✅ .specify/templates/plan-template.md - reviewed, compatible
✅ .specify/templates/spec-template.md - reviewed, compatible
✅ .specify/templates/tasks-template.md - reviewed, compatible
⚠ .specify/templates/commands/*.md - no command files found in templates/commands/

Follow-up TODOs:
- None
-->

# PokéProfit Constitution

## Core Principles

### I. Data-Driven
Toutes les décisions sont basées sur des données réelles du marché, pas sur des suppositions. On analyse ce qui SE VEND, pas ce qu'on PENSE qui va se vendre.

**MUST requirements:**
- Scraper les ventes complétées (eBay FR, Vinted) pour identifier les produits rentables
- Calculer volume de ventes, prix moyen, marge vs MSRP pour chaque produit
- Baser les recommandations uniquement sur des données mesurables et vérifiables
- NEVER faire de recommandations basées sur des intuitions ou tendances non-vérifiées

**Rationale:** Dans le reselling, les pertes proviennent d'achats basés sur des suppositions. Seules les données de ventes réelles révèlent ce qui est effectivement rentable sur le marché actuel.

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
- Chaque alerte MUST inclure le calcul de ROI (prix retail vs prix marché)
- Prioriser les features qui augmentent directement le profit utilisateur
- Rejeter les features "nice-to-have" qui ne contribuent pas au ROI
- Mesurer le succès en euros gagnés, pas en features livrées

**MUST NOT:**
- NEVER implémenter des features purement esthétiques
- NEVER créer des dashboards complexes sans valeur actionnable
- NEVER ajouter de la complexité qui n'améliore pas les profits

**Rationale:** Les utilisateurs paient pour gagner de l'argent, pas pour des interfaces jolies. Chaque euro de développement doit générer des euros de profit utilisateur.

### IV. Simplicité
L'utilisateur veut des réponses claires : "achète ça", "vends ça", "profit = X€". Pas de dashboards complexes inutiles.

**MUST requirements:**
- Messages Discord concis avec les informations essentielles (produit, prix, ROI, lien)
- Commandes simples et intuitives (`/top`, `/alerts`)
- Pas de configuration complexe requise pour obtenir de la valeur
- Données chiffrées claires et actionnables

**MUST NOT:**
- NEVER créer des interfaces nécessitant une formation
- NEVER cacher l'information essentielle derrière des clics multiples
- NEVER utiliser du jargon technique face à l'utilisateur

**Rationale:** L'utilisateur est un revendeur occupé, pas un data analyst. Il a besoin d'informations claires pour prendre des décisions rapides.

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

### Module 1: Volume Analyzer (CORE)
**But:** Identifier les produits rentables via l'analyse des ventes réelles

**Composants:**
- Scraper eBay FR (ventes complétées)
- Scraper Vinted (ventes complétées)
- Calculateur: volume de ventes, prix moyen, marge vs MSRP
- Scorer: Volume × Marge = Score rentabilité
- Discord bot command: `/top` pour exposer les top produits

**Principe:** Le marché nous dit ce qui est rentable, on ne devine pas.

### Module 2: Restock Monitor
**But:** Alerter quand les produits rentables sont disponibles

**Composants:**
- Moniteurs pour retailers FR: Pokemon Center, FNAC, Micromania, Amazon, Cultura
- Intégration avec Module 1 pour calculer ROI par produit
- Système d'alertes Discord avec: lien direct, prix, ROI calculé, stock disponible

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
- **Free:** Accès limité aux top 5 produits, 3 alertes/jour
- **Pro (15€/mois):** Accès complet Volume Analyzer + Restock Monitor, alertes illimitées
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
- Refresh des données Volume Analyzer toutes les 24h minimum
- API Discord répondant en < 500ms
- Database queries optimisées (indexes, no N+1)

**Benchmarks:**
- Latency p95 < 30s pour alertes restock
- Throughput: 1000 produits scannés en < 5 minutes
- Memory usage < 512MB (base) + 1MB per 100 produits

## Contraintes Business

### Budget
- **Initial:** Quelques milliers d'euros maximum
- **Infrastructure:** Budget VPS + proxies + storage < 100€/mois initial
- **Scaling:** Budget croît avec MRR (max 20% du MRR en infra)

### Timeline
- **Phase 1 MVP (Volume Analyzer):** 3-4 semaines
- **Phase 2 (Restock Monitor):** 3-4 semaines
- **Phase 1+2 total:** 2-3 mois pour MVP complet
- **Phase 3-4:** 2-3 mois additionnels
- **Phase 5 (Web Dashboard):** 2-3 mois

### Validation
**MUST requirements:**
- L'outil doit d'abord être utile au créateur lui-même (dogfooding)
- Validation avec 5-10 beta users avant monétisation
- ROI prouvé sur données réelles avant scaling

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

**Rationale:** Rester focus sur la mission core = alertes intelligentes pour maximiser ROI. Éviter la feature creep qui dilue la valeur.

## Métriques de Succès

### Phase 1 (MVP - Volume Analyzer)
- Identifier 10+ produits rentables par semaine
- Taux de précision ROI > 80% (prédictions vs résultats réels)
- 5 beta users utilisent l'outil activement

### Phase 2 (Restock Monitor)
- Latence alerte < 30 secondes (p95)
- 0 faux positifs par semaine (alertes stock erronées)
- 10+ beta users utilisent les alertes
- Conversion alerte → achat > 20%

### Phase 3 (Monétisation)
- 50 utilisateurs payants à 6 mois du lancement
- MRR > 500€
- Churn rate < 10% mensuel
- Net Promoter Score > 40

### Long terme (12 mois)
- 200+ utilisateurs payants
- MRR > 3000€
- Taux de précision ROI maintenu > 80%
- Feature requests alignées avec ROI First principle

## Roadmap Simplifiée

**Phase 1: Volume Analyzer** (3-4 semaines)
→ Savoir QUOI acheter
- Scrapers eBay FR + Vinted
- Database schema + calculateurs
- Discord bot `/top` command
- **Deliverable:** Liste des top produits rentables mise à jour quotidiennement

**Phase 2: Restock Monitor** (3-4 semaines)
→ Savoir QUAND acheter
- Scrapers retailers FR (Pokemon Center, FNAC, Micromania, Amazon, Cultura)
- Système d'alertes Discord
- Intégration ROI avec Volume Analyzer
- **Deliverable:** Alertes temps réel pour restocks de produits rentables

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

**Phase 5: Dashboard Web + Scale** (8-12 semaines)
→ Monétisation et croissance
- Interface web pour configuration
- Système de paiement Stripe
- Feature gating par tier
- Analytics utilisateur
- **Deliverable:** SaaS complet avec abonnements payants

## Ton et Communication

### Discord (Interface Principale)
**MUST:**
- Messages concis (< 280 caractères idéalement)
- Emojis pour lisibilité (📈 profit, 🔔 alerte, 💰 ROI)
- Données chiffrées précises (prix en €, ROI en %, volume en unités)
- Call-to-action clair (lien direct vers produit)

**Exemple d'alerte:**
```
🔔 RESTOCK ALERTE
📦 Coffret Dracaufeu Ultra Premium
💰 Prix: 119.99€ | Vente moyenne: 179€
📈 ROI estimé: +49% (59€ profit)
🔗 [Acheter maintenant](lien)
⏰ Stock limité détecté
```

### Communication Générale
**MUST:**
- Pas de bullshit: ROI réel basé sur données, pas de promesses exagérées
- Transparence: Si une alerte était fausse, l'admettre et corriger
- Communautaire: Écouter feedback beta users, itérer rapidement
- Français par défaut (marché FR/EU)

**MUST NOT:**
- NEVER promettre des gains garantis
- NEVER cacher les risques du reselling
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
- Toute PR MUST vérifier alignement avec ROI First
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

**Version**: 1.0.0 | **Ratified**: 2026-01-07 | **Last Amended**: 2026-01-07
