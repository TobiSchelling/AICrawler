package database

import (
	"database/sql"
	"encoding/json"
)

// InsertStoryline creates a storyline and links it to articles.
func (db *DB) InsertStoryline(periodID, label string, articleIDs []int64) (int64, error) {
	tx, err := db.conn.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	result, err := tx.Exec(
		`INSERT INTO storylines (period_id, label, article_count) VALUES (?, ?, ?)`,
		periodID, label, len(articleIDs),
	)
	if err != nil {
		return 0, err
	}

	storylineID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	for _, aid := range articleIDs {
		if _, err := tx.Exec(
			"INSERT INTO storyline_articles (storyline_id, article_id) VALUES (?, ?)",
			storylineID, aid,
		); err != nil {
			return 0, err
		}
	}

	return storylineID, tx.Commit()
}

// GetStorylinesForPeriod returns storylines ordered by article_count DESC.
func (db *DB) GetStorylinesForPeriod(periodID string) ([]Storyline, error) {
	rows, err := db.conn.Query(
		`SELECT id, period_id, label, article_count, created_at
		FROM storylines WHERE period_id = ? ORDER BY article_count DESC`, periodID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var storylines []Storyline
	for rows.Next() {
		var s Storyline
		if err := rows.Scan(&s.ID, &s.PeriodID, &s.Label, &s.ArticleCount, &s.CreatedAt); err != nil {
			return nil, err
		}
		storylines = append(storylines, s)
	}
	return storylines, rows.Err()
}

// GetStorylineArticleIDs returns the article IDs linked to a storyline.
func (db *DB) GetStorylineArticleIDs(storylineID int64) ([]int64, error) {
	rows, err := db.conn.Query(
		"SELECT article_id FROM storyline_articles WHERE storyline_id = ?", storylineID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// GetStorylineArticles returns the full articles linked to a storyline.
func (db *DB) GetStorylineArticles(storylineID int64) ([]Article, error) {
	rows, err := db.conn.Query(
		`SELECT a.id, a.url, a.title, a.source, a.published_date, a.content,
		a.content_fetched, a.period_id, a.collected_at
		FROM articles a JOIN storyline_articles sa ON a.id = sa.article_id
		WHERE sa.storyline_id = ?`, storylineID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanArticles(rows)
}

// ClearStorylinesForPeriod removes existing storylines for re-clustering.
func (db *DB) ClearStorylinesForPeriod(periodID string) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	rows, err := tx.Query("SELECT id FROM storylines WHERE period_id = ?", periodID)
	if err != nil {
		return err
	}

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		ids = append(ids, id)
	}
	rows.Close()

	for _, id := range ids {
		if _, err := tx.Exec("DELETE FROM storyline_articles WHERE storyline_id = ?", id); err != nil {
			return err
		}
		if _, err := tx.Exec("DELETE FROM storyline_narratives WHERE storyline_id = ?", id); err != nil {
			return err
		}
	}

	if _, err := tx.Exec("DELETE FROM storylines WHERE period_id = ?", periodID); err != nil {
		return err
	}

	return tx.Commit()
}

// InsertStorylineNarrative inserts a narrative for a storyline.
func (db *DB) InsertStorylineNarrative(storylineID int64, periodID, title, narrativeText string, sourceRefs []SourceReference) (int64, error) {
	var refsJSON *string
	if sourceRefs != nil {
		data, err := json.Marshal(sourceRefs)
		if err != nil {
			return 0, err
		}
		s := string(data)
		refsJSON = &s
	}

	result, err := db.conn.Exec(
		`INSERT INTO storyline_narratives
		(storyline_id, period_id, title, narrative_text, source_references)
		VALUES (?, ?, ?, ?, ?)`,
		storylineID, periodID, title, narrativeText, refsJSON,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetNarrativesForPeriod returns narratives ordered by storyline article_count DESC.
func (db *DB) GetNarrativesForPeriod(periodID string) ([]StorylineNarrative, error) {
	rows, err := db.conn.Query(
		`SELECT sn.id, sn.storyline_id, sn.period_id, sn.title, sn.narrative_text,
		sn.source_references, sn.generated_at
		FROM storyline_narratives sn
		JOIN storylines s ON s.id = sn.storyline_id
		WHERE sn.period_id = ?
		ORDER BY s.article_count DESC`, periodID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanNarratives(rows)
}

// GetNarrativeForStoryline returns the narrative for a specific storyline.
func (db *DB) GetNarrativeForStoryline(storylineID int64) (*StorylineNarrative, error) {
	row := db.conn.QueryRow(
		`SELECT id, storyline_id, period_id, title, narrative_text, source_references, generated_at
		FROM storyline_narratives WHERE storyline_id = ?`, storylineID,
	)

	var n StorylineNarrative
	var refsJSON *string
	if err := row.Scan(&n.ID, &n.StorylineID, &n.PeriodID, &n.Title,
		&n.NarrativeText, &refsJSON, &n.GeneratedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if refsJSON != nil {
		if err := json.Unmarshal([]byte(*refsJSON), &n.SourceReferences); err != nil {
			n.SourceReferences = nil
		}
	}

	return &n, nil
}

func scanNarratives(rows *sql.Rows) ([]StorylineNarrative, error) {
	var narratives []StorylineNarrative
	for rows.Next() {
		var n StorylineNarrative
		var refsJSON *string
		if err := rows.Scan(&n.ID, &n.StorylineID, &n.PeriodID, &n.Title,
			&n.NarrativeText, &refsJSON, &n.GeneratedAt); err != nil {
			return nil, err
		}
		if refsJSON != nil {
			if err := json.Unmarshal([]byte(*refsJSON), &n.SourceReferences); err != nil {
				n.SourceReferences = nil
			}
		}
		narratives = append(narratives, n)
	}
	return narratives, rows.Err()
}
