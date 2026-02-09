package database

import (
	"database/sql"
)

// InsertArticle inserts an article. Returns the ID on success, 0 if duplicate.
func (db *DB) InsertArticle(url, title string, source, publishedDate, content, periodID *string) (int64, error) {
	result, err := db.conn.Exec(
		`INSERT INTO articles (url, title, source, published_date, content, period_id)
		VALUES (?, ?, ?, ?, ?, ?)`,
		url, title, source, publishedDate, content, periodID,
	)
	if err != nil {
		// Duplicate URL constraint
		return 0, nil //nolint: nilerr
	}
	return result.LastInsertId()
}

// GetArticlesForPeriod returns articles for a given period, ordered by collected_at DESC.
func (db *DB) GetArticlesForPeriod(periodID string) ([]Article, error) {
	rows, err := db.conn.Query(
		`SELECT id, url, title, source, published_date, content, content_fetched, period_id, collected_at
		FROM articles WHERE period_id = ? ORDER BY collected_at DESC`, periodID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanArticles(rows)
}

// GetArticlesNeedingFetch returns articles with empty content that haven't been fetched.
func (db *DB) GetArticlesNeedingFetch(periodID *string) ([]Article, error) {
	query := `SELECT id, url, title, source, published_date, content, content_fetched, period_id, collected_at
		FROM articles WHERE (content IS NULL OR content = '') AND content_fetched = 0`
	var args []any
	if periodID != nil {
		query += " AND period_id = ?"
		args = append(args, *periodID)
	}
	query += " ORDER BY collected_at DESC"

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanArticles(rows)
}

// UpdateArticleContent updates article content after fetching.
func (db *DB) UpdateArticleContent(articleID int64, content *string) error {
	_, err := db.conn.Exec(
		"UPDATE articles SET content = ?, content_fetched = 1 WHERE id = ?",
		content, articleID,
	)
	return err
}

// MarkArticleFetchAttempted marks that we tried to fetch content.
func (db *DB) MarkArticleFetchAttempted(articleID int64) error {
	_, err := db.conn.Exec(
		"UPDATE articles SET content_fetched = 1 WHERE id = ?", articleID,
	)
	return err
}

// GetUntriagedArticles returns articles that haven't been triaged yet.
func (db *DB) GetUntriagedArticles(periodID *string) ([]Article, error) {
	query := `SELECT a.id, a.url, a.title, a.source, a.published_date, a.content,
		a.content_fetched, a.period_id, a.collected_at
		FROM articles a LEFT JOIN article_triage t ON a.id = t.article_id
		WHERE t.article_id IS NULL`
	var args []any
	if periodID != nil {
		query += " AND a.period_id = ?"
		args = append(args, *periodID)
	}
	query += " ORDER BY a.collected_at DESC"

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanArticles(rows)
}

// GetRelevantArticles returns articles triaged as relevant for a period.
func (db *DB) GetRelevantArticles(periodID string) ([]Article, error) {
	rows, err := db.conn.Query(
		`SELECT a.id, a.url, a.title, a.source, a.published_date, a.content,
		a.content_fetched, a.period_id, a.collected_at
		FROM articles a JOIN article_triage t ON a.id = t.article_id
		WHERE a.period_id = ? AND t.verdict = 'relevant'
		ORDER BY t.practical_score DESC`, periodID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanArticles(rows)
}

// GetArticleByID returns a single article by ID.
func (db *DB) GetArticleByID(articleID int64) (*Article, error) {
	row := db.conn.QueryRow(
		`SELECT id, url, title, source, published_date, content, content_fetched, period_id, collected_at
		FROM articles WHERE id = ?`, articleID,
	)
	a, err := scanArticle(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return a, nil
}

func scanArticles(rows *sql.Rows) ([]Article, error) {
	var articles []Article
	for rows.Next() {
		var a Article
		var fetched int
		if err := rows.Scan(&a.ID, &a.URL, &a.Title, &a.Source, &a.PublishedDate,
			&a.Content, &fetched, &a.PeriodID, &a.CollectedAt); err != nil {
			return nil, err
		}
		a.ContentFetched = fetched != 0
		articles = append(articles, a)
	}
	return articles, rows.Err()
}

func scanArticle(row *sql.Row) (*Article, error) {
	var a Article
	var fetched int
	if err := row.Scan(&a.ID, &a.URL, &a.Title, &a.Source, &a.PublishedDate,
		&a.Content, &fetched, &a.PeriodID, &a.CollectedAt); err != nil {
		return nil, err
	}
	a.ContentFetched = fetched != 0
	return &a, nil
}
