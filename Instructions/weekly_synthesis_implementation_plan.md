# Weekly Synthesis Implementation Plan

## Vision

Transform AICrawler from a simple article aggregator into an intelligent weekly synthesis system that:
- **Identifies emerging themes** across multiple articles automatically
- **Creates coherent narratives** that explain trends, patterns, and key developments
- **Learns from your feedback** to improve theme detection and narrative relevance over time
- **Adapts to your interests** by understanding which themes and explanations matter to you

## User Preferences (Selected)

- **Organization**: By emerging themes/trends (LLM-detected)
- **Feedback Types**: Narrative usefulness + Theme importance + Explain reasoning (full learning)
- **Generation**: Automatically weekly after article summarization

## Architecture Overview

```
Collection → Summarization → Theme Detection → Narrative Synthesis → User Feedback → Learning
    ↓              ↓               ↓                    ↓                    ↓           ↓
 articles     summaries    weekly_themes      weekly_narratives      theme_feedback  Future
  table        table          table               table               table         Filtering
```

## Implementation Phases

### Phase 1: Theme Detection & Storage (Foundation)
**Goal**: LLM identifies recurring themes across articles for a week

### Phase 2: Narrative Synthesis (Core Feature)
**Goal**: Generate coherent narratives for each detected theme

### Phase 3: Feedback Collection (Learning Infrastructure)
**Goal**: Capture multi-dimensional user feedback on narratives and themes

### Phase 4: Adaptive Learning (Intelligence Layer)
**Goal**: Use feedback to improve future theme detection and narrative generation

---

## Phase 1: Theme Detection & Storage

### Database Schema (Migration v2)

**New Table: `weekly_themes`**
```sql
CREATE TABLE weekly_themes (
    id INTEGER PRIMARY KEY,
    week_number TEXT NOT NULL,
    theme_name TEXT NOT NULL,
    theme_description TEXT,
    article_ids TEXT NOT NULL,  -- JSON array of article IDs
    article_count INTEGER,
    confidence_score REAL,      -- 0.0-1.0: LLM's confidence in theme
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(week_number, theme_name)
);

CREATE INDEX idx_themes_week ON weekly_themes(week_number);
CREATE INDEX idx_themes_confidence ON weekly_themes(confidence_score);
```

**Purpose**:
- One record per detected theme per week
- `theme_name`: Short label (e.g., "LLM Cost Optimization", "Agentic Testing Tools")
- `theme_description`: 1-2 sentences explaining the theme
- `article_ids`: JSON array linking to articles that contribute to this theme
- `confidence_score`: LLM's confidence (helps filter weak themes)

### Implementation: ThemeDetector Class

**New File**: `src/theme_detector.py`

```python
class ThemeDetector:
    """Detects emerging themes across weekly articles using LLM."""

    def __init__(self, provider: LLMProvider, db: Database):
        self.provider = provider
        self.db = db

    def detect_themes(self, week_number: str, min_articles: int = 3) -> list[DetectedTheme]:
        """
        Analyze articles for a week and identify recurring themes.

        Args:
            week_number: ISO week to analyze
            min_articles: Minimum articles required to form a theme

        Returns:
            List of DetectedTheme objects with metadata
        """
        # 1. Get all articles for the week
        # 2. Prepare article summaries for LLM
        # 3. Call LLM with theme detection prompt
        # 4. Parse theme response
        # 5. Store themes in weekly_themes table

    def _build_theme_prompt(self, articles: list[Article]) -> str:
        """Construct prompt for theme detection."""
        # Include article titles, summaries, topics
        # Ask LLM to identify 3-5 recurring themes
        # Request confidence scores and explanations
```

### Theme Detection Prompt Design

