# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

AICrawler is a Python CLI tool that generates daily narrative briefings about practical AI developments. It collects articles from RSS feeds and NewsAPI, triages them with LLMs, clusters related articles into storylines using sentence-transformer embeddings, synthesizes per-storyline narratives, and composes a cohesive briefing served via a Flask web app. Supports smart catch-up: if days are missed, one combined briefing covers the gap.

## Commands

```bash
# Setup
uv venv && uv pip install -e ".[dev]"

# Linting
uv run ruff check src/
uv run ruff check --fix src/

# Tests
uv run pytest tests/ -v
uv run pytest tests/test_database.py -v           # Single file
uv run pytest tests/test_database.py::test_name    # Single test

# Run the app
aicrawler run                     # Full 6-step pipeline (daily, with catch-up)
aicrawler run --days-back 3       # Override lookback window (e.g. bootstrap fresh DB)
aicrawler run --dry-run           # Preview without executing
aicrawler collect                 # Fetch articles only
aicrawler serve                   # Flask on localhost:8000
aicrawler status                  # Database stats
aicrawler priorities list         # Manage research priorities
aicrawler priorities add "Topic"  # Add a priority
```

## Architecture

### Data Pipeline

```
RSS Feeds + NewsAPI
    ↓ collect (feed_parser.py, api_client.py → collector.py)
SQLite DB (database.py)
    ↓ fetch content (content_fetcher.py: httpx + trafilatura)
    ↓ triage (triage.py: LLM → relevant/skip, key_points, practical_score)
    ↓ cluster (clusterer.py: sentence-transformer embeddings + agglomerative clustering → storylines)
    ↓ synthesize (synthesizer.py: LLM per storyline → narrative)
    ↓ compose (composer.py: LLM → full briefing with TL;DR)
Flask Web App (server.py → Jinja2 templates)
```

### Module Responsibilities

| Module | Purpose |
|--------|---------|
| `src/llm.py` | LLM provider abstraction (LLMProvider ABC, OllamaProvider, OpenAIProvider, `create_provider`, `parse_json_response`) |
| `src/collector.py` | Collects articles from RSS feeds and NewsAPI, inserts into DB with `days_back` parameter |
| `src/content_fetcher.py` | Fetches full article text via httpx + trafilatura for feeds with empty RSS content |
| `src/triage.py` | Per-article LLM triage: verdict (relevant/skip), article_type, key_points, practical_score |
| `src/clusterer.py` | Sentence-transformer embeddings + scipy agglomerative clustering into storylines |
| `src/synthesizer.py` | Per-storyline LLM narrative; "Briefly Noted" gets bullet-point treatment (no LLM) |
| `src/composer.py` | Assembles full briefing with LLM-generated TL;DR |
| `src/database.py` | SQLite schema, dataclasses, CRUD operations |
| `src/server.py` | Flask routes: archive, briefing view, priorities CRUD |
| `src/cli.py` | Click CLI: `run` (with catch-up), `collect`, `serve`, `status`, `priorities` |
| `src/feed_parser.py` | RSS/Atom feed parsing |
| `src/api_client.py` | NewsAPI client |

### LLM Provider Abstraction

`src/llm.py` defines an `LLMProvider` ABC with `generate(prompt, system_prompt)` and `is_configured()`. Concrete providers: `OllamaProvider` (default, local via httpx to `localhost:11434`) and `OpenAIProvider`. All pipeline modules that need LLM (triage, synthesizer, composer) import from `src/llm.py`. Default model: `qwen2.5:7b` via Ollama.

The `parse_json_response` helper extracts JSON from LLM output, handling markdown code fences (`\`\`\`json ... \`\`\``).

### Database

SQLite via `database.py`. Key tables:

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

Dataclasses: `Article`, `ArticleTriage`, `Storyline`, `StorylineNarrative`, `Briefing`, `ResearchPriority`, `RunReport`. Access via `get_db()` singleton; `reset_db()` for tests. JSON-serialized columns: `key_points`, `keywords`, `source_references`.

Period IDs use `YYYY-MM-DD` format for single days or `YYYY-MM-DD..YYYY-MM-DD` for date ranges. Utility functions: `get_today()`, `make_period_id(start, end)`, `format_period_display(period_id)`.

### Web Routes

| Route | Template | Purpose |
|-------|----------|---------|
| `GET /` | index.html | Archive listing (newest first) |
| `GET /briefing/<period_id>` | briefing.html | Briefing with TL;DR + narratives |
| `GET /priorities` | priorities.html | Research priority CRUD |
| `POST /priorities/add` | — | Add priority |
| `POST /priorities/<id>/toggle` | — | Toggle active state |
| `POST /priorities/<id>/delete` | — | Delete priority |

Briefing body is stored as markdown in DB, rendered to HTML at serve-time via Python `markdown` library (Jinja2 `|markdown` filter). Period IDs are formatted for display via `|format_period` filter.

### Research Priorities

User-defined topics (e.g., "LLM Agents for Testing") that: generate additional NewsAPI queries during collection and get a relevance boost during triage. Managed via `/priorities` web UI or `aicrawler priorities` CLI.

### CLI Structure

Click-based (`cli.py`). Top-level group with `--verbose` and `--config` options. Config loaded from `config.yaml` into `ctx.obj["config"]`. The `run` command auto-detects catch-up scenarios via `db.get_last_run_date()`, computes the appropriate `period_id` and `days_back`, and confirms with the user if >5 days missed. The `--days-back N` option overrides auto-detection, useful for bootstrapping a fresh database or testing. All pipeline modules accept optional `db` parameter for testability.

## Key Conventions

- Ruff: line-length 100, rules E/F/I/N/W/UP, E501 ignored
- Type hints on all function signatures (`str | None` syntax, not `Optional`)
- Conventional commits: `feat(triage): add article_type`, `fix(clusterer): threshold`
- Config in `config.yaml`, secrets in `.env` (OPENAI_API_KEY, NEWSAPI_KEY)
- Templates: semantic HTML + CSS only, no JS frameworks
- CSS: dark mode via `prefers-color-scheme`, max-width ~65ch

## Test Patterns

Tests use a shared `temp_db` fixture from `conftest.py`:

```python
@pytest.fixture
def temp_db():
    reset_db()
    with tempfile.TemporaryDirectory() as tmpdir:
        db = Database(f"{tmpdir}/test.db")
        yield db
    reset_db()
```

LLM-dependent tests mock the provider with `unittest.mock.MagicMock`. Test files: `test_database.py` (19 tests), `test_llm.py` (6), `test_triage.py` (5), `test_clusterer.py` (4), `test_synthesizer.py` (3), `test_composer.py` (3).

## Configuration

`config.yaml` drives source feeds, API settings, keywords, LLM provider choice (ollama/openai), model selection, data directory, and server port. Default: Ollama at `http://localhost:11434` with model `qwen2.5:7b`.
