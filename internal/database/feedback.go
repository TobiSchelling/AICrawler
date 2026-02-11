package database

import "database/sql"

// UpsertStorylineFeedback inserts or updates feedback for a storyline.
func (db *DB) UpsertStorylineFeedback(storylineID int64, periodID, rating string) error {
	_, err := db.conn.Exec(
		`INSERT OR REPLACE INTO storyline_feedback (storyline_id, period_id, rating) VALUES (?, ?, ?)`,
		storylineID, periodID, rating,
	)
	return err
}

// DeleteStorylineFeedback removes feedback for a storyline (toggle off).
func (db *DB) DeleteStorylineFeedback(storylineID int64) error {
	_, err := db.conn.Exec(`DELETE FROM storyline_feedback WHERE storyline_id = ?`, storylineID)
	return err
}

// GetStorylineFeedback returns feedback for a single storyline.
func (db *DB) GetStorylineFeedback(storylineID int64) (*StorylineFeedback, error) {
	row := db.conn.QueryRow(
		`SELECT storyline_id, period_id, rating, created_at FROM storyline_feedback WHERE storyline_id = ?`,
		storylineID,
	)
	var f StorylineFeedback
	if err := row.Scan(&f.StorylineID, &f.PeriodID, &f.Rating, &f.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &f, nil
}

// GetStorylineFeedbackMap returns a map of storyline_id → rating for a period.
func (db *DB) GetStorylineFeedbackMap(periodID string) (map[int64]string, error) {
	rows, err := db.conn.Query(
		`SELECT storyline_id, rating FROM storyline_feedback WHERE period_id = ?`, periodID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	m := make(map[int64]string)
	for rows.Next() {
		var id int64
		var rating string
		if err := rows.Scan(&id, &rating); err != nil {
			return nil, err
		}
		m[id] = rating
	}
	return m, rows.Err()
}

// UpsertArticleFeedback inserts or updates feedback for an article.
func (db *DB) UpsertArticleFeedback(articleID int64, rating string) error {
	_, err := db.conn.Exec(
		`INSERT OR REPLACE INTO article_feedback (article_id, rating) VALUES (?, ?)`,
		articleID, rating,
	)
	return err
}

// DeleteArticleFeedback removes feedback for an article (toggle off).
func (db *DB) DeleteArticleFeedback(articleID int64) error {
	_, err := db.conn.Exec(`DELETE FROM article_feedback WHERE article_id = ?`, articleID)
	return err
}

// GetArticleFeedback returns feedback for a single article.
func (db *DB) GetArticleFeedback(articleID int64) (*ArticleFeedback, error) {
	row := db.conn.QueryRow(
		`SELECT article_id, rating, created_at FROM article_feedback WHERE article_id = ?`,
		articleID,
	)
	var f ArticleFeedback
	if err := row.Scan(&f.ArticleID, &f.Rating, &f.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &f, nil
}

// GetArticleFeedbackMap returns a map of article_id → rating for a set of article IDs.
func (db *DB) GetArticleFeedbackMap(articleIDs []int64) (map[int64]string, error) {
	if len(articleIDs) == 0 {
		return make(map[int64]string), nil
	}

	// Build query with placeholders
	query := "SELECT article_id, rating FROM article_feedback WHERE article_id IN (?" +
		repeatString(",?", len(articleIDs)-1) + ")"

	args := make([]any, len(articleIDs))
	for i, id := range articleIDs {
		args[i] = id
	}

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	m := make(map[int64]string)
	for rows.Next() {
		var id int64
		var rating string
		if err := rows.Scan(&id, &rating); err != nil {
			return nil, err
		}
		m[id] = rating
	}
	return m, rows.Err()
}

// GetFeedbackSummary aggregates all feedback for triage prompt injection.
func (db *DB) GetFeedbackSummary() (*FeedbackSummary, error) {
	summary := &FeedbackSummary{}

	// Source feedback: join article_feedback with articles to group by source
	sourceRows, err := db.conn.Query(`
		SELECT COALESCE(a.source, 'Unknown') as source,
			SUM(CASE WHEN af.rating = 'positive' THEN 1 ELSE 0 END) as positive,
			SUM(CASE WHEN af.rating = 'negative' THEN 1 ELSE 0 END) as negative
		FROM article_feedback af
		JOIN articles a ON a.id = af.article_id
		GROUP BY COALESCE(a.source, 'Unknown')
		HAVING positive > 0 OR negative > 0
		ORDER BY (positive - negative) DESC`)
	if err != nil {
		return nil, err
	}
	defer sourceRows.Close()

	for sourceRows.Next() {
		var sf SourceFeedback
		if err := sourceRows.Scan(&sf.Source, &sf.Positive, &sf.Negative); err != nil {
			return nil, err
		}
		summary.Sources = append(summary.Sources, sf)
	}
	if err := sourceRows.Err(); err != nil {
		return nil, err
	}

	// Type feedback: join article_feedback with article_triage to group by article_type
	typeRows, err := db.conn.Query(`
		SELECT COALESCE(at.article_type, 'other') as article_type,
			SUM(CASE WHEN af.rating = 'positive' THEN 1 ELSE 0 END) as positive,
			SUM(CASE WHEN af.rating = 'negative' THEN 1 ELSE 0 END) as negative
		FROM article_feedback af
		JOIN article_triage at ON at.article_id = af.article_id
		GROUP BY COALESCE(at.article_type, 'other')
		HAVING positive > 0 OR negative > 0
		ORDER BY (positive - negative) DESC`)
	if err != nil {
		return nil, err
	}
	defer typeRows.Close()

	for typeRows.Next() {
		var tf TypeFeedback
		if err := typeRows.Scan(&tf.ArticleType, &tf.Positive, &tf.Negative); err != nil {
			return nil, err
		}
		summary.Types = append(summary.Types, tf)
	}
	return summary, typeRows.Err()
}

func repeatString(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