```
You are analyzing {N} AI/software engineering articles from the week of {date_range}.
Your task is to identify 3-5 RECURRING THEMES or trends across these articles.

A good theme:
- Appears in at least 3 articles
- Represents a meaningful pattern or trend
- Is actionable or relevant to software engineers

ARTICLES:
[For each article]
- Title: {title}
- Summary: {summary}
- Topics: {topics}
- Source: {source}

INSTRUCTIONS:
1. Identify 3-5 recurring themes across these articles
2. For each theme, provide:
   - theme_name: Short label (2-5 words)
   - theme_description: 1-2 sentence explanation
   - article_indices: List of article numbers that relate to this theme
   - confidence: Your confidence this is a genuine theme (0.0-1.0)
   - reasoning: Brief explanation of why this is a theme

OUTPUT FORMAT (JSON):
{
  "themes": [
    {
      "theme_name": "LLM Cost Optimization",
      "theme_description": "Multiple articles discuss techniques and tools for reducing LLM API costs.",
      "article_indices": [0, 3, 7, 12],
      "confidence": 0.9,
      "reasoning": "Four articles explicitly discuss cost reduction strategies"
    },
    ...
  ]
}
```

### Integration Point

**Modified**: `src/cli.py` - Update `run` command

```python
@main.command()
def run():
    """Full pipeline: collect → summarize → detect themes → synthesize."""
    # ... existing collection and summarization ...

    # NEW: Theme detection
    click.echo("Detecting weekly themes...")
    from .theme_detector import ThemeDetector

    detector = ThemeDetector(provider, db)
    themes = detector.detect_themes(current_week, min_articles=3)
    click.echo(f"  Detected {len(themes)} themes")
```

---

## Phase 2: Narrative Synthesis

### Database Schema (Migration v2 continued)

**New Table: `weekly_narratives`**
```sql
CREATE TABLE weekly_narratives (
    id INTEGER PRIMARY KEY,
    week_number TEXT NOT NULL,
    theme_id INTEGER REFERENCES weekly_themes(id),
    narrative_type TEXT DEFAULT 'theme',  -- 'theme', 'executive', 'priority'
    narrative_text TEXT NOT NULL,
    key_points TEXT,                      -- JSON array of bullet points
    article_references TEXT NOT NULL,     -- JSON array of {article_id, relevance}
    generated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    regenerated_count INTEGER DEFAULT 0,
    FOREIGN KEY (theme_id) REFERENCES weekly_themes(id)
);

CREATE INDEX idx_narratives_week ON weekly_narratives(week_number);
CREATE INDEX idx_narratives_theme ON weekly_narratives(theme_id);
```

**Purpose**:
- One narrative per theme (or special types like executive summary)
- `narrative_text`: 2-4 paragraph synthesis explaining the theme
- `key_points`: Bulleted takeaways for scanning
- `article_references`: Which articles contributed and how much
- `narrative_type`: Supports different synthesis modes

### Implementation: NarrativeSynthesizer Class

**New File**: `src/narrative_synthesizer.py`

```python
class NarrativeSynthesizer:
    """Synthesizes coherent narratives from themed articles."""

    def __init__(self, provider: LLMProvider, db: Database):
        self.provider = provider
        self.db = db

    def synthesize_theme_narrative(
        self,
        theme: DetectedTheme,
        articles: list[Article],
        user_context: UserContext
    ) -> Narrative:
        """
        Create a narrative explaining a detected theme.

        Args:
            theme: Detected theme metadata
            articles: Articles that contribute to this theme
            user_context: User preferences and priorities for personalization

        Returns:
            Narrative object with synthesized text and metadata
        """
        # 1. Retrieve full article content for theme articles
        # 2. Build synthesis prompt with theme + articles + user context
        # 3. Call LLM with synthesis prompt
        # 4. Parse narrative response
        # 5. Store in weekly_narratives table
```

### Narrative Synthesis Prompt Design

