# AI News Aggregator: Revised Implementation Plan (v2)

## Overview

A command-line tool that collects, filters, summarizes, and aggregates AI news articles (focused on SW Development, Architecture, and Testing) into a **locally hosted website** with weekly reports, full history archive, and **user-defined research priorities**.

**Key changes from v1:**
- Output: Static HTML website (not AsciiDoc)
- History: Auto-generated index page with all past reports
- Viewing: Local Flask server (`aicrawler serve`)
- Styling: Minimal, clean, typography-focused design
- **Research Priorities**: Editable list of topics to prioritize during collection

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        CLI Entry Point                          │
│                         (cli.py)                                │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────────┐                                               │
│  │  Research    │ ← User edits via web UI                       │
│  │  Priorities  │                                               │
│  └──────┬───────┘                                               │
│         ↓                                                       │
│  ┌──────────┐   ┌──────────┐   ┌──────────┐   ┌──────────────┐ │
│  │ Collect  │ → │  Filter  │ → │Summarize │ → │  Aggregate   │ │
│  └──────────┘   └──────────┘   └──────────┘   └──────────────┘ │
│       ↓              ↑                               ↓          │
│  ┌──────────┐        │                        ┌──────────────┐  │
│  │ SQLite   │ ───────┘                        │ Web Generator│  │
│  │   DB     │ ←───────────────────────────────└──────────────┘  │
│  └──────────┘                                        ↓          │
│       ↑                                       ┌──────────────┐  │
│       │                                       │ Flask Server │  │
│       └───────────────────────────────────────│ (port 8000)  │  │
│         (priorities CRUD)                     └──────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

### How Research Priorities Work

1. **You define priorities** via the web UI (e.g., "LLM agents for test automation", "AI code review tools")
2. **Collector uses priorities** to build targeted search queries for NewsAPI
3. **Filter boosts scores** for articles matching your priorities
4. **Aggregator highlights** priority-matched articles in the weekly report
5. **Priorities page** shows which articles matched each priority

---

## Directory Structure

```
AICrawler/
├── pyproject.toml              # Project config, dependencies
├── config.yaml                 # Sources, keywords, settings
├── src/
│   ├── __init__.py
│   ├── cli.py                  # Click-based CLI
│   ├── collector.py            # Article collection orchestrator
│   ├── feed_parser.py          # RSS/Atom feed parser
│   ├── api_client.py           # NewsAPI integration
│   ├── filter.py               # Keyword-based filtering
│   ├── summarizer.py           # LLM summarization
│   ├── aggregator.py           # Weekly digest creation
│   ├── web_generator.py        # HTML generation (NEW)
│   ├── server.py               # Flask web server (NEW)
│   ├── database.py             # SQLite operations (NEW)
│   └── priorities.py           # Research priorities logic (NEW)
├── templates/                  # Jinja2 templates (NEW)
│   ├── base.html
│   ├── index.html              # Archive listing
│   ├── report.html             # Weekly report
│   └── priorities.html         # Priorities management (NEW)
├── static/                     # CSS assets (NEW)
│   └── style.css
├── data/
│   └── articles.db             # SQLite database
└── tests/
    └── ...
```

**Note:** No separate `web/` directory needed — Flask serves templates directly and generates report HTML on-the-fly from the database.

---

## 1. Data Collection (Simplified)

### Components
- **FeedParser**: Fetch RSS/Atom feeds using `feedparser`
- **APIClient**: Fetch from NewsAPI (primary external source)
- **ArticleCollector**: Orchestrates all sources

### Simplifications from v1
- **Removed**: WebScraper with Scrapy (over-engineered)
- **Removed**: Complex NLP tagging (LLM does this during summarization)
- **Added**: Simple deduplication by URL

### Database Schema

