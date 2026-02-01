# AI News Aggregator: Inspiration Engine (v3)

---

## Implementation Progress

> **Last Updated:** 2026-01-26
>
> **Current Phase:** ALL PHASES COMPLETE
>
> **Status:** Inspiration Engine fully implemented

| Phase | Status | Completion Date | Notes |
| ----- | ------ | --------------- | ----- |
| 1. Feedback Intelligence | **COMPLETE** | 2026-01-26 | All deliverables implemented |
| 2. Trend Detection | **COMPLETE** | 2026-01-26 | All deliverables implemented |
| 3. Source Discovery | **COMPLETE** | 2026-01-26 | All deliverables implemented |
| 4. Recommendations | **COMPLETE** | 2026-01-26 | All deliverables implemented |
| 5. Story Tracking | **COMPLETE** | 2026-01-26 | All deliverables implemented |
| 6. Dashboard | **COMPLETE** | 2026-01-26 | All deliverables implemented |

### Phase 6 Deliverables (COMPLETE)

- [x] `src/dashboard.py` module with DashboardBuilder
- [x] `/dashboard` route in server.py
- [x] `templates/dashboard.html` template
- [x] Navigation updated with Dashboard link
- [x] CSS styling for dashboard visualizations (bars, tables, cards)

### Files Modified/Created in Phase 6

**New Files:**

- `src/dashboard.py` - Intelligence dashboard data aggregation
- `templates/dashboard.html` - Dashboard template with stats, focus areas, blind spots, predictions

**Modified Files:**

- `src/server.py` - Added `/dashboard` route
- `templates/base.html` - Added Dashboard link to navigation
- `static/style.css` - Added ~250 lines of dashboard styles

### Phase 5 Deliverables (COMPLETE)

- [x] Migration v7: Story tables (`stories`, `story_articles`, `story_events`, `article_embeddings`)
- [x] New dependency: sentence-transformers for semantic embeddings
- [x] `src/story_tracker.py` module with clustering and similarity matching
- [x] Story database operations in `database.py`
- [x] Aggregator integration (WeekData now includes stories)
- [x] Report template "Ongoing Stories" section with timeline visualization
- [x] CSS styling for story cards and timeline

### Files Modified/Created in Phase 5

**New Files:**

- `src/story_tracker.py` - Story detection and tracking with semantic embeddings

**Modified Files:**

- `src/database.py` - Added migration v7, Story/StoryEvent dataclasses, story DB operations
- `src/aggregator.py` - Added StorySummary dataclass, stories integration
- `templates/report.html` - Added "Ongoing Stories" section
- `static/style.css` - Added story timeline styles
- `pyproject.toml` - Added sentence-transformers and numpy dependencies

### Phase 4 Deliverables (COMPLETE)

- [x] `src/recommender.py` module with scoring formula and exploration/exploitation split
- [x] Aggregator integration (WeekData now includes recommendations)
- [x] Report template "Recommended For You" section
- [x] Report template "Explore Something Different" serendipity section
- [x] CSS styling for recommendation cards

### Files Modified/Created in Phase 4

**New Files:**

- `src/recommender.py` - Personalized recommendation engine

**Modified Files:**

- `src/aggregator.py` - Added Recommendations import, WeekData.recommendations field, recommender integration
- `templates/report.html` - Added "Recommended For You" and "Explore" sections
- `static/style.css` - Added recommendation card styles

### Phase 3 Deliverables (COMPLETE)

- [x] Migration v6: Source tables (`discovered_sources`, `source_evaluations`)
- [x] New dependency: feedfinder2 for RSS feed discovery
- [x] `src/source_discovery.py` module with discovery and lifecycle management
- [x] CLI sources commands (`sources list`, `discover`, `add`, `promote`, `disable`, `reject`, `stats`)
- [x] `/sources` page and template with source management UI
- [x] Collector integration with dynamic sources
- [x] Navigation updated with Sources link
- [x] CSS styling for sources page

### Files Modified/Created in Phase 3

**New Files:**

- `src/source_discovery.py` - Source discovery logic (HN, Reddit, citations)
- `templates/sources.html` - Source management page

**Modified Files:**

- `src/database.py` - Added migration v6 and source DB operations
- `src/server.py` - Added source management routes
- `src/cli.py` - Added sources command group
- `src/collector.py` - Integrated dynamic sources
- `templates/base.html` - Added Sources nav link
- `static/style.css` - Added source page styles
- `pyproject.toml` - Added feedfinder2 dependency

### Phase 2 Deliverables (COMPLETE)

- [x] Migration v5: Trend tables (`weekly_term_frequencies`, `weekly_trends`)
- [x] `src/trend_detector.py` module with term extraction and velocity analysis
- [x] Pipeline integration (Step 4/6 in `aicrawler run`)
- [x] Report template trends section with emerging/accelerating/cooling display
- [x] CLI trends commands (`trends show`, `extract`, `detect`)
- [x] CSS styling for trends section

### Files Modified/Created in Phase 2

**New Files:**

- `src/trend_detector.py` - Trend detection logic (term extraction, velocity calculation)

**Modified Files:**

- `src/database.py` - Added migration v5 and trend DB operations
- `src/server.py` - Added trend data to report route
- `src/cli.py` - Added trends command group, integrated into pipeline
- `templates/report.html` - Added trends section
- `static/style.css` - Added trend section styles

### Phase 1 Deliverables (COMPLETE)

- [x] Migration v4: Preference tables (`source_preferences`, `topic_preferences`, `keyword_preferences`)
- [x] `src/preference_learner.py` module
- [x] Updated feedback routes with learning integration
- [x] `/preferences` page and template
- [x] CLI preference commands (`preferences show`, `reset`, `rebuild`)
- [x] Navigation updated with Preferences link
- [x] CSS styling for preferences page

### Files Modified/Created in Phase 1

**New Files:**

- `src/preference_learner.py` - Preference learning logic

**Modified Files:**

- `src/database.py` - Added migration v4 and preference DB operations
- `src/server.py` - Added preference routes and learning integration
- `src/cli.py` - Added preferences command group
- `templates/base.html` - Added Preferences nav link
- `templates/preferences.html` - New preferences page
- `static/style.css` - Added preference page styles

---

## Overview

Transform the existing news aggregator from a **passive weekly digest** into an **active inspiration engine** that learns user preferences, detects trends, discovers sources, and surfaces serendipitous content.

**Key changes from v2:**
- **Feedback Intelligence**: Learn from user votes to build preference models
- **Trend Detection**: Surface emerging topics and velocity changes
- **Dynamic Sources**: Self-growing, quality-scored source list
- **Personalization**: "For You" recommendations with serendipity injection
- **Story Tracking**: Cross-week narrative threads
- **Intelligence Dashboard**: Meta-view of information diet

**Design Principles:**
- Inspiration over search (divergent, not convergent)
- Curation over crawling (bounded domain, not Google)
- Serendipity by design (avoid filter bubbles)
- Incremental value (each phase delivers standalone benefits)

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         Inspiration Engine v3                           │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  DISCOVERY LAYER                                                        │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐                  │
│  │   RSS Feeds  │  │   NewsAPI    │  │  HN/Reddit   │ ← Phase 3        │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘                  │
│         │                 │                 │                           │
│         └────────────┬────┴─────────────────┘                           │
│                      ▼                                                  │
│  ┌───────────────────────────────────────────────────────────────────┐ │
│  │                    Source Discovery & Scoring                      │ │
│  │    Candidates → Quality Score → Probation → Active/Disabled       │ │
│  └───────────────────────────────────────────────────────────────────┘ │
│                      │                                                  │
│  PROCESSING LAYER    ▼                                                  │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐               │
│  │ Collect  │→ │Summarize │→ │  Theme   │→ │Narrative │ (existing)    │
│  └──────────┘  └──────────┘  │ Detect   │  │Synthestic│               │
│                              └──────────┘  └──────────┘               │
│                      │                                                  │
│  INTELLIGENCE LAYER  ▼                                                  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐                  │
│  │   Trend      │  │  Preference  │  │    Story     │                  │
│  │  Detection   │  │    Model     │  │   Tracking   │                  │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘                  │
│         │                 │                 │                           │
│         └────────────┬────┴─────────────────┘                           │
│                      ▼                                                  │
│  PRESENTATION LAYER                                                     │
│  ┌───────────────────────────────────────────────────────────────────┐ │
│  │                        Weekly Report                               │ │
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐  │ │
│  │  │  Trending   │ │ For You     │ │  Stories    │ │   Topics    │  │ │
│  │  │  This Week  │ │ + Explore   │ │  Continuing │ │  (existing) │  │ │
│  │  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘  │ │
│  └───────────────────────────────────────────────────────────────────┘ │
│                      │                                                  │
│                      ▼                                                  │
│  ┌───────────────────────────────────────────────────────────────────┐ │
│  │                    Intelligence Dashboard                          │ │
│  │         Preferences │ Sources │ Blind Spots │ Predictions         │ │
│  └───────────────────────────────────────────────────────────────────┘ │
│                      ▲                                                  │
│  FEEDBACK LOOP       │                                                  │
│  ┌───────────────────┴───────────────────────────────────────────────┐ │
│  │   User Votes (existing) → Source Prefs → Topic Prefs → Keywords   │ │
│  └───────────────────────────────────────────────────────────────────┘ │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Implementation Phases

