package database

import (
	"database/sql"
	"encoding/json"
)

// InsertTriage inserts or replaces a triage result.
func (db *DB) InsertTriage(articleID int64, verdict string, articleType *string, keyPoints []string, relevanceReason *string, practicalScore int) error {
	var kpJSON *string
	if keyPoints != nil {
		data, err := json.Marshal(keyPoints)
		if err != nil {
			return err
		}
		s := string(data)
		kpJSON = &s
	}

	_, err := db.conn.Exec(
		`INSERT OR REPLACE INTO article_triage
		(article_id, verdict, article_type, key_points, relevance_reason, practical_score)
		VALUES (?, ?, ?, ?, ?, ?)`,
		articleID, verdict, articleType, kpJSON, relevanceReason, practicalScore,
	)
	return err
}

// GetTriage returns the triage result for an article.
func (db *DB) GetTriage(articleID int64) (*ArticleTriage, error) {
	row := db.conn.QueryRow(
		`SELECT article_id, verdict, article_type, key_points, relevance_reason, practical_score, triaged_at
		FROM article_triage WHERE article_id = ?`, articleID,
	)

	var t ArticleTriage
	var kpJSON *string
	if err := row.Scan(&t.ArticleID, &t.Verdict, &t.ArticleType, &kpJSON,
		&t.RelevanceReason, &t.PracticalScore, &t.TriagedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if kpJSON != nil {
		if err := json.Unmarshal([]byte(*kpJSON), &t.KeyPoints); err != nil {
			t.KeyPoints = nil
		}
	}

	return &t, nil
}

// GetTriageStats returns triage statistics for a period.
func (db *DB) GetTriageStats(periodID string) (*TriageStats, error) {
	row := db.conn.QueryRow(
		`SELECT
			COUNT(*) as total,
			SUM(CASE WHEN verdict = 'relevant' THEN 1 ELSE 0 END) as relevant,
			SUM(CASE WHEN verdict = 'skip' THEN 1 ELSE 0 END) as skipped
		FROM article_triage t
		JOIN articles a ON a.id = t.article_id
		WHERE a.period_id = ?`, periodID,
	)

	var s TriageStats
	var relevant, skipped *int
	if err := row.Scan(&s.Total, &relevant, &skipped); err != nil {
		return nil, err
	}
	if relevant != nil {
		s.Relevant = *relevant
	}
	if skipped != nil {
		s.Skipped = *skipped
	}
	return &s, nil
}
