# Quickstart: Volume Analyzer Phase 1 - eBay

**Date**: 2026-01-07
**Feature Branch**: `001-ebay-volume-analyzer`

## Prerequisites

- Go 1.21+
- PostgreSQL 14+
- Redis 7+
- Discord Bot Token (with slash commands permission)

---

## Setup

### 1. Clone and Install Dependencies

```bash
cd ~/projects/pokeprofit
go mod init github.com/youruser/pokeprofit
go mod tidy
```

### 2. Environment Configuration

Create `.env` file at project root:

```env
# Database
DATABASE_URL=postgres://user:password@localhost:5432/pokeprofit?sslmode=disable

# Redis
REDIS_URL=redis://localhost:6379/0

# Discord
DISCORD_TOKEN=your_bot_token_here
DISCORD_GUILD_ID=your_test_guild_id

# Scraper
SCRAPER_RATE_LIMIT_MS=2000
SCRAPER_MAX_RETRIES=3
SCRAPER_USER_AGENTS=Mozilla/5.0 (Windows NT 10.0; Win64; x64)...,Mozilla/5.0 (Macintosh)...

# Logging
LOG_LEVEL=info
```

### 3. Database Setup

```bash
# Create database
createdb pokeprofit

# Run migrations
go run cmd/migrate/main.go up

# (Or manually apply SQL from data-model.md)
```

### 4. Discord Bot Setup

1. Go to [Discord Developer Portal](https://discord.com/developers/applications)
2. Create new application
3. Bot → Add Bot → Copy Token
4. OAuth2 → URL Generator:
   - Scopes: `bot`, `applications.commands`
   - Bot Permissions: `Send Messages`, `Embed Links`, `Use Slash Commands`
5. Invite bot to your test server

---

## Running

### Development Mode

```bash
# Run bot with hot reload (requires air)
go install github.com/cosmtrek/air@latest
air

# Or directly
go run cmd/bot/main.go
```

### Manual Analysis (CLI)

```bash
# Run analysis without Discord
go run cmd/analyzer/main.go --query "Pokemon Display 151" --max-pages 10
```

---

## Testing Discord Commands

Once the bot is running, test in Discord:

```
/analyze query:Pokemon Display 151
```

Wait for analysis to complete, then:

```
/results sort:margin_percent limit:10
```

Filter by category:

```
/filter category:display
```

---

## Project Structure Quick Reference

```
cmd/
  bot/main.go         # Discord bot entry point
  analyzer/main.go    # CLI analysis tool
  migrate/main.go     # Database migrations

internal/
  config/             # Env loading
  database/           # PostgreSQL connection
  cache/              # Redis client
  scraper/ebay/       # eBay scraper
  analyzer/           # Business logic
  repository/         # Data access
  discord/            # Bot handlers
  models/             # Domain entities
```

---

## Key Files to Implement (in order)

1. `internal/config/config.go` - Configuration loading
2. `internal/database/postgres.go` - DB connection
3. `internal/models/*.go` - Domain entities
4. `internal/scraper/ebay/scraper.go` - eBay scraper
5. `internal/analyzer/normalizer.go` - Product matching
6. `internal/repository/*.go` - Data persistence
7. `internal/analyzer/volume.go` - Analysis orchestration
8. `internal/discord/bot.go` - Bot setup
9. `internal/discord/commands/*.go` - Slash commands
10. `cmd/bot/main.go` - Entry point

---

## Common Issues

### "rate limited by eBay"

- Ensure `SCRAPER_RATE_LIMIT_MS` is at least 2000 (2 seconds)
- Check if your IP is temporarily blocked (use proxy)

### "no data available"

- Run `/analyze` first before `/results`
- Check database connection

### "slash commands not appearing"

- Wait up to 1 hour for global commands
- For instant testing, use guild-specific commands (set `DISCORD_GUILD_ID`)

### "product not matched"

- Check normalizer patterns in `internal/analyzer/normalizer.go`
- Add new patterns for unrecognized products

---

## Next Steps After MVP

1. ~~Add more Pokemon TCG sets to MSRP reference~~ (Done - see migration 004)
2. Implement proxy rotation for scraper
3. Add scheduled analysis (cron job)
4. ~~Implement pagination for /results~~ (Done - Phase 5)
5. Add user-specific settings storage

---

## Implemented Features (Phase 6 Complete)

- **Structured Logging**: JSON-formatted logs throughout all packages via `internal/logger/`
- **Graceful Shutdown**: Bot handles SIGINT/SIGTERM with proper cleanup
- **User Rate Limiting**: Discord command rate limits (`/analyze`: 1/5min, `/results`: 5/min, `/filter`: 10/min)
- **Health Check Endpoint**: `scraper.HealthCheck()` returns detailed status with response times
- **Integration Tests**: Full test coverage for rate limiting and error handling
- **Additional MSRP Data**: 2024-2025 Pokemon TCG sets with MSRP reference