### Phase Summary

| Phase | Name | Deliverable | Dependencies | Effort |
|-------|------|-------------|--------------|--------|
| 1 | Feedback Intelligence | Learned preferences from votes | Existing feedback | Low |
| 2 | Trend Detection | "Trending This Week" section | Phase 1 | Low |
| 3 | Source Discovery | Self-growing source list | Phases 1-2 | Medium |
| 4 | Personalized Recommendations | "For You" + "Explore" sections | Phases 1-3 | Medium |
| 5 | Story Tracking | Cross-week narrative threads | Phases 1-4 | Medium |
| 6 | Intelligence Dashboard | Meta-view of information diet | All phases | Low |

---

## Phase 1: Feedback Intelligence

### Goal
Transform existing vote data into actionable preference signals for sources, topics, and keywords.

### Current State
- `article_feedback` table captures upvotes/downvotes
- Feedback only affects article sort order within weekly report
- No learning or preference extraction

### Target State
- Build preference models from feedback patterns
- Track source-level preferences (which sources does user prefer?)
- Track topic-level preferences (which topics does user engage with?)
- Extract keyword signals from voted articles
- Display learned preferences to user

### Database Schema Changes

```sql
-- Source preference scores (aggregated from article votes)
CREATE TABLE source_preferences (
    source TEXT PRIMARY KEY,
    preference_score INTEGER DEFAULT 0,  -- Sum of article votes
    upvote_count INTEGER DEFAULT 0,
    downvote_count INTEGER DEFAULT 0,
    article_count INTEGER DEFAULT 0,
    updated_at TIMESTAMP
);

-- Topic preference scores
CREATE TABLE topic_preferences (
    topic TEXT PRIMARY KEY,
    preference_score INTEGER DEFAULT 0,
    upvote_count INTEGER DEFAULT 0,
    downvote_count INTEGER DEFAULT 0,
    updated_at TIMESTAMP
);

-- Keyword signals extracted from voted articles
CREATE TABLE keyword_preferences (
    keyword TEXT PRIMARY KEY,
    preference_score REAL DEFAULT 0.0,  -- Weighted by vote strength
    occurrence_count INTEGER DEFAULT 0,
    updated_at TIMESTAMP
);

-- Schema version tracking (add to migrations)
-- Migration 4: Preference tables
```

### New Module: `src/preference_learner.py`

```python
"""Learn user preferences from feedback patterns."""

from dataclasses import dataclass
from .database import Database, get_db


@dataclass
class UserPreferences:
    """Aggregated user preferences."""

    source_scores: dict[str, int]      # source -> preference score
    topic_scores: dict[str, int]       # topic -> preference score
    keyword_scores: dict[str, float]   # keyword -> weighted score

    @property
    def preferred_sources(self) -> list[str]:
        """Sources with positive preference."""
        return [s for s, score in self.source_scores.items() if score > 0]

    @property
    def disfavored_sources(self) -> list[str]:
        """Sources with negative preference."""
        return [s for s, score in self.source_scores.items() if score < 0]


class PreferenceLearner:
    """Learn and update user preferences from feedback."""

    def __init__(self, db: Database | None = None):
        self.db = db or get_db()

    def learn_from_feedback(self, article_id: int, vote: int) -> None:
        """
        Update preference models when user votes on an article.

        Args:
            article_id: The voted article
            vote: +1 (upvote), -1 (downvote), 0 (reset)
        """
        article = self.db.get_article_by_id(article_id)
        if not article:
            return

        # Update source preference
        self._update_source_preference(article.source, vote)

        # Update topic preferences
        for topic in article.topics:
            self._update_topic_preference(topic, vote)

        # Extract and update keyword preferences
        keywords = self._extract_keywords(article)
        for keyword, weight in keywords:
            self._update_keyword_preference(keyword, vote * weight)

    def get_preferences(self) -> UserPreferences:
        """Get current user preference model."""
        return UserPreferences(
            source_scores=self._get_source_scores(),
            topic_scores=self._get_topic_scores(),
            keyword_scores=self._get_keyword_scores(),
        )

    def rebuild_preferences(self) -> None:
        """Rebuild all preferences from scratch using all feedback."""
        # Clear existing preferences
        # Iterate through all feedback and relearn
        pass

    def _extract_keywords(self, article) -> list[tuple[str, float]]:
        """Extract weighted keywords from article title and summary."""
        # Simple approach: title words weighted 1.0, summary words 0.5
        # Filter stopwords, minimum length, etc.
        pass

    # ... helper methods for DB operations
```

### Integration Points

**1. Feedback routes in `server.py`:**
```python
@app.route("/feedback/<int:article_id>/upvote", methods=["POST"])
def upvote_article(article_id: int):
    db = get_db()
    db.set_article_feedback(article_id, user_score=1)

    # NEW: Update preference model
    learner = PreferenceLearner(db)
    learner.learn_from_feedback(article_id, vote=1)

    # ... redirect
```

**2. New route for preferences page:**
```python
@app.route("/preferences")
def preferences_page():
    """Display learned preferences."""
    learner = PreferenceLearner()
    prefs = learner.get_preferences()
    return render_template("preferences.html", preferences=prefs)
```

### New Template: `templates/preferences.html`

```html
{% extends "base.html" %}
{% block content %}
<div class="container">
    <h1>Your Preferences</h1>
    <p class="intro">Learned from your article votes</p>

    <section class="preference-section">
        <h2>Preferred Sources</h2>
        {% if preferences.preferred_sources %}
        <ul class="preference-list positive">
            {% for source in preferences.preferred_sources %}
            <li>{{ source }} <span class="score">+{{ preferences.source_scores[source] }}</span></li>
            {% endfor %}
        </ul>
        {% else %}
        <p class="empty">No preferred sources yet. Upvote articles to train.</p>
        {% endif %}
    </section>

    <section class="preference-section">
        <h2>Disfavored Sources</h2>
        {% if preferences.disfavored_sources %}
        <ul class="preference-list negative">
            {% for source in preferences.disfavored_sources %}
            <li>{{ source }} <span class="score">{{ preferences.source_scores[source] }}</span></li>
            {% endfor %}
        </ul>
        {% else %}
        <p class="empty">No disfavored sources yet.</p>
        {% endif %}
    </section>

    <section class="preference-section">
        <h2>Topic Interests</h2>
        <!-- Similar structure for topics -->
    </section>

    <section class="actions">
        <form action="{{ url_for('reset_preferences') }}" method="post">
            <button type="submit" class="btn-secondary">Reset All Preferences</button>
        </form>
    </section>
</div>
{% endblock %}
```

### CLI Commands

```bash
# Show learned preferences
aicrawler preferences show

# Reset preferences
aicrawler preferences reset

# Rebuild from all historical feedback
aicrawler preferences rebuild
```

### Implementation Steps

1. Add database migration for preference tables
2. Create `src/preference_learner.py` module
3. Integrate learning into feedback routes
4. Create `/preferences` route and template
5. Add CLI commands for preference management
6. Update navigation to include preferences link

