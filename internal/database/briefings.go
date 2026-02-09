package database

import (
	"database/sql"
	"strings"
)

// InsertBriefing inserts or replaces a briefing for a period.
func (db *DB) InsertBriefing(periodID, tldr, bodyMarkdown string, storylineCount, articleCount int) (int64, error) {
	result, err := db.conn.Exec(
		`INSERT OR REPLACE INTO briefings
		(period_id, tldr, body_markdown, storyline_count, article_count)
		VALUES (?, ?, ?, ?, ?)`,
		periodID, tldr, bodyMarkdown, storylineCount, articleCount,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetBriefing returns the briefing for a period.
func (db *DB) GetBriefing(periodID string) (*Briefing, error) {
	row := db.conn.QueryRow(
		`SELECT id, period_id, tldr, body_markdown, storyline_count, article_count, generated_at
		FROM briefings WHERE period_id = ?`, periodID,
	)

	var b Briefing
	if err := row.Scan(&b.ID, &b.PeriodID, &b.TLDR, &b.BodyMarkdown,
		&b.StorylineCount, &b.ArticleCount, &b.GeneratedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &b, nil
}

// GetAllBriefings returns all briefings ordered by period_id DESC.
func (db *DB) GetAllBriefings() ([]Briefing, error) {
	rows, err := db.conn.Query(
		"SELECT id, period_id, tldr, body_markdown, storyline_count, article_count, generated_at FROM briefings ORDER BY period_id DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var briefings []Briefing
	for rows.Next() {
		var b Briefing
		if err := rows.Scan(&b.ID, &b.PeriodID, &b.TLDR, &b.BodyMarkdown,
			&b.StorylineCount, &b.ArticleCount, &b.GeneratedAt); err != nil {
			return nil, err
		}
		briefings = append(briefings, b)
	}
	return briefings, rows.Err()
}

// InsertReport inserts or replaces a run report.
func (db *DB) InsertReport(periodID string, articleCount, storylineCount int) (int64, error) {
	result, err := db.conn.Exec(
		`INSERT OR REPLACE INTO run_reports (period_id, article_count, storyline_count)
		VALUES (?, ?, ?)`,
		periodID, articleCount, storylineCount,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetLastRunDate returns the end date from the most recent run report.
// Returns empty string if no runs exist.
func (db *DB) GetLastRunDate() (string, error) {
	row := db.conn.QueryRow(
		"SELECT period_id FROM run_reports ORDER BY period_id DESC LIMIT 1",
	)

	var periodID string
	if err := row.Scan(&periodID); err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}

	// Range format: "YYYY-MM-DD..YYYY-MM-DD" — return end date
	if strings.Contains(periodID, "..") {
		parts := strings.SplitN(periodID, "..", 2)
		if len(parts) == 2 {
			return parts[1], nil
		}
	}
	return periodID, nil
}

// GetStats returns aggregate database statistics.
func (db *DB) GetStats() (*Stats, error) {
	s := &Stats{}

	queries := []struct {
		sql  string
		dest *int
	}{
		{"SELECT COUNT(*) FROM articles", &s.TotalArticles},
		{"SELECT COUNT(*) FROM article_triage", &s.TriagedArticles},
		{"SELECT COUNT(*) FROM article_triage WHERE verdict = 'relevant'", &s.RelevantArticles},
		{"SELECT COUNT(DISTINCT period_id) FROM articles", &s.PeriodsWithArticles},
		{"SELECT COUNT(*) FROM briefings", &s.Briefings},
		{"SELECT COUNT(*) FROM storylines", &s.Storylines},
		{"SELECT COUNT(*) FROM research_priorities", &s.TotalPriorities},
		{"SELECT COUNT(*) FROM research_priorities WHERE is_active = 1", &s.ActivePriorities},
	}

	for _, q := range queries {
		if err := db.conn.QueryRow(q.sql).Scan(q.dest); err != nil {
			return nil, err
		}
	}

	return s, nil
}
