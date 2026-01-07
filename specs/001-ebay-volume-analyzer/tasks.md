# Tasks: Volume Analyzer Phase 1 - eBay

**Input**: Design documents from `/specs/001-ebay-volume-analyzer/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Not explicitly requested - test tasks are omitted.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

Based on plan.md Go project structure:
- `cmd/` for entry points
- `internal/` for private packages
- `tests/` for integration tests and mocks

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [x] T001 Initialize Go module with `go mod init` at repository root
- [x] T002 [P] Create project directory structure per plan.md (cmd/, internal/, tests/)
- [x] T003 [P] Create .env.example with all required environment variables in project root
- [x] T004 [P] Add .gitignore for Go projects at repository root

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**CRITICAL**: No user story work can begin until this phase is complete

- [x] T005 Implement configuration loader in internal/config/config.go
- [x] T006 [P] Implement PostgreSQL connection pool in internal/database/postgres.go
- [x] T007 [P] Implement Redis client wrapper in internal/cache/redis.go
- [x] T008 Create database migration runner in cmd/migrate/main.go
- [x] T009 Create initial migration with full schema in internal/database/migrations/001_initial_schema.sql
- [x] T010 [P] Define domain models (Product, Sale, Analysis, ProductStats) in internal/models/product.go, internal/models/sale.go, internal/models/analysis.go, internal/models/stats.go
- [x] T011 [P] Define custom error types in internal/errors/errors.go
- [x] T012 [P] Implement rate limiter utility in internal/scraper/common/ratelimit.go
- [x] T013 [P] Implement retry utility with exponential backoff in internal/scraper/common/retry.go
- [x] T014 Implement base repository with sqlx connection in internal/repository/base.go

**Checkpoint**: Foundation ready - user story implementation can now begin

---

## Phase 3: User Story 1 - Discover High-Volume Products (Priority: P1) MVP

**Goal**: Allow users to see which Pokemon TCG products sell the most on eBay to identify profitable resale opportunities.

**Independent Test**: Launch an analysis on a Pokemon TCG category and verify products are sorted by sales volume.

### Implementation for User Story 1

- [x] T015 [P] [US1] Define scraper types (RawSale, ScrapeOptions, ScrapeResult) in internal/scraper/ebay/types.go
- [x] T016 [P] [US1] Implement HTML parser for eBay completed listings in internal/scraper/ebay/parser.go
- [x] T017 [US1] Implement eBay scraper with colly in internal/scraper/ebay/scraper.go
- [x] T018 [US1] Implement ProductRepository (FindByNormalizedName, FindOrCreate, List) in internal/repository/product.go
- [x] T019 [US1] Implement SaleRepository (Create, BulkCreate, FindByProductID) in internal/repository/sale.go
- [x] T020 [US1] Implement AnalysisRepository (Create, Update, GetLatest, GetByID) in internal/repository/analysis.go
- [x] T021 [P] [US1] Implement product name normalizer in internal/analyzer/normalizer.go
- [x] T022 [US1] Implement VolumeAnalyzer orchestrator (Run, GetStatus) in internal/analyzer/volume.go
- [x] T023 [P] [US1] Implement Discord bot setup and connection in internal/discord/bot.go
- [x] T024 [US1] Implement /analyze slash command handler in internal/discord/commands/analyze.go
- [x] T025 [US1] Implement /results slash command handler in internal/discord/commands/results.go
- [x] T026 [P] [US1] Implement results embed builder in internal/discord/embeds/results.go
- [x] T027 [US1] Create bot entry point in cmd/bot/main.go
- [x] T028 [US1] Create CLI analyzer entry point in cmd/analyzer/main.go

**Checkpoint**: User Story 1 complete - users can run /analyze and /results to see products sorted by volume

---

## Phase 4: User Story 2 - Calculate Product Profitability (Priority: P2)

**Goal**: Show potential profit for each analyzed product to help users prioritize resale purchases.

**Independent Test**: Verify each product displays a profitability indicator based on purchase vs. resale prices.

### Implementation for User Story 2

- [x] T029 [US2] Implement StatsRepository (GetProductStats, RefreshStats) in internal/repository/stats.go
- [x] T030 [US2] Create materialized view refresh function in internal/database/migrations/002_product_stats_view.sql
- [x] T031 [US2] Implement profit calculator (margin EUR, margin %) in internal/analyzer/profit.go
- [x] T032 [US2] Add MSRP seed data migration in internal/database/migrations/003_seed_msrp.sql
- [x] T033 [US2] Update results embed to display ROI/margin fields in internal/discord/embeds/results.go
- [x] T034 [US2] Update /results handler to support sort by margin_percent in internal/discord/commands/results.go

**Checkpoint**: User Story 2 complete - users can see profit margins and sort by ROI

---

## Phase 5: User Story 3 - Filter and Targeted Search (Priority: P3)

**Goal**: Allow users to filter results by product category (boosters, displays, singles) to focus on their resale niche.

**Independent Test**: Apply different filters and verify only matching products are displayed.

### Implementation for User Story 3

- [x] T035 [US3] Implement /filter slash command handler in internal/discord/commands/filter.go
- [x] T036 [US3] Update StatsRepository.GetProductStats to support category filtering in internal/repository/stats.go
- [x] T037 [US3] Implement filter embed builder with category-specific colors in internal/discord/embeds/filter.go
- [x] T038 [US3] Add pagination support with Discord buttons in internal/discord/commands/results.go
- [x] T039 [US3] Implement button interaction handlers for pagination in internal/discord/bot.go

**Checkpoint**: User Story 3 complete - users can filter by category and paginate results

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [x] T040 [P] Add structured logging throughout all packages
- [x] T041 [P] Implement graceful shutdown for bot in cmd/bot/main.go
- [x] T042 [P] Add user rate limiting for Discord commands in internal/discord/ratelimit.go
- [x] T043 Add Health check endpoint for scraper in internal/scraper/ebay/scraper.go
- [x] T044 Run quickstart.md validation - test full setup flow
- [x] T045 Add seed data for additional Pokemon TCG sets (2024-2025)
- [x] T046 [P] Add integration test for rate limiting (verify 1 req/s max to eBay) in tests/integration/ratelimit_test.go
- [x] T047 [P] Add integration test for error handling (connection failures, eBay downtime simulation) in tests/integration/error_handling_test.go

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Story 1 (Phase 3)**: Depends on Foundational phase completion
- **User Story 2 (Phase 4)**: Depends on Foundational + builds on US1 repositories
- **User Story 3 (Phase 5)**: Depends on Foundational + extends US1/US2 handlers
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - Core analysis flow
- **User Story 2 (P2)**: Can start after T029 (US1 repositories must exist) - Adds profit calculations
- **User Story 3 (P3)**: Can start after T035 (US1/US2 results infrastructure) - Adds filtering

### Within Each User Story

- Types/Models before implementations
- Repositories before services/analyzers
- Backend before Discord handlers
- Core implementation before UI polish

### Parallel Opportunities

**Phase 1 (Setup)**:
- T002, T003, T004 can run in parallel

**Phase 2 (Foundational)**:
- T006, T007 can run in parallel (database + cache)
- T010, T011, T012, T013 can run in parallel (models + utilities)

**Phase 3 (User Story 1)**:
- T015, T016 can run in parallel (scraper types + parser)
- T021, T023, T026 can run in parallel (normalizer + bot setup + embeds)

**Phase 6 (Polish)**:
- T040, T041, T042 can run in parallel

---

## Parallel Example: User Story 1

```bash
# Launch scraper components together:
Task: "Define scraper types in internal/scraper/ebay/types.go"
Task: "Implement HTML parser for eBay in internal/scraper/ebay/parser.go"

# Launch independent components together:
Task: "Implement product name normalizer in internal/analyzer/normalizer.go"
Task: "Implement Discord bot setup in internal/discord/bot.go"
Task: "Implement results embed builder in internal/discord/embeds/results.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T004)
2. Complete Phase 2: Foundational (T005-T014)
3. Complete Phase 3: User Story 1 (T015-T028)
4. **STOP and VALIDATE**: Test with `/analyze` and `/results` commands
5. Deploy/demo if ready - users can discover high-volume products

### Incremental Delivery

1. Setup + Foundational -> Foundation ready
2. Add User Story 1 -> Test independently -> Deploy (MVP!)
   - Users can run analyses and see volume-sorted results
3. Add User Story 2 -> Test independently -> Deploy
   - Users can now see profit margins and sort by ROI
4. Add User Story 3 -> Test independently -> Deploy
   - Users can filter by category and paginate results
5. Each story adds value without breaking previous stories

### Suggested MVP Scope

**MVP = Phase 1 + Phase 2 + Phase 3 (User Story 1)**

This delivers:
- Working eBay scraper
- Product normalization
- Database persistence
- `/analyze` command to launch analysis
- `/results` command to view volume-sorted products

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Rate limiting (1 req/s max for eBay) is critical - implemented in T012