### Testing

```python
# tests/test_preference_learner.py

def test_upvote_updates_source_preference():
    """Upvoting article should increase source preference."""
    db = create_test_db()
    article_id = db.insert_article(url="...", title="...", source="TechCrunch")

    learner = PreferenceLearner(db)
    learner.learn_from_feedback(article_id, vote=1)

    prefs = learner.get_preferences()
    assert prefs.source_scores["TechCrunch"] == 1

def test_downvote_decreases_preference():
    """Downvoting should decrease preference scores."""
    # ...

def test_preference_accumulates():
    """Multiple votes on same source should accumulate."""
    # ...
```

### Deliverables

- [x] Migration 4: Preference tables (COMPLETE - 2026-01-26)
- [x] `src/preference_learner.py` module (COMPLETE - 2026-01-26)
- [x] Updated feedback routes with learning (COMPLETE - 2026-01-26)
- [x] `/preferences` page (COMPLETE - 2026-01-26)
- [x] CLI preference commands (COMPLETE - 2026-01-26)
- [ ] Unit tests (deferred - basic functionality verified via manual testing)

---

## Phase 2: Trend Detection

### Goal
Surface what's NEW and ACCELERATING compared to previous weeks.

### Key Metrics

| Metric | Definition | Use Case |
|--------|------------|----------|
| **Velocity** | % change in term frequency vs. baseline | "This topic is 340% more frequent" |
| **Novelty** | First appearance in recent history | "First time we're seeing this term" |
| **Momentum** | Sustained growth over multiple weeks | "3 weeks of consistent growth" |
| **Cooling** | Declining frequency | "This topic is fading" |

### Database Schema Changes

```sql
-- Term frequency tracking per week
CREATE TABLE weekly_term_frequencies (
    id INTEGER PRIMARY KEY,
    week_number TEXT NOT NULL,
    term TEXT NOT NULL,
    frequency INTEGER NOT NULL,
    source_distribution TEXT,  -- JSON: {"TechCrunch": 5, "ArXiv": 3}
    UNIQUE(week_number, term)
);

-- Detected trends per week
CREATE TABLE weekly_trends (
    id INTEGER PRIMARY KEY,
    week_number TEXT NOT NULL,
    term TEXT NOT NULL,
    trend_type TEXT NOT NULL,  -- "emerging", "accelerating", "cooling", "stable"
    current_frequency INTEGER,
    baseline_frequency REAL,
    velocity REAL,             -- % change
    first_seen_week TEXT,      -- NULL if not new
    momentum_weeks INTEGER,    -- Consecutive weeks of growth
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(week_number, term)
);

CREATE INDEX idx_trends_week ON weekly_trends(week_number);
CREATE INDEX idx_trends_type ON weekly_trends(trend_type);
```

### New Module: `src/trend_detector.py`

```python
"""Detect trending topics across weeks."""

import re
from collections import Counter
from dataclasses import dataclass
from datetime import datetime

from .database import Database, get_db


@dataclass
class TrendSignal:
    """A detected trend signal."""

    term: str
    trend_type: str           # "emerging", "accelerating", "cooling", "stable"
    current_frequency: int
    baseline_frequency: float
    velocity: float           # % change (-100 to +inf)
    first_seen_week: str | None
    momentum_weeks: int

    @property
    def is_new(self) -> bool:
        return self.first_seen_week is not None

    @property
    def velocity_display(self) -> str:
        if self.velocity == float('inf'):
            return "NEW"
        elif self.velocity > 0:
            return f"+{self.velocity:.0f}%"
        else:
            return f"{self.velocity:.0f}%"


class TrendDetector:
    """Detect and classify topic trends."""

    def __init__(
        self,
        db: Database | None = None,
        lookback_weeks: int = 4,
        min_frequency: int = 3,
        velocity_threshold: float = 50.0,
    ):
        """
        Initialize trend detector.

        Args:
            db: Database instance
            lookback_weeks: Number of weeks for baseline calculation
            min_frequency: Minimum occurrences to be considered a trend
            velocity_threshold: % change to be considered "accelerating"
        """
        self.db = db or get_db()
        self.lookback_weeks = lookback_weeks
        self.min_frequency = min_frequency
        self.velocity_threshold = velocity_threshold

    def detect_trends(self, week_number: str) -> list[TrendSignal]:
        """
        Detect trends for a specific week.

        Args:
            week_number: Week to analyze (e.g., "2026-W05")

        Returns:
            List of TrendSignal sorted by velocity descending
        """
        # Get current week's term frequencies
        current_terms = self._get_term_frequencies(week_number)

        # Get baseline from previous weeks
        baseline = self._calculate_baseline(week_number)

        # Get historical first appearances
        first_appearances = self._get_first_appearances()

        # Classify each term
        trends = []
        for term, frequency in current_terms.items():
            if frequency < self.min_frequency:
                continue

            baseline_freq = baseline.get(term, 0)

            # Calculate velocity
            if baseline_freq == 0:
                velocity = float('inf')
                trend_type = "emerging"
                first_seen = week_number
            else:
                velocity = (frequency - baseline_freq) / baseline_freq * 100
                first_seen = first_appearances.get(term)

                if velocity > self.velocity_threshold:
                    trend_type = "accelerating"
                elif velocity < -self.velocity_threshold:
                    trend_type = "cooling"
                else:
                    trend_type = "stable"

            # Calculate momentum (consecutive weeks of growth)
            momentum = self._calculate_momentum(term, week_number)

            trends.append(TrendSignal(
                term=term,
                trend_type=trend_type,
                current_frequency=frequency,
                baseline_frequency=baseline_freq,
                velocity=velocity,
                first_seen_week=first_seen if first_seen == week_number else None,
                momentum_weeks=momentum,
            ))

        # Sort by velocity (emerging first, then accelerating)
        trends.sort(key=lambda t: (
            0 if t.trend_type == "emerging" else 1,
            -t.velocity if t.velocity != float('inf') else float('inf')
        ))

        return trends

    def extract_and_store_terms(self, week_number: str) -> int:
        """
        Extract terms from articles and store frequencies.

        Should be called after summarization in the pipeline.

        Returns:
            Number of unique terms extracted
        """
        articles = self.db.get_articles_for_week(week_number)
        term_counter = Counter()

        for article in articles:
            terms = self._extract_terms(article)
            term_counter.update(terms)

        # Store in database
        for term, frequency in term_counter.items():
            self._store_term_frequency(week_number, term, frequency)

        return len(term_counter)

    def _extract_terms(self, article) -> list[str]:
        """Extract significant terms from article."""
        text = f"{article.title} {article.summary or ''}"

        # Extract multi-word phrases (2-3 words)
        # Extract technical terms, proper nouns
        # Filter stopwords

        # Simple approach: extract capitalized phrases and technical terms
        terms = []

        # Bi-grams and tri-grams from title
        words = re.findall(r'\b[A-Za-z][a-z]+(?:\s+[A-Za-z][a-z]+){1,2}\b', article.title)
        terms.extend([w.lower() for w in words])

        # Capitalized terms (likely proper nouns/products)
        proper_nouns = re.findall(r'\b[A-Z][a-z]+(?:\s+[A-Z][a-z]+)*\b', text)
        terms.extend([p.lower() for p in proper_nouns if len(p) > 3])

        # Include article topics
        terms.extend([t.lower() for t in article.topics])

        return terms

    def _get_term_frequencies(self, week_number: str) -> dict[str, int]:
        """Get term frequencies for a week from database."""
        # Query weekly_term_frequencies table
        pass

    def _calculate_baseline(self, week_number: str) -> dict[str, float]:
        """Calculate average frequency over lookback period."""
        # Get previous N weeks, average each term's frequency
        pass

    def _calculate_momentum(self, term: str, week_number: str) -> int:
        """Count consecutive weeks of growth for a term."""
        pass

    # ... additional helper methods
```

### Integration into Pipeline

Update `cli.py` `run` command:

```python
@main.command()
@click.pass_context
def run(ctx: click.Context) -> None:
    """Run the full pipeline."""
    # ... existing steps 1-5 ...

    # Step 6: Extract terms and detect trends (NEW)
    click.echo("\nStep 6/6: Detecting trends...")
    from .trend_detector import TrendDetector

    detector = TrendDetector(db=db)
    detector.extract_and_store_terms(week_number)
    trends = detector.detect_trends(week_number)

    # Store trends in database
    for trend in trends:
        db.insert_trend(week_number, trend)

    emerging = [t for t in trends if t.trend_type == "emerging"]
    accelerating = [t for t in trends if t.trend_type == "accelerating"]
    click.echo(f"  {len(emerging)} emerging topics, {len(accelerating)} accelerating")
```

### Template Updates

Add trends section to `templates/report.html`:

```html
{# Trending Topics Section - NEW #}
{% if trends %}
<section class="trends-section">
    <h2>Trending This Week</h2>

    {% set emerging = trends|selectattr("trend_type", "eq", "emerging")|list %}
    {% set accelerating = trends|selectattr("trend_type", "eq", "accelerating")|list %}
    {% set cooling = trends|selectattr("trend_type", "eq", "cooling")|list %}

    {% if emerging %}
    <div class="trend-group emerging">
        <h3>New This Week</h3>
        <ul class="trend-list">
            {% for trend in emerging[:5] %}
            <li>
                <span class="trend-term">{{ trend.term }}</span>
                <span class="trend-badge new">NEW</span>
                <span class="trend-freq">{{ trend.current_frequency }} mentions</span>
            </li>
            {% endfor %}
        </ul>
    </div>
    {% endif %}

    {% if accelerating %}
    <div class="trend-group accelerating">
        <h3>Gaining Momentum</h3>
        <ul class="trend-list">
            {% for trend in accelerating[:5] %}
            <li>
                <span class="trend-term">{{ trend.term }}</span>
                <span class="trend-badge up">{{ trend.velocity_display }}</span>
                {% if trend.momentum_weeks > 1 %}
                <span class="momentum">{{ trend.momentum_weeks }} weeks</span>
                {% endif %}
            </li>
            {% endfor %}
        </ul>
    </div>
    {% endif %}

    {% if cooling %}
    <div class="trend-group cooling">
        <h3>Cooling Down</h3>
        <ul class="trend-list">
            {% for trend in cooling[:3] %}
            <li>
                <span class="trend-term">{{ trend.term }}</span>
                <span class="trend-badge down">{{ trend.velocity_display }}</span>
            </li>
            {% endfor %}
        </ul>
    </div>
    {% endif %}
</section>
{% endif %}
```

### CLI Commands

```bash
# Show trends for current week
aicrawler trends

# Show trends for specific week
aicrawler trends --week 2026-W04

# Compare two weeks
aicrawler trends --compare 2026-W03 2026-W04
```

### Implementation Steps

1. Add database migration for trend tables
2. Create `src/trend_detector.py` module
3. Integrate term extraction into pipeline
4. Update aggregator to include trends in WeekData
5. Update report template with trends section
6. Add trends CLI command
7. Add CSS styling for trend badges

### Deliverables

- [x] Migration 5: Trend tables (COMPLETE - 2026-01-26)
- [x] `src/trend_detector.py` module (COMPLETE - 2026-01-26)
- [x] Pipeline integration (term extraction + trend detection) (COMPLETE - 2026-01-26)
- [x] Updated report template with trends section (COMPLETE - 2026-01-26)
- [x] CLI `trends` command (COMPLETE - 2026-01-26)
- [x] Trend-specific CSS styling (COMPLETE - 2026-01-26)
- [ ] Unit tests (deferred - basic functionality verified via manual testing)

---

## Phase 3: Dynamic Source Discovery

### Goal
Automatically discover, evaluate, and manage sources to grow the content funnel while maintaining quality.

### Discovery Channels

| Channel | Method | Signal Quality |
|---------|--------|----------------|
| **Citations** | Extract domains from article content | Medium |
| **Hacker News** | High-scoring posts in AI/ML topics | High |
| **Reddit** | Top posts from r/MachineLearning, r/LocalLLaMA | Medium-High |
| **Lobsters** | AI-tagged posts | High |
| **User Suggestions** | Web UI submission form | Variable |

### Source Lifecycle

```
┌──────────┐     ┌───────────┐     ┌──────────┐     ┌──────────┐
│Discovered│ ──► │ Candidate │ ──► │Probation │ ──► │  Active  │
└──────────┘     └───────────┘     └──────────┘     └──────────┘
                       │                 │                │
                       ▼                 ▼                ▼
                 ┌──────────┐     ┌──────────┐     ┌──────────┐
                 │ Rejected │     │ Disabled │     │ Disabled │
                 │(no feed) │     │(low qual)│     │(inactive)│
                 └──────────┘     └──────────┘     └──────────┘
```

### Database Schema Changes

```sql
-- Discovered and managed sources
CREATE TABLE discovered_sources (
    id INTEGER PRIMARY KEY,
    domain TEXT UNIQUE NOT NULL,
    name TEXT,                      -- Human-readable name
    feed_url TEXT,                  -- Discovered RSS/Atom feed
    discovery_method TEXT NOT NULL, -- "citation", "hackernews", "reddit", "user", "seed"
    discovery_context TEXT,         -- Where/how it was found

    -- Quality metrics
    quality_score REAL DEFAULT 0.0,
    relevance_score REAL DEFAULT 0.0,
    freshness_score REAL DEFAULT 0.0,  -- How often updated

    -- User feedback integration
    user_preference_score INTEGER DEFAULT 0,  -- From Phase 1

    -- Lifecycle
    status TEXT DEFAULT 'candidate',  -- candidate, probation, active, disabled, rejected
    status_reason TEXT,

    -- Stats
    articles_collected INTEGER DEFAULT 0,
    articles_summarized INTEGER DEFAULT 0,
    last_article_date TIMESTAMP,

    -- Timestamps
    discovered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    evaluated_at TIMESTAMP,
    promoted_at TIMESTAMP,
    disabled_at TIMESTAMP
);

CREATE INDEX idx_sources_status ON discovered_sources(status);
CREATE INDEX idx_sources_domain ON discovered_sources(domain);

-- Source evaluation history
CREATE TABLE source_evaluations (
    id INTEGER PRIMARY KEY,
    source_id INTEGER REFERENCES discovered_sources(id),
    evaluated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    quality_score REAL,
    relevance_score REAL,
    freshness_score REAL,
    notes TEXT
);
```

### New Module: `src/source_discovery.py`

