# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

AICrawler is a Go CLI tool that generates daily narrative briefings about practical AI developments. It collects articles from RSS feeds and NewsAPI, triages them with LLMs, clusters related articles into storylines using Ollama embeddings + Ward's agglomerative clustering, synthesizes per-storyline narratives, and composes a cohesive briefing served via a built-in web server. Supports smart catch-up: if days are missed, one combined briefing covers the gap. Distributed as a single static binary via Homebrew.

## Commands

```bash
# Build
make build                        # CGO_ENABLED=0 go build
make test                         # go test ./...
make lint                         # go vet ./...

# Install (development)
go install ./cmd/aicrawler

# Install (Homebrew)
brew install TobiSchelling/tap/aicrawler
aicrawler init                    # Create ~/.config/aicrawler/config.yaml

# Run the app
aicrawler init                    # Initialize config (first-time setup)
aicrawler run                     # Full 6-step pipeline (daily, with catch-up)
aicrawler run --days-back 3       # Override lookback window
aicrawler run --dry-run           # Preview without executing
aicrawler collect                 # Fetch articles only
aicrawler serve                   # Web server on localhost:8000
aicrawler status                  # Database stats
aicrawler priorities list         # Manage research priorities
aicrawler priorities add "Topic"  # Add a priority

# Test specific packages
go test ./internal/database/... -v
go test ./internal/llm/... -v
go test ./internal/cluster/... -v
```

## Architecture

### Data Pipeline

```
RSS Feeds + NewsAPI
    ↓ collect (collect/feed.go, collect/newsapi.go → collect/collect.go)
SQLite DB (database/)
    ↓ fetch content (fetch/fetch.go: net/http + go-readability)
    ↓ triage (triage/triage.go: LLM → relevant/skip, key_points, practical_score)
    ↓ cluster (cluster/: Ollama embeddings + Ward's linkage → storylines)
    ↓ synthesize (synthesize/synthesize.go: LLM per storyline → narrative)
    ↓ compose (compose/compose.go: LLM → full briefing with TL;DR)
Built-in Web Server (server/server.go → Go html/template)
Pipeline Orchestrator (pipeline/pipeline.go)
```

### Package Responsibilities

| Package | Purpose |
|---------|---------|
| `internal/llm` | LLM provider interface (`Provider`, `Embedder`), OllamaProvider, OpenAIProvider, `CreateProvider`, `ParseJSONResponse` |
| `internal/collect` | Collects articles from RSS feeds (gofeed) and NewsAPI, inserts into DB with `daysBack` parameter |
| `internal/fetch` | Fetches full article text via net/http + go-readability for feeds with empty RSS content |
| `internal/triage` | Per-article LLM triage: verdict (relevant/skip), article_type, key_points, practical_score |
| `internal/cluster` | Ollama embeddings + Ward's agglomerative clustering (from-scratch implementation) into storylines |
| `internal/synthesize` | Per-storyline LLM narrative; "Briefly Noted" gets bullet-point treatment (no LLM) |
| `internal/compose` | Assembles full briefing with LLM-generated TL;DR |
| `internal/database` | SQLite schema (modernc.org/sqlite, pure Go), model structs, CRUD operations, period utilities |
| `internal/config` | Config struct + YAML loading (gopkg.in/yaml.v3), XDG path resolution, embedded default.yaml |
| `internal/server` | net/http handlers + routes, embedded templates (html/template) + CSS, goldmark markdown rendering |
| `internal/pipeline` | 6-step orchestrator with StepResult pattern, dry-run support |
| `cmd/aicrawler` | Cobra CLI: `run` (catch-up detection, --days-back, --dry-run), `collect`, `serve`, `status`, `priorities`, `init` |

### LLM Provider Abstraction

`internal/llm/llm.go` defines a `Provider` interface with `Generate(ctx, prompt, maxTokens)` and `IsConfigured()`, plus an `Embedder` interface with `Embed(ctx, texts)`. Concrete providers: `OllamaProvider` (default, local via HTTP to `localhost:11434`) and `OpenAIProvider`. All pipeline modules that need LLM receive a `Provider` via constructor injection. Default model: `qwen2.5:7b` via Ollama.

`ParseJSONResponse` extracts JSON from LLM output, handling markdown code fences.

### Database

SQLite via `modernc.org/sqlite` (pure Go, no CGO). Key tables:

