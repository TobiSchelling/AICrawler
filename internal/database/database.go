package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const schema = `
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
`

// DB wraps a SQLite database connection.
type DB struct {
	conn *sql.DB
	path string
}

// Open creates or opens a SQLite database at the given path.
func Open(dbPath string) (*DB, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("creating data directory: %w", err)
	}

	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	if _, err := conn.Exec("PRAGMA journal_mode=WAL"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("setting journal mode: %w", err)
	}
	if _, err := conn.Exec("PRAGMA foreign_keys=ON"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("enabling foreign keys: %w", err)
	}

	if _, err := conn.Exec(schema); err != nil {
		conn.Close()
		return nil, fmt.Errorf("initializing schema: %w", err)
	}

	return &DB{conn: conn, path: dbPath}, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.conn.Close()
}

// Path returns the database file path.
func (db *DB) Path() string {
	return db.path
}