```python
"""Discover and manage content sources."""

import re
from dataclasses import dataclass
from urllib.parse import urlparse

import feedfinder2
import httpx

from .database import Database, get_db


@dataclass
class CandidateSource:
    """A discovered source candidate."""

    domain: str
    name: str | None
    feed_url: str | None
    discovery_method: str
    discovery_context: str | None
    quality_score: float = 0.0


class SourceDiscoverer:
    """Discover new content sources."""

    def __init__(self, db: Database | None = None):
        self.db = db or get_db()

    def discover_from_citations(self, articles: list) -> list[CandidateSource]:
        """
        Extract potential sources from URLs cited in articles.

        Looks for blog posts, news sites referenced in article content.
        """
        candidates = []
        seen_domains = set()

        for article in articles:
            urls = self._extract_urls(article.content or "")
            for url in urls:
                domain = self._extract_domain(url)
                if domain and domain not in seen_domains:
                    seen_domains.add(domain)
                    candidates.append(CandidateSource(
                        domain=domain,
                        name=None,
                        feed_url=None,
                        discovery_method="citation",
                        discovery_context=f"Cited in: {article.title}",
                    ))

        return candidates

    def discover_from_hackernews(
        self,
        min_points: int = 100,
        topics: list[str] = ["AI", "LLM", "machine learning"],
    ) -> list[CandidateSource]:
        """
        Get sources from high-scoring Hacker News posts.

        Uses HN Algolia API to find relevant posts.
        """
        candidates = []

        for topic in topics:
            # Query HN Algolia API
            url = f"https://hn.algolia.com/api/v1/search"
            params = {
                "query": topic,
                "tags": "story",
                "numericFilters": f"points>{min_points}",
            }

            response = httpx.get(url, params=params)
            if response.status_code == 200:
                hits = response.json().get("hits", [])
                for hit in hits:
                    story_url = hit.get("url")
                    if story_url:
                        domain = self._extract_domain(story_url)
                        if domain:
                            candidates.append(CandidateSource(
                                domain=domain,
                                name=None,
                                feed_url=None,
                                discovery_method="hackernews",
                                discovery_context=f"HN: {hit.get('title', '')} ({hit.get('points', 0)} pts)",
                            ))

        return self._deduplicate(candidates)

    def discover_from_reddit(
        self,
        subreddits: list[str] = ["MachineLearning", "LocalLLaMA"],
        min_upvotes: int = 100,
    ) -> list[CandidateSource]:
        """Get sources from top Reddit posts."""
        # Similar implementation using Reddit API
        pass

    def evaluate_source(self, candidate: CandidateSource) -> CandidateSource:
        """
        Evaluate a candidate source for quality.

        Checks:
        - Does it have an RSS feed?
        - How often is it updated?
        - Is content relevant to our topics?
        """
        # Try to find RSS feed
        try:
            feeds = feedfinder2.find_feeds(f"https://{candidate.domain}")
            if feeds:
                candidate.feed_url = feeds[0]
                candidate.quality_score += 0.3
        except Exception:
            pass

        # Check update frequency by parsing feed
        if candidate.feed_url:
            freshness = self._check_feed_freshness(candidate.feed_url)
            candidate.quality_score += freshness * 0.3

        # Check content relevance (sample a few articles)
        relevance = self._check_relevance(candidate)
        candidate.quality_score += relevance * 0.4

        return candidate

    def _extract_urls(self, text: str) -> list[str]:
        """Extract URLs from text."""
        url_pattern = r'https?://[^\s<>"{}|\\^`\[\]]+'
        return re.findall(url_pattern, text)

    def _extract_domain(self, url: str) -> str | None:
        """Extract domain from URL, filtering common non-source domains."""
        try:
            parsed = urlparse(url)
            domain = parsed.netloc.lower()

            # Filter out common non-source domains
            excluded = {
                "github.com", "twitter.com", "x.com", "youtube.com",
                "arxiv.org",  # We handle this separately
                "linkedin.com", "facebook.com",
            }

            if domain in excluded:
                return None

            # Remove www prefix
            if domain.startswith("www."):
                domain = domain[4:]

            return domain
        except Exception:
            return None

    # ... additional helper methods


class SourceManager:
    """Manage source lifecycle and quality."""

    def __init__(self, db: Database | None = None):
        self.db = db or get_db()
        self.discoverer = SourceDiscoverer(db)

    def run_discovery(self) -> dict:
        """Run full discovery pipeline."""
        results = {"discovered": 0, "evaluated": 0, "added": 0}

        # Discover from multiple channels
        candidates = []
        candidates.extend(self.discoverer.discover_from_hackernews())
        candidates.extend(self.discoverer.discover_from_reddit())

        # Also check citations from recent articles
        recent_articles = self.db.get_articles_for_week(get_current_week())
        candidates.extend(self.discoverer.discover_from_citations(recent_articles))

        results["discovered"] = len(candidates)

        # Evaluate and add new candidates
        for candidate in candidates:
            if self._is_known_source(candidate.domain):
                continue

            evaluated = self.discoverer.evaluate_source(candidate)
            results["evaluated"] += 1

            if evaluated.quality_score > 0.5 and evaluated.feed_url:
                self._add_candidate(evaluated)
                results["added"] += 1

        return results

    def promote_source(self, source_id: int) -> None:
        """Promote source from probation to active."""
        pass

    def disable_source(self, source_id: int, reason: str) -> None:
        """Disable a source."""
        pass

    def get_active_feeds(self) -> list[str]:
        """Get feed URLs for all active sources."""
        pass