```
You are synthesizing a weekly narrative for a senior software engineering manager.

THEME: {theme_name}
Description: {theme_description}

CONTEXT: Week of {date_range}, analyzing {N} related articles

USER PRIORITIES:
{list of active research priorities with descriptions}

USER FEEDBACK PATTERNS:
- Prefers insights on: {learned_preferences}
- Less interested in: {negative_patterns}

ARTICLES CONTRIBUTING TO THIS THEME:
[For each article]
---
Title: {title}
Source: {source}
Summary: {summary}
Key points: {extracted_points}
---

YOUR TASK:
Write a 2-4 paragraph narrative that:
1. EXPLAINS the theme and why it's emerging now
2. SYNTHESIZES insights across the articles (don't just list them)
3. HIGHLIGHTS what's most relevant to the user's priorities
4. IDENTIFIES implications for software engineering practice
5. NOTES any contradictions or debates in the articles

TONE: Professional, insightful, action-oriented. Suitable for executive reading.

OUTPUT FORMAT (JSON):
{
  "narrative": "2-4 paragraphs of synthesized text...",
  "key_takeaways": [
    "Bulleted insight 1",
    "Bulleted insight 2",
    "Bulleted insight 3"
  ],
  "priority_relevance": {
    "priority_id_1": "Brief note on how this theme relates",
    "priority_id_2": "Another relevance note"
  },
  "confidence": 0.85,  // How confident you are in this synthesis
  "article_contributions": [
    {"article_id": 123, "contribution": "Provided cost data"},
    {"article_id": 456, "contribution": "Offered implementation examples"}
  ]
}
```

### Display Updates

**Modified**: `templates/report.html`

Add new section **before** individual articles:

```jinja2
{% if week.narratives %}
<section class="synthesis-section">
    <h2>📊 Weekly Synthesis</h2>
    <p class="section-intro">Key themes and trends from this week's articles</p>

    {% for narrative in week.narratives %}
    <article class="narrative-card">
        <header class="narrative-header">
            <h3>{{ narrative.theme_name }}</h3>
            <span class="article-count">Based on {{ narrative.article_count }} articles</span>
        </header>

        <div class="narrative-body">
            {{ narrative.narrative_text|safe }}
        </div>

        {% if narrative.key_takeaways %}
        <div class="key-takeaways">
            <h4>Key Takeaways:</h4>
            <ul>
                {% for takeaway in narrative.key_takeaways %}
                <li>{{ takeaway }}</li>
                {% endfor %}
            </ul>
        </div>
        {% endif %}

        <!-- Feedback controls (added in Phase 3) -->
        <div class="narrative-feedback">
            <!-- Placeholder for feedback UI -->
        </div>

        <details class="narrative-sources">
            <summary>View source articles ({{ narrative.article_count }})</summary>
            <ul class="source-articles">
                {% for article_ref in narrative.article_references %}
                <li>
                    <a href="{{ article_ref.url }}">{{ article_ref.title }}</a>
                    <span class="contribution">{{ article_ref.contribution }}</span>
                </li>
                {% endfor %}
            </ul>
        </details>
    </article>
    {% endfor %}
</section>
{% endif %}
```

**Modified**: `src/aggregator.py`

Update `WeekData` dataclass:

```python
@dataclass
class WeekData:
    week_number: str
    date_range: str
    sections: dict[str, list[ArticleSummary]]
    total_articles: int
    priority_highlights: dict[int, list[ArticleSummary]]
    narratives: list[NarrativeSummary] = field(default_factory=list)  # NEW
```

---

## Phase 3: Feedback Collection

### Database Schema (Migration v3)

**New Table: `theme_feedback`**
```sql
CREATE TABLE theme_feedback (
    id INTEGER PRIMARY KEY,
    theme_id INTEGER REFERENCES weekly_themes(id),
    is_important BOOLEAN,           -- Theme importance toggle
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE narrative_feedback (
    id INTEGER PRIMARY KEY,
    narrative_id INTEGER REFERENCES weekly_narratives(id),
    is_useful BOOLEAN NOT NULL,     -- Useful/not useful toggle
    reasoning_text TEXT,            -- Optional: why useful/not useful
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE INDEX idx_theme_feedback ON theme_feedback(theme_id);
CREATE INDEX idx_narrative_feedback ON narrative_feedback(narrative_id);
```

### UI Components for Feedback

**Updated**: `templates/report.html` - Add to narrative card

