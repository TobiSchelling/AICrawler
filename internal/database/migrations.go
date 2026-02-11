package database

import "database/sql"

// Migration represents a single schema migration step.
type Migration struct {
	Version     int
	Description string
	Up          func(tx *sql.Tx) error
}

// migrations is the ordered list of all schema migrations.
// Append new migrations to the end with incrementing Version numbers.
var migrations = []Migration{
	{
		Version:     1,
		Description: "initial schema",
		Up: func(tx *sql.Tx) error {
			_, err := tx.Exec(`
CREATE TABLE IF NOT EXISTS articles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    url TEXT UNIQUE NOT NULL,
    title TEXT NOT NULL,
    source TEXT,
    published_date TEXT,
    content TEXT,
    content_fetched INTEGER DEFAULT 0,
    period_id TEXT,
    collected_at TEXT DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS article_triage (
    article_id INTEGER PRIMARY KEY REFERENCES articles(id),
    verdict TEXT NOT NULL,
    article_type TEXT,
    key_points TEXT,
    relevance_reason TEXT,
    practical_score INTEGER DEFAULT 0,
    triaged_at TEXT DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS storylines (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    period_id TEXT NOT NULL,
    label TEXT NOT NULL,
    article_count INTEGER DEFAULT 0,
    created_at TEXT DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS storyline_articles (
    storyline_id INTEGER NOT NULL REFERENCES storylines(id),
    article_id INTEGER NOT NULL REFERENCES articles(id),
    PRIMARY KEY (storyline_id, article_id)
);

CREATE TABLE IF NOT EXISTS storyline_narratives (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    storyline_id INTEGER NOT NULL REFERENCES storylines(id),
    period_id TEXT NOT NULL,
    title TEXT NOT NULL,
    narrative_text TEXT NOT NULL,
    source_references TEXT,
    generated_at TEXT DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS briefings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    period_id TEXT UNIQUE NOT NULL,
    tldr TEXT NOT NULL,
    body_markdown TEXT NOT NULL,
    storyline_count INTEGER DEFAULT 0,
    article_count INTEGER DEFAULT 0,
    generated_at TEXT DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS research_priorities (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    description TEXT,
    keywords TEXT,
    is_active INTEGER DEFAULT 1,
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS run_reports (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    period_id TEXT UNIQUE NOT NULL,
    generated_at TEXT DEFAULT (datetime('now')),
    article_count INTEGER DEFAULT 0,
    storyline_count INTEGER DEFAULT 0
);

CREATE TABLE IF NOT EXISTS storyline_feedback (
    storyline_id INTEGER PRIMARY KEY REFERENCES storylines(id),
    period_id TEXT NOT NULL,
    rating TEXT NOT NULL CHECK(rating IN ('useful', 'not_useful')),
    created_at TEXT DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS article_feedback (
    article_id INTEGER PRIMARY KEY REFERENCES articles(id),
    rating TEXT NOT NULL CHECK(rating IN ('positive', 'negative')),
    created_at TEXT DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_articles_period ON articles(period_id);
CREATE INDEX IF NOT EXISTS idx_articles_url ON articles(url);
CREATE INDEX IF NOT EXISTS idx_storylines_period ON storylines(period_id);
CREATE INDEX IF NOT EXISTS idx_storyline_narratives_period ON storyline_narratives(period_id);
CREATE INDEX IF NOT EXISTS idx_briefings_period ON briefings(period_id);
CREATE INDEX IF NOT EXISTS idx_storyline_feedback_period ON storyline_feedback(period_id);
`)
			return err
		},
	},
}

// latestVersion returns the highest migration version number.
func latestVersion() int {
	if len(migrations) == 0 {
		return 0
	}
	return migrations[len(migrations)-1].Version
}