```

### Web UI for Source Management

New template `templates/sources.html`:

```html
{% extends "base.html" %}
{% block content %}
<div class="container">
    <h1>Content Sources</h1>

    {# Suggest a source form #}
    <section class="suggest-source">
        <h2>Suggest a Source</h2>
        <form action="{{ url_for('suggest_source') }}" method="post">
            <input type="url" name="url" placeholder="https://example.com" required>
            <button type="submit">Suggest</button>
        </form>
    </section>

    {# Active sources #}
    <section class="source-list">
        <h2>Active Sources ({{ active_sources|length }})</h2>
        <table>
            <thead>
                <tr>
                    <th>Source</th>
                    <th>Articles</th>
                    <th>Your Preference</th>
                    <th>Actions</th>
                </tr>
            </thead>
            <tbody>
                {% for source in active_sources %}
                <tr>
                    <td>
                        <strong>{{ source.name or source.domain }}</strong>
                        <br><small>{{ source.domain }}</small>
                    </td>
                    <td>{{ source.articles_collected }}</td>
                    <td>
                        {% if source.user_preference_score > 0 %}
                        <span class="pref-positive">+{{ source.user_preference_score }}</span>
                        {% elif source.user_preference_score < 0 %}
                        <span class="pref-negative">{{ source.user_preference_score }}</span>
                        {% else %}
                        <span class="pref-neutral">—</span>
                        {% endif %}
                    </td>
                    <td>
                        <form action="{{ url_for('disable_source', source_id=source.id) }}" method="post" class="inline">
                            <button type="submit" class="btn-small">Disable</button>
                        </form>
                    </td>
                </tr>
                {% endfor %}
            </tbody>
        </table>
    </section>

    {# Sources in probation #}
    <section class="source-list">
        <h2>In Probation ({{ probation_sources|length }})</h2>
        <!-- Similar table with Promote/Reject actions -->
    </section>

    {# Candidates #}
    <section class="source-list">
        <h2>Candidates ({{ candidate_sources|length }})</h2>
        <!-- Similar table with Approve/Reject actions -->
    </section>
</div>
{% endblock %}
```

### Integration with Collector

Update `collector.py` to use dynamic sources:

```python
class ArticleCollector:
    def __init__(self, config: dict):
        self.config = config
        self.db = get_db()
        self.source_manager = SourceManager(self.db)

    def collect(self) -> CollectionResult:
        # Get feeds from config (seed sources)
        seed_feeds = self.config.get("sources", {}).get("feeds", [])

        # Get feeds from active discovered sources
        dynamic_feeds = self.source_manager.get_active_feeds()

        all_feeds = seed_feeds + dynamic_feeds

        # ... rest of collection logic
```

### CLI Commands

```bash
# Run source discovery
aicrawler sources discover

# List all sources by status
aicrawler sources list
aicrawler sources list --status active
aicrawler sources list --status probation

# Manually add a source
aicrawler sources add https://example.com/blog

# Promote/disable sources
aicrawler sources promote <id>
aicrawler sources disable <id> --reason "Low quality"
```

### New Dependencies

```toml
# pyproject.toml
[project.dependencies]
feedfinder2 = "^0.0.4"    # RSS feed discovery
trafilatura = "^1.6.0"     # Article content extraction
httpx = "^0.27.0"          # HTTP client for API calls
```

### Implementation Steps

1. Add database migration for source tables
2. Add new dependencies to pyproject.toml
3. Create `src/source_discovery.py` module
4. Create `/sources` route and template
5. Integrate dynamic sources into collector
6. Add source management CLI commands
7. Add source discovery to weekly pipeline (optional)

### Deliverables

- [x] Migration 6: Source tables (COMPLETE - 2026-01-26)
- [x] New dependency: feedfinder2 (COMPLETE - 2026-01-26)
- [x] `src/source_discovery.py` module (COMPLETE - 2026-01-26)
- [x] `/sources` management page (COMPLETE - 2026-01-26)
- [x] Collector integration with dynamic sources (COMPLETE - 2026-01-26)
- [x] CLI source commands (COMPLETE - 2026-01-26)
- [ ] Unit tests (deferred - basic functionality verified via manual testing)

---

## Phase 4: Personalized Recommendations

### Goal
Create "For You" recommendations based on learned preferences, with deliberate serendipity injection.

### Recommendation Strategy

```
Article Score = (
    source_preference * 0.25 +      # Phase 1
    topic_preference * 0.25 +       # Phase 1
    trend_bonus * 0.15 +            # Phase 2
    base_relevance * 0.20 +         # Existing
    recency_bonus * 0.15            # Newer = better
)

Selection:
- Top 80%: Highest scoring (exploitation)
- Bottom 20%: Random from remaining (exploration/serendipity)
```

### New Module: `src/recommender.py`

```python
"""Personalized article recommendations."""

import random
from dataclasses import dataclass

from .database import Article, Database, get_db
from .preference_learner import PreferenceLearner, UserPreferences
from .trend_detector import TrendDetector


@dataclass
class ScoredArticle:
    """Article with recommendation score and explanation."""

    article: Article
    score: float
    reasons: list[str]  # Why this was recommended


@dataclass
class Recommendations:
    """Personalized recommendations for a user."""

    for_you: list[ScoredArticle]      # High preference match
    explore: list[ScoredArticle]       # Serendipity picks
    because_trending: list[ScoredArticle]  # Trending topics


class Recommender:
    """Generate personalized article recommendations."""

    def __init__(
        self,
        db: Database | None = None,
        explore_ratio: float = 0.2,
    ):
        self.db = db or get_db()
        self.preference_learner = PreferenceLearner(db)
        self.trend_detector = TrendDetector(db)
        self.explore_ratio = explore_ratio

    def recommend(
        self,
        week_number: str,
        max_for_you: int = 10,
        max_explore: int = 3,
    ) -> Recommendations:
        """
        Generate recommendations for a week.

        Args:
            week_number: Week to recommend from
            max_for_you: Maximum "For You" articles
            max_explore: Maximum "Explore" articles

        Returns:
            Recommendations with scored articles
        """
        articles = self.db.get_articles_for_week(week_number)
        if not articles:
            return Recommendations(for_you=[], explore=[], because_trending=[])

        preferences = self.preference_learner.get_preferences()
        trends = self.trend_detector.detect_trends(week_number)
        trending_terms = {t.term.lower() for t in trends if t.trend_type in ("emerging", "accelerating")}

        # Score all articles
        scored = []
        for article in articles:
            score, reasons = self._score_article(article, preferences, trending_terms)
            scored.append(ScoredArticle(article=article, score=score, reasons=reasons))

        # Sort by score
        scored.sort(key=lambda x: x.score, reverse=True)

        # Split into "For You" and "Explore"
        split_idx = int(len(scored) * (1 - self.explore_ratio))

        for_you = scored[:min(max_for_you, split_idx)]

        # Random selection from lower-scored for serendipity
        explore_pool = scored[split_idx:]
        random.shuffle(explore_pool)
        explore = explore_pool[:max_explore]

        # Add "why explore" reasons
        for item in explore:
            item.reasons.append("Expanding your horizons")

        # Separate trending articles
        because_trending = [s for s in scored if any("trending" in r.lower() for r in s.reasons)][:5]

        return Recommendations(
            for_you=for_you,
            explore=explore,
            because_trending=because_trending,
        )

    def _score_article(
        self,
        article: Article,
        preferences: UserPreferences,
        trending_terms: set[str],
    ) -> tuple[float, list[str]]:
        """
        Score an article based on preferences and trends.

        Returns:
            Tuple of (score, list of reasons)
        """
        score = 0.0
        reasons = []

        # Source preference (25%)
        source_pref = preferences.source_scores.get(article.source, 0)
        if source_pref > 0:
            score += 0.25 * min(source_pref / 5, 1.0)  # Normalize
            reasons.append(f"From {article.source} (preferred source)")
        elif source_pref < 0:
            score -= 0.1  # Penalty for disfavored source

        # Topic preference (25%)
        topic_score = 0
        matched_topics = []
        for topic in article.topics:
            topic_pref = preferences.topic_scores.get(topic, 0)
            if topic_pref > 0:
                topic_score += topic_pref
                matched_topics.append(topic)

        if topic_score > 0:
            score += 0.25 * min(topic_score / 10, 1.0)
            reasons.append(f"Matches interests: {', '.join(matched_topics)}")

        # Trend bonus (15%)
        article_terms = set(t.lower() for t in article.topics)
        article_terms.add(article.title.lower())

        trending_matches = article_terms & trending_terms
        if trending_matches:
            score += 0.15
            reasons.append(f"Trending: {', '.join(list(trending_matches)[:2])}")

        # Base relevance (20%)
        if article.relevance_score:
            score += 0.20 * (article.relevance_score / 5)

        # Recency bonus (15%)
        # Newer articles get slight boost
        score += 0.10  # Simplified; could use actual date comparison

        return score, reasons
```

### Template Updates

Update `templates/report.html` to include recommendations:

```html
{# Personalized Recommendations - NEW #}
{% if recommendations %}
<section class="recommendations-section">
    <h2>Recommended For You</h2>
    <p class="section-intro">Based on your reading patterns</p>

    <div class="recommendation-list">
        {% for rec in recommendations.for_you %}
        <article class="recommendation-card">
            <h3><a href="{{ rec.article.url }}" target="_blank">{{ rec.article.title }}</a></h3>
            <p class="rec-summary">{{ rec.article.summary[:150] }}...</p>
            <div class="rec-meta">
                <span class="source">{{ rec.article.source }}</span>
                <span class="rec-reasons">
                    {% for reason in rec.reasons %}
                    <span class="reason-tag">{{ reason }}</span>
                    {% endfor %}
                </span>
            </div>
        </article>
        {% endfor %}
    </div>
</section>

{% if recommendations.explore %}
<section class="explore-section">
    <h2>Explore Something Different</h2>
    <p class="section-intro">Deliberately outside your usual interests</p>

    <div class="explore-list">
        {% for rec in recommendations.explore %}
        <article class="explore-card">
            <h3><a href="{{ rec.article.url }}" target="_blank">{{ rec.article.title }}</a></h3>
            <p class="explore-summary">{{ rec.article.summary[:100] }}...</p>
            <span class="source">{{ rec.article.source }}</span>
        </article>
        {% endfor %}
    </div>
</section>
{% endif %}
{% endif %}
```

### Integration

Update aggregator to include recommendations:

```python
# In aggregator.py

@dataclass
class WeekData:
    # ... existing fields ...
    recommendations: Recommendations | None = None


class Aggregator:
    def aggregate_week(self, week_number: str | None = None) -> WeekData:
        # ... existing logic ...

        # Add recommendations
        recommender = Recommender(db=self.db)
        recommendations = recommender.recommend(week_number)

        return WeekData(
            # ... existing fields ...
            recommendations=recommendations,
        )
```

### Implementation Steps

1. Create `src/recommender.py` module
2. Update WeekData dataclass with recommendations
3. Integrate recommender into aggregator
4. Update report template with recommendation sections
5. Add CSS styling for recommendation cards
6. Test with various preference profiles

### Deliverables

- [x] `src/recommender.py` module (COMPLETE - 2026-01-26)
- [x] Aggregator integration (COMPLETE - 2026-01-26)
- [x] Report template updates (For You + Explore sections) (COMPLETE - 2026-01-26)
- [x] Recommendation-specific CSS (COMPLETE - 2026-01-26)
- [ ] Unit tests for scoring logic (deferred - basic functionality verified)

---

## Phase 5: Story Tracking

### Goal
Track evolving stories across weeks, connecting related articles into narrative threads.

### Story Detection Approach

1. **Embedding-based similarity**: Articles about the same story have similar embeddings
2. **Entity overlap**: Articles mentioning same companies, products, people
3. **Temporal clustering**: Related articles appear in sequence
4. **Manual linking**: User can manually connect articles to stories

### Database Schema Changes

```sql
-- Stories spanning multiple weeks
CREATE TABLE stories (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    status TEXT DEFAULT 'active',  -- active, dormant, concluded
    first_week TEXT NOT NULL,
    last_week TEXT,
    article_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);

-- Article-story associations
CREATE TABLE story_articles (
    story_id INTEGER REFERENCES stories(id),
    article_id INTEGER REFERENCES articles(id),
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    added_method TEXT,  -- "auto", "manual"
    PRIMARY KEY (story_id, article_id)
);

-- Story timeline events
CREATE TABLE story_events (
    id INTEGER PRIMARY KEY,
    story_id INTEGER REFERENCES stories(id),
    week_number TEXT NOT NULL,
    event_summary TEXT,
    article_count INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_stories_status ON stories(status);
CREATE INDEX idx_story_articles_story ON story_articles(story_id);
```

### New Module: `src/story_tracker.py`

```python
"""Track evolving stories across weeks."""

from dataclasses import dataclass

from sentence_transformers import SentenceTransformer

from .database import Article, Database, get_db


@dataclass
class Story:
    """A story spanning multiple weeks."""

    id: int | None
    title: str
    description: str | None
    status: str
    first_week: str
    last_week: str | None
    article_ids: list[int]
    timeline: list[dict]  # [{week, summary, articles}]


class StoryTracker:
    """Track and link articles into stories."""

    def __init__(
        self,
        db: Database | None = None,
        similarity_threshold: float = 0.75,
    ):
        self.db = db or get_db()
        self.similarity_threshold = similarity_threshold
        self.model = SentenceTransformer('all-MiniLM-L6-v2')

    def process_week(self, week_number: str) -> dict:
        """
        Process new articles and link to existing or new stories.

        Returns:
            Stats about story processing
        """
        articles = self.db.get_articles_for_week(week_number)
        active_stories = self._get_active_stories()

        stats = {"linked": 0, "new_stories": 0}

        for article in articles:
            if self._is_already_linked(article.id):
                continue

            # Find matching story
            match = self._find_matching_story(article, active_stories)

            if match:
                self._add_to_story(article, match)
                stats["linked"] += 1
            else:
                # Check if this article starts a new story cluster
                cluster = self._find_cluster(article, articles)
                if len(cluster) >= 3:  # Minimum articles for a story
                    story = self._create_story(cluster, week_number)
                    stats["new_stories"] += 1

        # Update story statuses
        self._update_story_statuses(week_number)

        return stats

    def _find_matching_story(
        self,
        article: Article,
        stories: list[Story],
    ) -> Story | None:
        """Find a story that matches this article."""
        article_embedding = self._get_embedding(article)

        best_match = None
        best_score = 0.0

        for story in stories:
            # Get embeddings of recent story articles
            story_embedding = self._get_story_embedding(story)

            similarity = self._cosine_similarity(article_embedding, story_embedding)

            if similarity > self.similarity_threshold and similarity > best_score:
                best_match = story
                best_score = similarity

        return best_match

    def _find_cluster(
        self,
        seed_article: Article,
        all_articles: list[Article],
    ) -> list[Article]:
        """Find articles that cluster with the seed article."""
        seed_embedding = self._get_embedding(seed_article)
        cluster = [seed_article]

        for article in all_articles:
            if article.id == seed_article.id:
                continue

            article_embedding = self._get_embedding(article)
            similarity = self._cosine_similarity(seed_embedding, article_embedding)

            if similarity > self.similarity_threshold:
                cluster.append(article)

        return cluster

    def _create_story(self, articles: list[Article], week_number: str) -> Story:
        """Create a new story from a cluster of articles."""
        # Generate title and description using LLM or heuristics
        title = self._generate_story_title(articles)

        story_id = self.db.insert_story(
            title=title,
            description=None,
            first_week=week_number,
        )

        for article in articles:
            self.db.add_article_to_story(story_id, article.id, method="auto")

        return Story(
            id=story_id,
            title=title,
            description=None,
            status="active",
            first_week=week_number,
            last_week=week_number,
            article_ids=[a.id for a in articles],
            timeline=[],
        )

    def _get_embedding(self, article: Article) -> list[float]:
        """Get embedding for article."""
        text = f"{article.title}. {article.summary or ''}"
        return self.model.encode(text).tolist()

    def _get_story_embedding(self, story: Story) -> list[float]:
        """Get aggregate embedding for a story."""
        # Average of recent article embeddings
        pass

    def _cosine_similarity(self, a: list[float], b: list[float]) -> float:
        """Calculate cosine similarity between two vectors."""
        import numpy as np
        a, b = np.array(a), np.array(b)
        return float(np.dot(a, b) / (np.linalg.norm(a) * np.linalg.norm(b)))

    def _update_story_statuses(self, current_week: str) -> None:
        """Mark stories as dormant if no new articles in 2+ weeks."""
        pass

    def get_active_stories(self) -> list[Story]:
        """Get all active stories with their timelines."""
        pass

    def get_story_timeline(self, story_id: int) -> list[dict]:
        """Get timeline of events for a story."""
        pass
```

### Template Updates

Add stories section to report:

```html
{# Ongoing Stories Section #}
{% if stories %}
<section class="stories-section">
    <h2>Ongoing Stories</h2>
    <p class="section-intro">Developments you've been following</p>

    {% for story in stories %}
    <article class="story-card">
        <header class="story-header">
            <h3>{{ story.title }}</h3>
            <span class="story-duration">{{ story.timeline|length }} weeks</span>
        </header>

        <div class="story-timeline">
            {% for event in story.timeline[-3:] %}
            <div class="timeline-event {% if loop.last %}current{% endif %}">
                <span class="event-week">{{ event.week }}</span>
                <span class="event-summary">{{ event.summary }}</span>
                {% if loop.last %}
                <span class="new-badge">NEW</span>
                {% endif %}
            </div>
            {% endfor %}
        </div>

        <details class="story-articles">
            <summary>View all {{ story.article_count }} articles</summary>
            <ul>
                {% for article in story.articles %}
                <li>
                    <a href="{{ article.url }}">{{ article.title }}</a>
                    <span class="week">{{ article.week_number }}</span>
                </li>
                {% endfor %}
            </ul>
        </details>
    </article>
    {% endfor %}
</section>
{% endif %}
```

### New Dependencies

```toml
# pyproject.toml
[project.dependencies]
sentence-transformers = "^2.2.0"  # For embeddings
```

### Implementation Steps

1. Add database migration for story tables
2. Add sentence-transformers dependency
3. Create `src/story_tracker.py` module
4. Integrate story processing into pipeline
5. Update aggregator to include stories
6. Update report template with stories section
7. Add story management routes (manual linking)
8. Add CSS for story timeline visualization

### Deliverables

- [x] Migration 7: Story tables (COMPLETE - 2026-01-26)
- [x] sentence-transformers dependency (COMPLETE - 2026-01-26)
- [x] `src/story_tracker.py` module (COMPLETE - 2026-01-26)
- [x] Aggregator integration (COMPLETE - 2026-01-26)
- [x] Report template with stories section (COMPLETE - 2026-01-26)
- [x] Story timeline CSS (COMPLETE - 2026-01-26)
- [ ] Pipeline integration (CLI command for story processing - deferred)
- [ ] Manual story linking UI (optional - deferred)

---

## Phase 6: Intelligence Dashboard

### Goal
Provide meta-view of information diet, blind spots, and predictions.

### Dashboard Components

1. **Reading Stats**: Articles processed, sources active, stories tracked
2. **Focus Areas**: Topics you engage with most
3. **Blind Spots**: Trending topics you're ignoring
4. **Source Health**: Which sources are performing well
5. **Predictions**: What's likely to trend next (simple extrapolation)

### New Template: `templates/dashboard.html`

```html
{% extends "base.html" %}
{% block content %}
<div class="container dashboard">
    <h1>Intelligence Dashboard</h1>

    {# Overview Stats #}
    <section class="stats-overview">
        <div class="stat-card">
            <span class="stat-value">{{ stats.articles_this_month }}</span>
            <span class="stat-label">Articles This Month</span>
        </div>
        <div class="stat-card">
            <span class="stat-value">{{ stats.active_sources }}</span>
            <span class="stat-label">Active Sources</span>
        </div>
        <div class="stat-card">
            <span class="stat-value">{{ stats.active_stories }}</span>
            <span class="stat-label">Stories Tracking</span>
        </div>
        <div class="stat-card">
            <span class="stat-value">{{ stats.feedback_given }}</span>
            <span class="stat-label">Votes Given</span>
        </div>
    </section>

    {# Focus vs Blind Spots #}
    <section class="focus-analysis">
        <div class="focus-column">
            <h2>Your Focus Areas</h2>
            <p class="subtitle">Topics you engage with most</p>
            <ul class="topic-bars">
                {% for topic, pct in focus_areas %}
                <li>
                    <span class="topic-name">{{ topic }}</span>
                    <div class="bar" style="width: {{ pct }}%"></div>
                    <span class="topic-pct">{{ pct }}%</span>
                </li>
                {% endfor %}
            </ul>
        </div>

        <div class="blindspot-column">
            <h2>Blind Spots</h2>
            <p class="subtitle">Trending topics you're skipping</p>
            <ul class="blindspot-list">
                {% for item in blind_spots %}
                <li>
                    <span class="topic-name">{{ item.topic }}</span>
                    <span class="trend-badge">{{ item.trend }}</span>
                    <span class="skip-count">{{ item.skipped }} articles skipped</span>
                </li>
                {% endfor %}
            </ul>
        </div>
    </section>

    {# Source Performance #}
    <section class="source-performance">
        <h2>Source Performance</h2>
        <table class="performance-table">
            <thead>
                <tr>
                    <th>Source</th>
                    <th>Articles</th>
                    <th>Your Rating</th>
                    <th>Avg Relevance</th>
                    <th>Status</th>
                </tr>
            </thead>
            <tbody>
                {% for source in source_stats %}
                <tr>
                    <td>{{ source.name }}</td>
                    <td>{{ source.article_count }}</td>
                    <td>
                        {% if source.user_pref > 0 %}👍{% elif source.user_pref < 0 %}👎{% else %}—{% endif %}
                    </td>
                    <td>{{ "%.1f"|format(source.avg_relevance) }}/5</td>
                    <td><span class="status-{{ source.status }}">{{ source.status }}</span></td>
                </tr>
                {% endfor %}
            </tbody>
        </table>
    </section>

    {# Simple Predictions #}
    <section class="predictions">
        <h2>What to Watch</h2>
        <p class="subtitle">Based on momentum trends</p>
        <ul class="prediction-list">
            {% for pred in predictions %}
            <li>
                <span class="pred-topic">{{ pred.topic }}</span>
                <span class="pred-reason">{{ pred.reason }}</span>
            </li>
            {% endfor %}
        </ul>
    </section>
</div>
{% endblock %}
```

### New Module: `src/dashboard.py`

```python
"""Intelligence dashboard data aggregation."""

from dataclasses import dataclass
from .database import get_db
from .preference_learner import PreferenceLearner
from .trend_detector import TrendDetector


@dataclass
class DashboardData:
    """Aggregated dashboard data."""

    stats: dict
    focus_areas: list[tuple[str, float]]
    blind_spots: list[dict]
    source_stats: list[dict]
    predictions: list[dict]


class DashboardBuilder:
    """Build intelligence dashboard data."""

    def __init__(self):
        self.db = get_db()
        self.preferences = PreferenceLearner(self.db)
        self.trends = TrendDetector(self.db)

    def build(self) -> DashboardData:
        """Build complete dashboard data."""
        return DashboardData(
            stats=self._get_overview_stats(),
            focus_areas=self._get_focus_areas(),
            blind_spots=self._find_blind_spots(),
            source_stats=self._get_source_stats(),
            predictions=self._generate_predictions(),
        )

    def _get_overview_stats(self) -> dict:
        """Get overview statistics."""
        # Query counts from database
        pass

    def _get_focus_areas(self) -> list[tuple[str, float]]:
        """Get topics user focuses on most."""
        prefs = self.preferences.get_preferences()
        # Convert to percentages
        pass

    def _find_blind_spots(self) -> list[dict]:
        """Find trending topics user is ignoring."""
        # Compare trending topics with user engagement
        # Topics that are trending but user hasn't upvoted
        pass

    def _get_source_stats(self) -> list[dict]:
        """Get performance stats for each source."""
        pass

    def _generate_predictions(self) -> list[dict]:
        """Generate simple predictions based on trends."""
        # Extrapolate accelerating trends
        # "X is likely to dominate next 2 weeks based on 3-week momentum"
        pass
```

### Implementation Steps

1. Create `src/dashboard.py` module
2. Add `/dashboard` route
3. Create `templates/dashboard.html`
4. Add CSS for dashboard visualizations
5. Link dashboard in navigation

### Deliverables

- [x] `src/dashboard.py` module (COMPLETE - 2026-01-26)
- [x] `/dashboard` route (COMPLETE - 2026-01-26)
- [x] Dashboard template (COMPLETE - 2026-01-26)
- [x] Dashboard CSS (bars, tables, cards) (COMPLETE - 2026-01-26)
- [x] Navigation update (COMPLETE - 2026-01-26)

---

## Updated Directory Structure

```
AICrawler/
├── pyproject.toml
├── config.yaml
├── src/
│   ├── __init__.py
│   ├── cli.py
│   ├── collector.py
│   ├── feed_parser.py
│   ├── api_client.py
│   ├── filter.py
│   ├── summarizer.py
│   ├── aggregator.py
│   ├── database.py
│   ├── server.py
│   ├── theme_detector.py
│   ├── narrative_synthesizer.py
│   ├── feedback.py
│   │
│   │   # NEW IN V3
│   ├── preference_learner.py    # Phase 1
│   ├── trend_detector.py        # Phase 2
│   ├── source_discovery.py      # Phase 3
│   ├── recommender.py           # Phase 4
│   ├── story_tracker.py         # Phase 5
│   └── dashboard.py             # Phase 6
│
├── templates/
│   ├── base.html
│   ├── index.html
│   ├── report.html              # Updated with trends, recommendations, stories
│   ├── priorities.html
│   │
│   │   # NEW IN V3
│   ├── preferences.html         # Phase 1
│   ├── sources.html             # Phase 3
│   └── dashboard.html           # Phase 6
│
├── static/
│   └── style.css                # Extended with new component styles
│
├── data/
│   └── articles.db
│
├── tests/
│   ├── test_preference_learner.py
│   ├── test_trend_detector.py
│   ├── test_source_discovery.py
│   ├── test_recommender.py
│   └── test_story_tracker.py
│
└── Instructions/
    ├── implementation_plan_v2.md
    └── implementation_plan_v3.md  # This document
```

---

## Dependencies Summary

```toml
# pyproject.toml additions for v3

[project.dependencies]
# Existing...
click = "^8.1"
feedparser = "^6.0"
flask = "^3.0"
pyyaml = "^6.0"
python-dotenv = "^1.0"
openai = "^1.0"
httpx = "^0.27"

# NEW for v3
feedfinder2 = "^0.0.4"           # Phase 3: RSS discovery
trafilatura = "^1.6.0"           # Phase 3: Content extraction
sentence-transformers = "^2.2.0" # Phase 5: Embeddings for story tracking
```

---

## Migration Path

| Migration | Phase | Tables Added |
|-----------|-------|--------------|
| v4 | Phase 1 | source_preferences, topic_preferences, keyword_preferences |
| v5 | Phase 2 | weekly_term_frequencies, weekly_trends |
| v6 | Phase 3 | discovered_sources, source_evaluations |
| v7 | Phase 5 | stories, story_articles, story_events |

---

## Success Metrics

| Phase | Metric | Target |
|-------|--------|--------|
| 1 | Preferences page populated | After 10+ votes |
| 2 | Trends section shows data | 3+ emerging topics per week |
| 3 | Sources grow automatically | 5+ new sources discovered/month |
| 4 | Recommendations feel relevant | >50% of "For You" get engagement |
| 5 | Stories tracked | 3+ active stories at any time |
| 6 | Dashboard provides insight | User visits weekly |

---

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Over-personalization (filter bubble) | Explicit serendipity injection (20% explore) |
| Low-quality source discovery | Quality scoring + probation period |
| Story tracker too aggressive | High similarity threshold (0.75) |
| Embedding costs | Use lightweight model (all-MiniLM-L6-v2) |
| Preference model cold start | Show "keep voting to train" messaging |

---

## Next Steps

1. **Implement Phase 1** (Feedback Intelligence) as foundation
2. **Validate learning** with manual testing
3. **Proceed to Phase 2** (Trends) once preferences work
4. **Iterate based on actual usage patterns**

Each phase delivers standalone value while building toward the complete inspiration engine.