| Table | Purpose |
|-------|---------|
| `articles` | Collected articles with `content_fetched` flag and `period_id` |
| `article_triage` | LLM triage results: verdict, article_type, key_points (JSON), practical_score |
| `storylines` | Clusters of related articles per period |
| `storyline_articles` | Junction table: storyline ↔ article |
| `storyline_narratives` | LLM-generated narrative per storyline with source_references (JSON) |
| `briefings` | Final composed briefing: tldr + body_markdown |
| `research_priorities` | User-defined topics with keywords (JSON) |
| `run_reports` | Metadata for pipeline runs |

Model structs: `Article`, `ArticleTriage`, `Storyline`, `StorylineNarrative`, `Briefing`, `ResearchPriority`, `RunReport`. No global singleton — `*database.DB` created in `main.go`, passed down. Each test creates its own DB via `t.TempDir()`.

Period IDs use `YYYY-MM-DD` format for single days or `YYYY-MM-DD..YYYY-MM-DD` for date ranges. Utility functions: `GetToday()`, `MakePeriodID(start, end)`, `FormatPeriodDisplay(periodID)`.

### Web Routes

| Route | Template | Purpose |
|-------|----------|---------|
| `GET /` | index.html | Archive listing (newest first) |
| `GET /briefing/{period_id}` | briefing.html | Briefing with TL;DR + narratives |
| `GET /priorities` | priorities.html | Research priority CRUD |
| `POST /priorities/add` | — | Add priority |
| `POST /priorities/{id}/toggle` | — | Toggle active state |
| `POST /priorities/{id}/delete` | — | Delete priority |

Briefing body is stored as markdown in DB, rendered to HTML at serve-time via goldmark. Period IDs are formatted for display via `formatPeriod` template function.

### Research Priorities

User-defined topics (e.g., "LLM Agents for Testing") that: generate additional NewsAPI queries during collection and get a relevance boost during triage. Managed via `/priorities` web UI or `aicrawler priorities` CLI.

### CLI Structure

Cobra-based (`cmd/aicrawler/main.go`). Root command with `--verbose` and `--config` flags. Config resolution: `--config` flag > `~/.config/aicrawler/config.yaml` > `./config.yaml`. Data directory: `config.output.data_dir` > `~/.local/share/aicrawler/`. The `init` command writes the embedded `default.yaml` to `~/.config/aicrawler/config.yaml`. The `run` command auto-detects catch-up scenarios via `db.GetLastRunDate()`, computes the appropriate `periodID` and `daysBack`, and confirms with the user if >5 days missed. The `--days-back N` option overrides auto-detection.

## Key Conventions

- Go 1.25+, pure Go (CGO_ENABLED=0)
- `go vet` for linting; `go test ./...` for tests
- Conventional commits: `feat(triage): add article_type`, `fix(clusterer): threshold`
- Config: `~/.config/aicrawler/config.yaml` (XDG) or `./config.yaml` (dev); secrets in `.env`
- Data: `~/.local/share/aicrawler/` (XDG default) or `config.output.data_dir`
- Templates and static CSS embedded via `//go:embed` in `internal/server/`
- Default config YAML embedded via `//go:embed` in `internal/config/`
- Templates: semantic HTML + CSS only, no JS frameworks
- Embeddings via Ollama `embedding_model` (default: `nomic-embed-text`)
- CSS: dark mode via `prefers-color-scheme`, max-width ~65ch
- Interface-based testing (no mock library): `llm.Provider` and `llm.Embedder` interfaces

## Test Patterns

Tests create per-test databases using `t.TempDir()`:

```go
func openTestDB(t *testing.T) *database.DB {
    t.Helper()
    db, err := database.Open(filepath.Join(t.TempDir(), "test.db"))
    if err != nil {
        t.Fatalf("failed to open test db: %v", err)
    }
    t.Cleanup(func() { db.Close() })
    return db
}
```

LLM-dependent tests use simple struct implementations of the `Provider`/`Embedder` interfaces. Test files: `database_test.go` (19), `llm_test.go` (6), `triage_test.go` (5), `cluster_test.go` (4), `ward_test.go` (5), `synthesize_test.go` (3), `compose_test.go` (3), `server_test.go` (3), `config_test.go` (4).

## Configuration

`config.yaml` drives source feeds, API settings, keywords, LLM provider choice (ollama/openai), model selection, embedding model, data directory, and server port. Default: Ollama at `http://localhost:11434` with model `qwen2.5:7b` and embedding model `nomic-embed-text`.

Config resolution order: `--config` flag > `~/.config/aicrawler/config.yaml` > `./config.yaml`. Run `aicrawler init` to create the XDG config from the bundled default. The `output.data_dir` setting overrides the default database location.

## Distribution

Single static binary via GoReleaser. Tag `v*` triggers `.github/workflows/release.yml` which builds for darwin/linux × amd64/arm64. GoReleaser auto-updates the Homebrew tap formula.