```jinja2
<div class="narrative-feedback">
    <!-- Narrative usefulness -->
    <div class="feedback-group">
        <span class="feedback-label">Was this synthesis useful?</span>
        <div class="feedback-buttons">
            <form action="{{ url_for('narrative_feedback', narrative_id=narrative.id, useful=true) }}"
                  method="post" class="inline-form">
                <button type="submit"
                        class="btn-feedback btn-useful {% if narrative.user_feedback_useful %}active{% endif %}"
                        title="Useful synthesis">👍 Useful</button>
            </form>
            <form action="{{ url_for('narrative_feedback', narrative_id=narrative.id, useful=false) }}"
                  method="post" class="inline-form">
                <button type="submit"
                        class="btn-feedback btn-not-useful {% if narrative.user_feedback_useful == false %}active{% endif %}"
                        title="Not useful">👎 Not useful</button>
            </form>
        </div>
    </div>

    <!-- Theme importance -->
    <div class="feedback-group">
        <span class="feedback-label">Theme importance:</span>
        <form action="{{ url_for('theme_importance', theme_id=narrative.theme_id) }}"
              method="post" class="inline-form">
            <button type="submit"
                    class="btn-feedback btn-important {% if narrative.theme_important %}active{% endif %}"
                    title="Mark as important">⭐ Important to me</button>
        </form>
    </div>

    <!-- Optional reasoning -->
    {% if narrative.user_feedback_useful is not none %}
    <details class="feedback-reasoning">
        <summary>Explain why (optional)</summary>
        <form action="{{ url_for('narrative_reasoning', narrative_id=narrative.id) }}"
              method="post" class="reasoning-form">
            <textarea name="reasoning"
                      placeholder="Why was this useful/not useful? This helps improve future syntheses."
                      rows="3">{{ narrative.user_reasoning }}</textarea>
            <button type="submit" class="btn btn-small">Save explanation</button>
        </form>
    </details>
    {% endif %}
</div>
```

### Flask Routes

**Modified**: `src/server.py`

```python
@app.route("/narrative/<int:narrative_id>/feedback/<useful>", methods=["POST"])
def narrative_feedback(narrative_id: int, useful: str):
    """Record whether narrative was useful."""
    db = get_db()
    is_useful = useful.lower() == 'true'
    db.set_narrative_feedback(narrative_id, is_useful=is_useful)

    # Redirect back to report
    narrative = db.get_narrative(narrative_id)
    return redirect(url_for("report", week=narrative.week_number))

@app.route("/theme/<int:theme_id>/importance", methods=["POST"])
def theme_importance(theme_id: int):
    """Toggle theme importance."""
    db = get_db()
    db.toggle_theme_importance(theme_id)

    theme = db.get_theme(theme_id)
    return redirect(url_for("report", week=theme.week_number))

@app.route("/narrative/<int:narrative_id>/reasoning", methods=["POST"])
def narrative_reasoning(narrative_id: int):
    """Save explanation for feedback."""
    db = get_db()
    reasoning = request.form.get("reasoning", "").strip()
    db.set_narrative_reasoning(narrative_id, reasoning_text=reasoning)

    narrative = db.get_narrative(narrative_id)
    return redirect(url_for("report", week=narrative.week_number))
```

---

## Phase 4: Adaptive Learning

### Learning Loop Architecture

```
Feedback Collection → Pattern Analysis → Prompt Adaptation → Future Synthesis
       ↓                     ↓                    ↓                  ↓
  theme_feedback      FeedbackAnalyzer     Enhanced prompts    Better themes
  narrative_feedback      patterns         User context        Better narratives
```

### Implementation: Enhanced FeedbackAnalyzer

**Modified**: `src/feedback.py`

```python
class FeedbackAnalyzer:
    """Analyzes user feedback to improve filtering, themes, and narratives."""

    # ... existing article feedback methods ...

    def get_theme_preferences(self) -> ThemePreferences:
        """
        Analyze which types of themes user finds important.

        Returns:
            ThemePreferences with preferred theme patterns
        """
        # Query theme_feedback table
        # Identify patterns in important vs unimportant themes
        # Return: preferred_keywords, preferred_topics, theme_patterns

    def get_narrative_quality_patterns(self) -> NarrativePatterns:
        """
        Analyze what makes narratives useful vs not useful.

        Returns:
            NarrativePatterns with quality indicators
        """
        # Query narrative_feedback and reasoning_text
        # Use LLM to analyze reasoning patterns
        # Identify: preferred_length, preferred_style, content_preferences

    def build_user_context(self) -> UserContext:
        """
        Compile comprehensive user context for personalized synthesis.

        Returns:
            UserContext for passing to synthesis prompts
        """
        return UserContext(
            theme_preferences=self.get_theme_preferences(),
            narrative_patterns=self.get_narrative_quality_patterns(),
            article_preferences=self.get_feedback_stats(),  # existing
            priorities=self.db.get_active_priorities()
        )
```