```sql
CREATE TABLE articles (
    id INTEGER PRIMARY KEY,
    url TEXT UNIQUE NOT NULL,
    title TEXT NOT NULL,
    source TEXT,
    published_date DATE,
    content TEXT,
    summary TEXT,
    topics TEXT,              -- JSON array: ["SW Dev", "Testing"]
    week_number TEXT,         -- "2026-W04"
    relevance_score INTEGER,  -- 1-5 from LLM
    priority_matches TEXT,    -- JSON array of matched priority IDs
    collected_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    summarized_at TIMESTAMP
);

CREATE TABLE weekly_reports (
    id INTEGER PRIMARY KEY,
    week_number TEXT UNIQUE,  -- "2026-W04"
    generated_at TIMESTAMP,
    article_count INTEGER
);

-- NEW: Research Priorities
CREATE TABLE research_priorities (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL,           -- Short label: "LLM Agents"
    description TEXT,              -- Context: "AI agents for test automation, especially..."
    keywords TEXT,                 -- JSON array: ["agent", "autonomous", "testing"]
    is_active BOOLEAN DEFAULT 1,   -- Can disable without deleting
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);

-- Track which articles matched which priorities
CREATE TABLE article_priority_matches (
    article_id INTEGER REFERENCES articles(id),
    priority_id INTEGER REFERENCES research_priorities(id),
    match_score REAL,              -- How strongly it matched (0.0-1.0)
    PRIMARY KEY (article_id, priority_id)
);
```

### Implementation Steps
1. Create `database.py` with SQLite helper functions
2. Implement `feed_parser.py` to fetch and parse RSS feeds
3. Implement `api_client.py` for NewsAPI integration
4. Implement `collector.py` to orchestrate and deduplicate

---

## 2. Research Priorities (NEW)

### Purpose
Allow you to define specific topics you want the system to watch for. These go beyond general keywords — they represent your current research interests with context.

### Example Priorities

| Title | Description |
|-------|-------------|
| LLM Agents for Testing | Autonomous AI agents that can write, execute, and maintain tests. Interested in frameworks like AutoGPT for QA. |
| AI Code Review | Tools that use LLMs to review pull requests, find bugs, suggest improvements. GitHub Copilot alternatives. |
| RAG Architectures | Retrieval-augmented generation patterns for enterprise apps. Vector databases, chunking strategies. |
| Claude/Anthropic Updates | New Claude models, API changes, Claude Code updates, Anthropic research papers. |

### Web Interface

The `/priorities` page provides:

```
┌─────────────────────────────────────────────────────────────────┐
│  Research Priorities                              [+ Add New]   │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ☑ LLM Agents for Testing                           [Edit] [×]  │
│    Autonomous AI agents that can write, execute...              │
│    Last matched: 3 articles this week                           │
│                                                                 │
│  ☑ AI Code Review                                   [Edit] [×]  │
│    Tools that use LLMs to review pull requests...               │
│    Last matched: 1 article this week                            │
│                                                                 │
│  ☐ RAG Architectures (paused)                       [Edit] [×]  │
│    Retrieval-augmented generation patterns...                   │
│    Paused — will not match new articles                         │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### How Priorities Affect the Pipeline

1. **Collection Phase**
   - Active priorities generate additional NewsAPI queries
   - Example: Priority "LLM Agents" → query "LLM agents software testing"

2. **Filtering Phase**
   - Articles are scored against each priority using keyword matching
   - High-match articles get a relevance boost

3. **Summarization Phase**
   - LLM prompt includes active priorities
   - LLM tags which priorities the article relates to

4. **Report Generation**
   - Priority-matched articles shown with a highlight
   - Optional "By Priority" view groups articles by matching priority

### Implementation: `priorities.py`

```python
class PriorityManager:
    def get_active_priorities(self) -> list[Priority]
    def add_priority(self, title: str, description: str) -> Priority
    def update_priority(self, id: int, ...) -> Priority
    def delete_priority(self, id: int) -> None
    def toggle_active(self, id: int) -> None
    def get_matches_for_week(self, week: str) -> dict[int, list[Article]]
```

---

## 3. Filtering

### Approach
Simple keyword matching against title and content. No heavy NLP libraries needed.

### Configuration (config.yaml)
```yaml
keywords:
  - "AI"
  - "artificial intelligence"
  - "machine learning"
  - "software development"
  - "software architecture"
  - "testing"
  - "MLOps"
  - "LLM"
  - "GPT"
  - "Claude"
