# AICrawler

A CLI tool that generates weekly narrative briefings about practical AI developments. Collects articles from RSS feeds and NewsAPI, triages them with LLMs, clusters related articles into storylines, and composes cohesive briefings served via a local web app.

## Features

- **6-Step Pipeline**: collect → fetch content → triage → cluster → synthesize → compose
- **LLM Triage**: Each article assessed for relevance, type, and practical value
- **Storyline Clustering**: Related articles grouped via sentence-transformer embeddings
- **Narrative Synthesis**: LLM weaves each storyline into a readable narrative section
- **Weekly Briefing**: TL;DR bullets + full narrative body, stored as markdown
- **Research Priorities**: Define topics for boosted collection and triage relevance
- **Local Web UI**: Flask-based reading interface at `http://localhost:8000`

## Quick Start

```bash
# Install with uv
uv venv
uv pip install -e .

# Set up environment variables
cp .env.example .env
# Edit .env with your API keys

# Run the full pipeline
aicrawler run

# Start the web server
aicrawler serve
# Open http://localhost:8000 in your browser
```

## Usage

### Full Pipeline

```bash
# Run complete pipeline: collect → fetch → triage → cluster → synthesize → compose
aicrawler run

# Preview what would happen without executing
aicrawler run --dry-run
```

### Individual Commands

```bash
# Collect articles from feeds and APIs
aicrawler collect

# Start web server
aicrawler serve
aicrawler serve --port 3000  # Custom port

# Show database status
aicrawler status
```

### Managing Priorities

Via CLI:

```bash
aicrawler priorities list
aicrawler priorities add "LLM Agents" "Autonomous AI for testing"
aicrawler priorities toggle 1
aicrawler priorities remove 1
```

Via Web UI:

- Navigate to `http://localhost:8000/priorities`
- Add, edit, pause, or delete priorities

## Configuration

Edit `config.yaml` to customize:

- **sources**: RSS feeds and API endpoints
- **keywords**: Terms for filtering articles
- **summarization**: LLM provider and model settings

## LLM Configuration

The pipeline uses **Ollama by default** (local, free), with OpenAI as a fallback.

### Using Ollama (Default)

```bash
# Install Ollama from https://ollama.com
# Pull the default model
ollama pull qwen2.5:7b

# Ensure Ollama is running (it starts automatically on macOS)
ollama serve
```

### Using OpenAI

To use OpenAI instead, edit `config.yaml`:

```yaml
summarization:
  provider: "openai"
  openai_model: "gpt-4o-mini"
```

Then set your API key in `.env`.

## Environment Variables

| Variable         | Description                            |
|------------------|----------------------------------------|
| `OPENAI_API_KEY` | Required only if using OpenAI provider |
| `NEWSAPI_KEY`    | Optional, for NewsAPI integration      |

## Project Structure

```text
AICrawler/
├── src/
│   ├── cli.py              # Click CLI interface
│   ├── collector.py        # Article collection from RSS + NewsAPI
│   ├── content_fetcher.py  # Full-text extraction (httpx + trafilatura)
│   ├── triage.py           # Per-article LLM triage
│   ├── clusterer.py        # Embedding-based storyline clustering
│   ├── synthesizer.py      # Per-storyline LLM narrative
│   ├── composer.py         # Weekly briefing composition
│   ├── llm.py              # LLM provider abstraction
│   ├── database.py         # SQLite schema and operations
│   ├── server.py           # Flask web server
│   ├── feed_parser.py      # RSS/Atom parsing
│   └── api_client.py       # NewsAPI client
├── templates/              # Jinja2 HTML templates
├── static/                 # CSS styles
├── tests/                  # pytest test suite
├── config.yaml             # Configuration
└── data/
    └── articles.db         # SQLite database
```

## Scheduling

For automated weekly runs, add to crontab:

```bash
# Run every Friday at 6 PM
0 18 * * 5 cd /path/to/AICrawler && /path/to/venv/bin/aicrawler run
```

## License

MIT