### Enhanced Theme Detection with Learning

**Modified**: `src/theme_detector.py`

```python
def detect_themes(
    self,
    week_number: str,
    user_context: UserContext
) -> list[DetectedTheme]:
    """Detect themes with user preference awareness."""

    # Build enhanced prompt with user context
    prompt = self._build_enhanced_theme_prompt(articles, user_context)

    # Prompt now includes:
    # - "User finds these themes important: {learned_themes}"
    # - "User prefers themes related to: {priorities}"
    # - "User has shown interest in: {positive_keywords}"
```

### Reasoning-Based Improvement

**New File**: `src/reasoning_analyzer.py`

```python
class ReasoningAnalyzer:
    """Analyzes user's explanatory text to extract improvement signals."""

    def analyze_reasoning_batch(
        self,
        reasoning_texts: list[str],
        feedback_type: str  # 'useful' or 'not_useful'
    ) -> ReasoningInsights:
        """
        Use LLM to analyze user reasoning and extract patterns.

        Example prompt:
        "Analyze these explanations for why syntheses were USEFUL:
        - 'Helped me see the connection between cost and latency'
        - 'Clear actionable recommendations'
        - 'Linked to our current project priorities'

        Extract patterns about what users value in syntheses."
        """
        # LLM analyzes reasoning texts
        # Returns: common_themes, valued_aspects, improvement_suggestions
```

### Adaptive Prompt Construction

The system builds increasingly personalized prompts:

**Week 1** (no feedback):
```
Standard theme detection prompt
Standard synthesis prompt
```

**Week 5** (some feedback):
```
Theme detection prompt + "User values themes about: {learned_patterns}"
Synthesis prompt + "User prefers narratives that: {quality_patterns}"
```

**Week 20** (rich feedback):
```
Theme detection prompt + full user context
Synthesis prompt + detailed personalization
+ "User reasoning: when syntheses are useful, they {patterns}"
+ "Avoid: {anti-patterns from negative feedback}"
```

---

## Integration with Existing Feedback System

### Synergies with Article Feedback

The article-level feedback system (upvote/downvote) and narrative feedback work together:

1. **Article feedback** signals → Theme importance
   - If user upvotes many articles in a theme → theme is likely important
   - Theme detector can weight themes higher if articles have positive feedback

2. **Theme feedback** signals → Article filtering
   - If user marks theme as important → boost similar articles in future
   - Extract keywords from important themes → add to filter weights

3. **Narrative reasoning** signals → Everything
   - Rich text explanations guide both theme detection AND article filtering
   - Example: "I need more implementation details, not theory" → adjust both

### Cross-Feature Learning

```python
class UnifiedFeedbackSystem:
    """Coordinates learning across article, theme, and narrative feedback."""

    def synthesize_learning_signals(self) -> LearningSignals:
        """Combine all feedback sources into unified learning signals."""

        article_signals = self.article_feedback_analyzer.get_patterns()
        theme_signals = self.theme_feedback_analyzer.get_patterns()
        narrative_signals = self.narrative_feedback_analyzer.get_patterns()

        # Find alignments and conflicts
        # Example: User upvotes articles about "testing" (article feedback)
        #          User marks "Testing Automation" theme as important (theme feedback)
        #          User says narratives are useful when they show "practical examples" (reasoning)
        # → Future: Prioritize testing articles with practical examples in theme narratives

        return unified_signals
```

---

## CLI Commands

### New Commands

```bash
# Generate themes for current week
aicrawler synthesize themes

# Generate narratives for current week
aicrawler synthesize narratives

# Regenerate with updated preferences
aicrawler synthesize regenerate --week 2026-W04

# View synthesis stats
aicrawler synthesize stats

# Analyze feedback patterns
aicrawler feedback analyze-reasoning
```

### Updated Run Command

```bash
aicrawler run
# Now includes: collect → summarize → detect-themes → synthesize-narratives
```