```

### Implementation
- Score articles by keyword matches
- Filter out low-relevance articles
- Pass to summarizer

---

## 3. Summarization

### Approach
Use OpenAI API (or Claude API) to:
1. Generate a 2-3 sentence summary
2. Extract topic tags (SW Dev, Architecture, Testing, General AI)
3. Rate relevance (1-5)

### Prompt Template
```
Summarize this AI news article in 2-3 sentences, focusing on its relevance
to software development, architecture, or testing.

Also provide:
- Topics: Choose from [SW Development, Architecture, Testing, General AI]
- Relevance: Rate 1-5 for a software engineering audience

Article:
{content}
```

### Implementation
- Use `openai` or `anthropic` Python SDK
- Store summary, topics, and relevance in DB
- Implement rate limiting and error handling

---

## 4. Aggregation

### Weekly Digest Creation
1. Query all summarized articles for current week
2. Group by topic
3. Sort by relevance within each topic
4. Select top articles (aim for ~20 total)

### Output Structure
```python
{
    "week_number": "2026-W04",
    "date_range": "Jan 20 - Jan 26, 2026",
    "sections": {
        "SW Development": [
            {"title": "...", "summary": "...", "url": "...", "source": "..."},
            ...
        ],
        "Architecture": [...],
        "Testing": [...],
        "General AI": [...]
    },
    "total_articles": 18
}
```

---

## 5. Web Generation (NEW)

### Components
- **Jinja2 Templates**: Clean, semantic HTML
- **Minimal CSS**: Typography-focused, dark mode support
- **WebGenerator class**: Renders templates, manages output

### Templates

#### base.html
```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{% block title %}AI News Digest{% endblock %}</title>
    <link rel="stylesheet" href="{{ static_url }}style.css">
</head>
<body>
    <header>
        <nav>
            <a href="{{ root_url }}index.html">← Archive</a>
        </nav>
        <h1>AI News Digest</h1>
    </header>
    <main>
        {% block content %}{% endblock %}
    </main>
    <footer>
        <p>Generated with AICrawler</p>
    </footer>
</body>
</html>
```

#### index.html (Archive)
- Lists all weekly reports by date
- Shows article count per week
- Most recent at top

#### report.html (Weekly Report)
- Date range header
- Sections by topic
- Each article: title (linked), summary, source

### CSS Design Principles
- System font stack (fast, native feel)
- Max-width for readability (~65ch)
- Subtle colors, high contrast
- Dark mode via `prefers-color-scheme`
- No JavaScript required

### Implementation Steps
1. Create Jinja2 templates in `templates/`
2. Create `static/style.css`
3. Implement `web_generator.py`:
   - `generate_report(week_data)` → creates weekly HTML
   - `generate_index()` → creates archive page
   - `copy_static_assets()` → copies CSS to web/

---

## 6. Local Server (NEW)

### Flask Web Application

Using Flask instead of `http.server` enables the interactive priorities management.

```python
# src/server.py
from flask import Flask, render_template, request, redirect, url_for
from .database import get_db
from .priorities import PriorityManager

app = Flask(__name__,
            template_folder="../templates",
            static_folder="../static")

# --- Read-only pages (reports) ---

@app.route("/")
def index():
    """Archive page listing all weekly reports."""
    reports = get_db().get_all_reports()
    return render_template("index.html", reports=reports)

@app.route("/report/<week>")
def report(week: str):
    """Individual weekly report."""
    data = get_db().get_report_data(week)
    return render_template("report.html", **data)

# --- Interactive pages (priorities) ---

@app.route("/priorities")
def priorities():
    """List and manage research priorities."""
    pm = PriorityManager()
    items = pm.get_all_with_stats()
    return render_template("priorities.html", priorities=items)

