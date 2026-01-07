---
name: git-commit
description: Conventions pour les commits Git. Messages propres, conventionnels, SANS référence à l'IA.
allowed-tools: Bash
---

# Git Commit Conventions

## Règle absolue
**JAMAIS de référence à l'IA dans les commits**
- Pas de "Generated with Claude Code"
- Pas de "Co-Authored-By: Claude"
- Commits comme si écrits par un humain

## Format Conventional Commits

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types
- `feat`: Nouvelle fonctionnalité
- `fix`: Correction de bug
- `refactor`: Refactoring (ni feat ni fix)
- `docs`: Documentation
- `style`: Formatting, semicolons, etc
- `test`: Ajout/modification de tests
- `chore`: Maintenance (deps, config, etc)
- `perf`: Amélioration de performance

### Scopes (pour ce projet)
- `scraper`: Scrapers (ebay, vinted, cardmarket)
- `analyzer`: Analyse de données (ROI, volume, etc)
- `bot`: Discord bot
- `db`: Database
- `worker`: Background jobs
- `api`: REST API
- `config`: Configuration

## Exemples

### Nouvelle feature
```
feat(scraper): add eBay scraper for sold Pokemon TCG listings

- Scrape sold listings from last 7 days
- Extract price, date, title, URL
- Rate limiting: 10 req/min
- Tests with mocked HTML
```

### Bug fix
```
fix(analyzer): correct ROI calculation with fees

eBay and PayPal fees were not included in profit calculation.
Now deducts 12.9% + 2.9% + shipping from sell price.
```

### Refactoring
```
refactor(db): extract repository pattern for sales

- Create SalesRepository with CRUD methods
- Move queries from handlers to repository
- Add bulk insert for scraper performance
```

### Chore
```
chore: add Claude Code agents and skills

- Added subagents: scraper-builder, debugger, code-reviewer, worker-builder
- Added skills: go-scraping, discord-bot, database, data-analysis, redis
- Added project documentation (CLAUDE.md)
- Added .gitignore for Go project
```

## Best Practices

1. **Subject ligne** : Impératif, lowercase, max 72 chars
2. **Body** : Expliquer le "pourquoi", pas le "quoi"
3. **Listes** : Utiliser `-` pour lister les changes
4. **Breaking changes** : Prefix avec `BREAKING CHANGE:`
5. **Issues** : Référencer avec `Fixes #123` ou `Closes #456`

## Anti-patterns

❌ `git commit -m "fix"`
❌ `git commit -m "WIP"`
❌ `git commit -m "Updated stuff"`
❌ `git commit -m "Generated with Claude Code"`

✅ `git commit -m "fix(scraper): handle pagination for eBay results"`
✅ `git commit -m "feat(bot): add /top command to show profitable products"`