---

## Database Migration Roadmap

### Migration v2: Core Synthesis Tables
- `weekly_themes`
- `weekly_narratives`

### Migration v3: Feedback Tables
- `theme_feedback`
- `narrative_feedback`

### Migration v4 (Future): Learning Cache
- `theme_patterns` - Cached theme preferences
- `narrative_quality_metrics` - Cached quality patterns

---

## Performance Considerations

### LLM Cost Management

1. **Theme Detection**: ~1 LLM call per week
   - Input: All article summaries (~500 tokens per article × 50 articles = 25K tokens)
   - Output: ~2K tokens
   - Cost (GPT-4o-mini): ~$0.05/week

2. **Narrative Synthesis**: ~1 LLM call per theme (3-5 themes)
   - Input per theme: ~10 articles × 1000 tokens = 10K tokens
   - Output: ~1K tokens per narrative
   - Cost: ~$0.15/week total

3. **Reasoning Analysis**: ~1 LLM call per 20 feedback items (batch)
   - Input: 20 reasoning texts × 100 tokens = 2K tokens
   - Output: ~500 tokens
   - Cost: ~$0.01/batch

**Total estimated cost**: ~$0.25/week with GPT-4o-mini (~$1/month)

### Caching Strategy

- Cache theme detection results (don't regenerate unless articles change)
- Cache narrative synthesis (only regenerate on user request or feedback threshold)
- Cache user context (rebuild weekly, not per-prompt)

---

## Success Metrics

### Engagement Metrics
- **Theme feedback rate**: % of themes receiving importance feedback
- **Narrative feedback rate**: % of narratives receiving useful/not useful feedback
- **Reasoning completion rate**: % of feedback with explanatory text

### Quality Metrics
- **Useful narrative ratio**: useful / (useful + not useful) feedback
- **Important theme ratio**: important themes / total themes
- **Feedback consistency**: Do reasoning patterns align with boolean feedback?

### Learning Metrics
- **Theme relevance improvement**: Are important themes detected earlier over time?
- **Narrative quality trend**: Is useful ratio increasing?
- **Personalization accuracy**: Do regenerated narratives get better feedback?

---

## Future Enhancements (Phase 5+)

### Multi-Week Trend Analysis
- Detect themes that persist across multiple weeks
- "This theme has been emerging for 3 weeks..."

### Comparative Narratives
- Compare current week to previous weeks
- "Cost optimization has become more prominent vs last month"

### Priority-Focused Narratives
- Generate special narratives for each research priority
- "How this week's articles relate to your LLM Agent testing priority"

### Interactive Regeneration
- "Regenerate this narrative with more technical depth"
- "Focus this synthesis on implementation details"

### Collaborative Filtering
- If multiple users: "Others with similar interests found these themes important"

### Export & Sharing
- Generate PDF weekly reports
- Email digest with top narratives
- Slack/Discord integration for team sharing

---

## Testing Strategy

### Unit Tests
- `test_theme_detector.py`: Theme detection logic
- `test_narrative_synthesizer.py`: Narrative generation
- `test_feedback_analyzer.py`: Pattern extraction from feedback

### Integration Tests
- `test_synthesis_pipeline.py`: End-to-end theme → narrative flow
- `test_feedback_loop.py`: Feedback collection → learning application

### Manual Testing Checklist
1. ✓ Themes detected for week with 20+ articles
2. ✓ Narratives generated for each theme
3. ✓ Feedback buttons appear and function
4. ✓ Reasoning text saves correctly
5. ✓ User context builds from feedback
6. ✓ Future syntheses reflect learned preferences

---

## Implementation Timeline

### Week 1-2: Phase 1 (Theme Detection)
- Migration v2: Add weekly_themes table
- Implement ThemeDetector class
- Test theme detection on historical weeks
- Integrate into CLI run command

### Week 3-4: Phase 2 (Narrative Synthesis)
- Add weekly_narratives table
- Implement NarrativeSynthesizer class
- Update report.html template
- Update aggregator to load narratives
- Test narrative display

### Week 5-6: Phase 3 (Feedback Collection)
- Migration v3: Add feedback tables
- Implement feedback routes
- Add feedback UI to templates
- Test feedback capture

### Week 7-8: Phase 4 (Learning)
- Enhance FeedbackAnalyzer
- Implement ReasoningAnalyzer
- Build user context system
- Test adaptive prompts
- Verify learning loop works

### Week 9+: Refinement & Future Features
- Performance optimization
- Additional narrative types
- Advanced learning algorithms
- Multi-week trend analysis

---

## File Structure After Implementation

```
src/
├── theme_detector.py          # NEW: Theme detection
├── narrative_synthesizer.py   # NEW: Narrative generation
├── reasoning_analyzer.py      # NEW: Reasoning analysis
├── feedback.py                # MODIFIED: Enhanced with theme/narrative learning
├── aggregator.py              # MODIFIED: Include narratives in WeekData
├── database.py                # MODIFIED: New tables and CRUD methods
├── server.py                  # MODIFIED: New feedback routes
├── cli.py                     # MODIFIED: New synthesis commands
└── ... (existing files)

templates/
├── report.html                # MODIFIED: Add synthesis section
└── ... (existing templates)

Instructions/
├── weekly_synthesis_implementation_plan.md  # THIS FILE
├── schema_migrations.md                     # Reference for migrations
└── implementation_plan_v2.md                # Original architecture docs
```

---

## Key Design Decisions

### Why Theme-Based Organization?
- More intelligent than fixed topics
- Captures emerging trends that don't fit predefined categories
- Adapts to what's actually happening in the field
- User feedback shapes future theme detection

### Why Three Feedback Types?
- **Narrative usefulness**: Quick binary signal for quality
- **Theme importance**: Guides future theme detection
- **Explain reasoning**: Rich signal for deep learning (optional but powerful)

### Why Automatic Generation?
- Always have fresh synthesis when you check weekly report
- Reduces friction (no manual "synthesize" button needed)
- Enables consistent feedback loop (always see results of learning)

### Why Not Replace Individual Articles?
- Synthesis provides high-level overview
- Individual articles provide depth and details
- Users can scan narratives, then dive into articles of interest
- Best of both worlds: breadth + depth

---

## Migration Path for Existing Data

### Handling Historical Weeks
- Phase 1-2 implementation: New tables are empty
- Historical weeks show individual articles only (no synthesis)
- Can optionally backfill: `aicrawler synthesize backfill --weeks 2026-W01:2026-W10`

### Gradual Rollout
1. Deploy Phase 1-2: Synthesis appears but no feedback yet
2. Gather initial user reactions
3. Deploy Phase 3: Add feedback UI
4. Wait for feedback data to accumulate (2-4 weeks)
5. Deploy Phase 4: Learning kicks in with sufficient data

---

## Next Steps for Implementation

### For Next Session: Start with Phase 1

1. **Read existing schema**: Review current database.py migration system
2. **Design Migration v2**: Write `_migration_v2_synthesis_tables()`
3. **Implement ThemeDetector**: Core theme detection logic
4. **Write theme detection prompt**: Test with sample articles
5. **Test on historical data**: Verify themes make sense for past weeks

### Commands to Run

```bash
# Create feature branch
git checkout -b feature/weekly-synthesis

# Start implementation
cd src/
# Create new files: theme_detector.py

# Test theme detection
python3 -c "from src.theme_detector import ThemeDetector; ..."

# View generated themes
aicrawler synthesize themes --week 2026-W04
```

---

## Questions for Future Sessions

1. Should themes be editable by user? (rename/merge/split)
2. Should narratives support different "styles"? (executive, technical, brief)
3. How many weeks of history to consider for learning? (weight recent higher?)
4. Should there be a "master narrative" synthesizing all themes for the week?
5. Integration with external tools? (Notion, Obsidian, email)

---

## References

- Current architecture: [src/aggregator.py](../src/aggregator.py)
- LLM integration: [src/summarizer.py](../src/summarizer.py)
- Feedback system: [src/feedback.py](../src/feedback.py)
- Migration system: [Instructions/schema_migrations.md](schema_migrations.md)
- Original plan: [Instructions/implementation_plan_v2.md](implementation_plan_v2.md)
