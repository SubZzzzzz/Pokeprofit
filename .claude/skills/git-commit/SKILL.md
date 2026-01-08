---
name: git-commit
description: Conventions pour les commits Git. Messages propres, conventionnels, sans reference a l'IA.
allowed-tools: Bash
---

# Git Commit Conventions

## Format

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

## Types

| Type       | Usage                                   |
| ---------- | --------------------------------------- |
| `feat`     | Nouvelle feature                        |
| `fix`      | Bug fix                                 |
| `refactor` | Refactoring sans changement fonctionnel |
| `perf`     | Amelioration performance                |
| `test`     | Ajout/modification tests                |
| `docs`     | Documentation                           |
| `chore`    | Maintenance, deps, config               |
| `ci`       | CI/CD                                   |

## Scopes (PokéProfit)

| Scope      | Description              |
| ---------- | ------------------------ |
| `scraper`  | Scrapers Vinted/LBC      |
| `analyzer` | Detection deals, pricing |
| `bot`      | Discord bot              |
| `db`       | Database, migrations     |
| `config`   | Configuration            |
| `api`      | CardMarket/TCGPlayer API |

## Exemples

```
feat(scraper): add Vinted listing scanner

fix(analyzer): handle missing price field

refactor(bot): extract embed builder to separate package

perf(scraper): implement connection pooling for proxies

test(analyzer): add edge cases for lot detection

docs: update constitution to v4.0.0
```

## Regles

1. **Imperatif present** : "add" pas "added" ou "adds"
2. **Lowercase** : Pas de majuscule au debut
3. **Pas de point** : Fin sans ponctuation
4. **< 72 caracteres** : Pour la premiere ligne
5. **Body optionnel** : Pour expliquer le "pourquoi"
6. **NO AI REFERENCES** : No ai reference or signature in commits unless its part of the feature"

## Breaking Changes

```
feat(analyzer)!: change margin threshold to 30%

BREAKING CHANGE: Alerts now require 30% minimum margin.
Previous threshold was 20%.
```
