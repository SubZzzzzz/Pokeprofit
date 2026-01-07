# Feature Specification: Volume Analyzer Phase 1 - eBay

**Feature Branch**: `001-ebay-volume-analyzer`
**Created**: 2026-01-07
**Status**: Draft
**Input**: User description: "J'aimerais intégrer le volume analyzer (phase 1), qui analyserait seulement eBay pour l'instant"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Découverte des produits à fort volume (Priority: P1)

En tant qu'utilisateur, je veux voir quels produits Pokemon TCG se vendent le plus sur eBay afin d'identifier les opportunités de revente rentables.

**Why this priority**: C'est la fonctionnalité centrale - sans données de volume, aucune analyse de rentabilité n'est possible.

**Independent Test**: Peut être testé en lançant une analyse sur une catégorie Pokemon TCG et en vérifiant que les produits sont triés par volume de ventes.

**Acceptance Scenarios**:

1. **Given** un utilisateur connecté, **When** il lance une analyse de volume, **Then** il voit une liste de produits triés par nombre de ventes récentes
2. **Given** une analyse terminée, **When** l'utilisateur consulte les résultats, **Then** il voit pour chaque produit: nom, prix moyen, et nombre de ventes

---

### User Story 2 - Calcul de rentabilité par produit (Priority: P2)

En tant qu'utilisateur, je veux voir le potentiel de profit de chaque produit analysé afin de prioriser mes achats de revente.

**Why this priority**: Transforme les données brutes en informations actionnables pour l'utilisateur.

**Independent Test**: Peut être testé en vérifiant que chaque produit affiche un indicateur de rentabilité basé sur les prix d'achat vs revente.

**Acceptance Scenarios**:

1. **Given** des résultats d'analyse disponibles, **When** l'utilisateur consulte un produit, **Then** il voit une estimation de marge potentielle
2. **Given** un produit avec plusieurs ventes, **When** le système calcule la rentabilité, **Then** il utilise le prix moyen des ventes récentes (30 derniers jours)

---

### User Story 3 - Filtrage et recherche ciblée (Priority: P3)

En tant qu'utilisateur, je veux filtrer les résultats par catégorie de produit (boosters, displays, cartes singles) afin de me concentrer sur mon créneau de revente.

**Why this priority**: Améliore l'expérience utilisateur mais n'est pas critique pour le MVP.

**Independent Test**: Peut être testé en appliquant différents filtres et en vérifiant que seuls les produits correspondants sont affichés.

**Acceptance Scenarios**:

1. **Given** des résultats d'analyse, **When** l'utilisateur filtre par "boosters", **Then** seuls les boosters Pokemon sont affichés
2. **Given** des filtres appliqués, **When** l'utilisateur réinitialise les filtres, **Then** tous les produits redeviennent visibles

---

### Edge Cases

- Que se passe-t-il quand eBay est temporairement inaccessible? Le système affiche un message d'erreur clair et utilise les données en cache si disponibles.
- Comment le système gère-t-il un produit avec zéro vente récente? Il est exclu des résultats ou affiché avec un indicateur "données insuffisantes".
- Que se passe-t-il si le taux de requêtes dépasse les limites d'eBay? Le système applique un rate limiting automatique et reprend l'analyse progressivement.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Le système DOIT collecter les données de ventes Pokemon TCG depuis eBay (ventes complétées) via scraping HTML des pages "ventes terminées"
- **FR-002**: Le système DOIT calculer le volume de ventes par produit sur les 30 derniers jours
- **FR-003**: Le système DOIT calculer le prix moyen de vente par produit
- **FR-004**: Le système DOIT afficher les résultats triés par volume de ventes (décroissant par défaut)
- **FR-005**: Le système DOIT permettre le filtrage par catégorie de produit (boosters, displays, singles, ETB, collections)
- **FR-006**: Le système DOIT stocker les données analysées pour consultation ultérieure
- **FR-007**: Le système DOIT afficher une estimation de marge pour chaque produit (prix moyen de vente - MSRP officiel)
- **FR-008**: Le système DOIT respecter un rate limiting pour éviter le blocage par eBay
- **FR-009**: Le système DOIT gérer les erreurs de connexion avec des messages utilisateur compréhensibles
- **FR-010**: Le système DOIT permettre de lancer une nouvelle analyse manuellement
- **FR-011**: Le bot Discord DOIT utiliser des commandes slash (/analyze, /results, /filter) comme interface principale
- **FR-012**: Les résultats DOIVENT être affichés via des embeds Discord riches (tableaux formatés, couleurs, champs structurés)

### Key Entities

- **Produit analysé**: Représente un produit Pokemon TCG identifié sur eBay - nom normalisé (set + type), catégorie, lien source. Identifié par matching de mots-clés normalisés pour regrouper les annonces similaires.
- **Vente**: Représente une transaction complétée - prix de vente, date, lien vers l'annonce
- **Analyse**: Session d'analyse groupant les résultats - date d'exécution, nombre de produits analysés, statut
- **Statistiques produit**: Données agrégées par produit - volume de ventes, prix moyen, prix min/max, tendance
- **Référentiel MSRP**: Prix retail officiels par type de produit Pokemon TCG (boosters, displays, ETB, etc.) utilisés comme base de calcul de marge

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Les utilisateurs peuvent consulter les données de volume pour au moins 100 produits Pokemon TCG en une seule analyse
- **SC-002**: L'analyse complète d'une catégorie de produits s'effectue en moins de 10 minutes
- **SC-003**: Les données de prix sont actualisées quotidiennement
- **SC-004**: 90% des utilisateurs interagissent avec au moins un produit (clic lien eBay ou commande /filter) lors de leur première session d'analyse, mesuré via logs Discord
- **SC-005**: Le système reste fonctionnel et collecte des données sans blocage pendant au moins 7 jours consécutifs

## Assumptions

- L'utilisateur a une connaissance de base du marché Pokemon TCG et sait interpréter les données de volume
- eBay reste la source principale pour la phase 1 - d'autres plateformes seront ajoutées dans les phases ultérieures
- Les prix affichés sont en EUR (marché européen ciblé en priorité)
- Le rate limiting d'eBay permet de collecter suffisamment de données pour une analyse pertinente
- L'utilisateur accède au système via un bot Discord (interface principale du projet)

## Out of Scope (Phase 1)

- Intégration avec d'autres plateformes (Vinted, CardMarket) - prévu pour phases ultérieures
- Alertes automatiques de nouvelles opportunités
- Historique de prix sur longue période (> 30 jours)
- Comparaison de prix cross-plateforme

## Clarifications

### Session 2026-01-07

- Q: Comment le système doit-il accéder aux données de ventes eBay? → A: Scraping HTML des pages eBay (ventes terminées)
- Q: Quel type de commandes le bot Discord doit-il utiliser? → A: Commandes slash Discord (/analyze, /results)
- Q: Comment le système doit-il identifier et regrouper les produits similaires? → A: Matching par mots-clés normalisés (nom set + type produit)
- Q: Quel format d'affichage Discord pour les résultats? → A: Embeds Discord riches (tableaux formatés, couleurs)
- Q: Comment déterminer le prix d'achat pour le calcul de marge? → A: Prix retail officiel (MSRP) stocké par produit