@app.route("/priorities/add", methods=["POST"])
def add_priority():
    """Add a new research priority."""
    pm = PriorityManager()
    pm.add_priority(
        title=request.form["title"],
        description=request.form["description"]
    )
    return redirect(url_for("priorities"))

@app.route("/priorities/<int:id>/edit", methods=["POST"])
def edit_priority(id: int):
    """Update an existing priority."""
    pm = PriorityManager()
    pm.update_priority(
        id=id,
        title=request.form["title"],
        description=request.form["description"]
    )
    return redirect(url_for("priorities"))

@app.route("/priorities/<int:id>/toggle", methods=["POST"])
def toggle_priority(id: int):
    """Enable/disable a priority."""
    pm = PriorityManager()
    pm.toggle_active(id)
    return redirect(url_for("priorities"))

@app.route("/priorities/<int:id>/delete", methods=["POST"])
def delete_priority(id: int):
    """Remove a priority."""
    pm = PriorityManager()
    pm.delete_priority(id)
    return redirect(url_for("priorities"))

def serve(port: int = 8000):
    """Start the local server."""
    print(f"Starting server at http://localhost:{port}")
    app.run(host="127.0.0.1", port=port, debug=False)
```

### Routes Summary

| Route | Method | Purpose |
|-------|--------|---------|
| `/` | GET | Archive page (list of weekly reports) |
| `/report/<week>` | GET | View a specific weekly report |
| `/priorities` | GET | View and manage research priorities |
| `/priorities/add` | POST | Add new priority |
| `/priorities/<id>/edit` | POST | Update priority |
| `/priorities/<id>/toggle` | POST | Enable/disable priority |
| `/priorities/<id>/delete` | POST | Remove priority |

### Usage

```bash
aicrawler serve              # Start on port 8000
aicrawler serve --port 3000  # Custom port
```

---

## 7. CLI Interface

### Commands

```bash
# Full pipeline (collect → filter → summarize → aggregate)
aicrawler run

# Individual steps
aicrawler collect            # Fetch new articles
aicrawler summarize          # Summarize unsummarized articles

# Web server (serves reports AND priorities UI)
aicrawler serve              # Start on port 8000
aicrawler serve --port 3000  # Custom port

# Priorities (CLI alternative to web UI)
aicrawler priorities list              # Show all priorities
aicrawler priorities add "Title" "Description..."
aicrawler priorities remove 3          # Remove by ID
aicrawler priorities toggle 2          # Enable/disable

# Utilities
aicrawler status             # Show DB stats, last run, etc.
aicrawler config             # Validate and display config
```

### Implementation
- Use `click` for CLI framework
- Add `--verbose` flag for debugging
- Add `--dry-run` for testing

---

## 8. Configuration

### config.yaml
```yaml
# Data sources
sources:
  feeds:
    - url: "https://ai.googleblog.com/atom.xml"
      name: "Google AI Blog"
    - url: "https://openai.com/blog/rss"
      name: "OpenAI Blog"
    - url: "https://arxiv.org/rss/cs.AI"
      name: "arXiv AI"
  apis:
    newsapi:
      enabled: true
      api_key_env: "NEWSAPI_KEY"
      query: "artificial intelligence software development"

# Filtering
keywords:
  - "AI"
  - "artificial intelligence"
  - "machine learning"
  - "software development"
  - "software architecture"
  - "testing"
  - "MLOps"

# Summarization
summarization:
  provider: "openai"          # or "anthropic"
  model: "gpt-4o-mini"        # Cost-effective for summaries
  api_key_env: "OPENAI_API_KEY"

# Output
output:
  web_dir: "./web"
  data_dir: "./data"

# Server
server:
  port: 8000

# Scheduling (external - use cron)
# Example crontab: 0 18 * * 5 cd /path/to/AICrawler && aicrawler run
```

---

## 9. Scheduling

### Approach
Use system cron (macOS/Linux) instead of Python scheduler.

### Setup (macOS)
```bash
# Edit crontab
crontab -e

