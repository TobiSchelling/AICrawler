package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

// InsertPriority creates a new research priority.
func (db *DB) InsertPriority(title, description string, keywords []string) (int64, error) {
	var kwJSON *string
	if keywords != nil {
		data, err := json.Marshal(keywords)
		if err != nil {
			return 0, err
		}
		s := string(data)
		kwJSON = &s
	}

	result, err := db.conn.Exec(
		`INSERT INTO research_priorities (title, description, keywords) VALUES (?, ?, ?)`,
		title, description, kwJSON,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetAllPriorities returns all research priorities.
func (db *DB) GetAllPriorities() ([]ResearchPriority, error) {
	return db.queryPriorities("SELECT * FROM research_priorities ORDER BY created_at DESC")
}

// GetActivePriorities returns only active research priorities.
func (db *DB) GetActivePriorities() ([]ResearchPriority, error) {
	return db.queryPriorities("SELECT * FROM research_priorities WHERE is_active = 1 ORDER BY created_at DESC")
}

// GetPriority returns a single priority by ID.
func (db *DB) GetPriority(priorityID int64) (*ResearchPriority, error) {
	row := db.conn.QueryRow(
		"SELECT id, title, description, keywords, is_active, created_at, updated_at FROM research_priorities WHERE id = ?",
		priorityID,
	)
	p, err := scanPriority(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return p, nil
}

// UpdatePriority updates specified fields of a priority.
func (db *DB) UpdatePriority(priorityID int64, title, description *string, keywords []string) error {
	var updates []string
	var args []any

	if title != nil {
		updates = append(updates, "title = ?")
		args = append(args, *title)
	}
	if description != nil {
		updates = append(updates, "description = ?")
		args = append(args, *description)
	}
	if keywords != nil {
		data, err := json.Marshal(keywords)
		if err != nil {
			return err
		}
		updates = append(updates, "keywords = ?")
		args = append(args, string(data))
	}
	if len(updates) == 0 {
		return nil
	}

	updates = append(updates, "updated_at = datetime('now')")
	args = append(args, priorityID)

	query := fmt.Sprintf("UPDATE research_priorities SET %s WHERE id = ?", strings.Join(updates, ", "))
	_, err := db.conn.Exec(query, args...)
	return err
}

// TogglePriority toggles the active state of a priority.
func (db *DB) TogglePriority(priorityID int64) error {
	_, err := db.conn.Exec(
		`UPDATE research_priorities SET is_active = NOT is_active, updated_at = datetime('now') WHERE id = ?`,
		priorityID,
	)
	return err
}

// DeletePriority removes a priority.
func (db *DB) DeletePriority(priorityID int64) error {
	_, err := db.conn.Exec("DELETE FROM research_priorities WHERE id = ?", priorityID)
	return err
}

func (db *DB) queryPriorities(query string, args ...any) ([]ResearchPriority, error) {
	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var priorities []ResearchPriority
	for rows.Next() {
		var p ResearchPriority
		var kwJSON, desc *string
		var active int
		if err := rows.Scan(&p.ID, &p.Title, &desc, &kwJSON, &active, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		p.Description = desc
		p.IsActive = active != 0
		if kwJSON != nil {
			if err := json.Unmarshal([]byte(*kwJSON), &p.Keywords); err != nil {
				p.Keywords = nil
			}
		}
		priorities = append(priorities, p)
	}
	return priorities, rows.Err()
}

func scanPriority(row *sql.Row) (*ResearchPriority, error) {
	var p ResearchPriority
	var kwJSON, desc *string
	var active int
	if err := row.Scan(&p.ID, &p.Title, &desc, &kwJSON, &active, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil, err
	}
	p.Description = desc
	p.IsActive = active != 0
	if kwJSON != nil {
		if err := json.Unmarshal([]byte(*kwJSON), &p.Keywords); err != nil {
			p.Keywords = nil
		}
	}
	return &p, nil
}