# Add line (runs every Friday at 6 PM)
0 18 * * 5 cd /path/to/AICrawler && /path/to/venv/bin/aicrawler run >> /path/to/AICrawler/logs/cron.log 2>&1
```

### Why cron over APScheduler
- Simpler, no daemon to manage
- Survives reboots
- Standard Unix tooling
- Logs are straightforward

---

## 10. Dependencies

### pyproject.toml

```toml
[project]
name = "aicrawler"
version = "0.1.0"
requires-python = ">=3.11"
dependencies = [
    "click>=8.0",
    "feedparser>=6.0",
    "flask>=3.0",
    "httpx>=0.27",
    "openai>=1.0",
    "pyyaml>=6.0",
]

[project.optional-dependencies]
dev = [
    "pytest>=8.0",
    "ruff>=0.4",
]

[project.scripts]
aicrawler = "src.cli:main"
```

**Note:** Flask includes Jinja2, so no separate jinja2 dependency needed.

### Removed from v1

- `newspaper3k` (unreliable, slow)
- `scrapy` (over-engineered)
- `spacy` (heavy NLP not needed)
- `APScheduler` (use cron)

---

## 11. Implementation Order

### Phase 1: Core Infrastructure

1. Set up `pyproject.toml` with `uv`
2. Implement `database.py` (SQLite helpers, all tables)
3. Implement `feed_parser.py`
4. Basic `cli.py` with `collect` command
5. **Milestone**: Can collect articles to DB

### Phase 2: Processing Pipeline

6. Implement `filter.py` (keyword matching + priority boosting)
7. Implement `summarizer.py` (OpenAI integration)
8. Implement `aggregator.py` (weekly grouping)
9. **Milestone**: Can process articles through full pipeline

### Phase 3: Web Server & Templates

10. Create base Flask app (`server.py`)
11. Create `templates/base.html` and CSS
12. Create `templates/index.html` (archive)
13. Create `templates/report.html` (weekly report)
14. **Milestone**: Can serve reports via Flask

### Phase 4: Research Priorities

15. Implement `priorities.py` (CRUD operations)
16. Create `templates/priorities.html`
17. Add priority routes to Flask
18. Integrate priorities into filter/summarizer
19. **Milestone**: Can manage priorities via web UI

### Phase 5: Polish

20. Add `run` command (full pipeline)
21. Add `status` command
22. Error handling and logging
23. **Milestone**: Production-ready

---

## 12. Example Workflow

```bash
# Initial setup
cd AICrawler
uv venv
uv pip install -e .
export OPENAI_API_KEY="sk-..."
export NEWSAPI_KEY="..."

# Start the server
aicrawler serve
# Open http://localhost:8000

# Step 1: Define your research priorities (via web UI)
# Go to http://localhost:8000/priorities
# Add topics like:
#   - "LLM Agents for Testing"
#   - "AI Code Review Tools"
#   - "Claude/Anthropic Updates"

# Step 2: Run the collection pipeline
aicrawler run

# Step 3: View your weekly report
# Go to http://localhost:8000
# Click on the latest week
# Articles matching your priorities are highlighted

# Weekly automation (via cron)
# Every Friday at 6 PM: aicrawler run
# Then just open the browser to see results

# Check status anytime
aicrawler status
```

---

## Summary of Changes

| Aspect | Original Plan | Revised Plan v2 |
|--------|--------------|-----------------|
| Output | AsciiDoc + PDF | Flask web app |
| History | Manual file browsing | Web archive page |
| Customization | Edit config.yaml | Web UI for priorities |
| Web scraping | Scrapy | Removed (RSS + API sufficient) |
| NLP | spaCy/HuggingFace | Keyword matching + LLM |
| Scheduler | APScheduler | System cron |
| Server | None | Flask (lightweight) |

### New Features in v2

- **Research Priorities**: Define topics you care about, managed via web UI
- **Priority Matching**: Articles are tagged with which priorities they match
- **History Archive**: Browse all past weekly reports from one page
- **Live Web App**: No static file generation — Flask serves everything dynamically

---

## Next Steps

1. Review this plan
2. Approve or request modifications
3. Begin Phase 1 implementation
