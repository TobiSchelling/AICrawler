package database

import (
	"database/sql"
	"fmt"
	"log"
)

// getSchemaVersion reads PRAGMA user_version from the database.
func getSchemaVersion(conn *sql.DB) (int, error) {
	var version int
	if err := conn.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		return 0, fmt.Errorf("reading schema version: %w", err)
	}
	return version, nil
}

// isLegacyDB returns true if the database has tables but no user_version set.
// This detects databases created before the migration system existed.
func isLegacyDB(conn *sql.DB) (bool, error) {
	var count int
	err := conn.QueryRow(
		"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='articles'",
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("checking for legacy tables: %w", err)
	}
	return count > 0, nil
}

// migrate brings the database schema up to the latest version.
// It uses PRAGMA user_version to track which migrations have been applied.
func migrate(conn *sql.DB) error {
	current, err := getSchemaVersion(conn)
	if err != nil {
		return err
	}

	// Legacy DB detection: tables exist but user_version is 0.
	// Stamp as version 1 since the schema already matches migration 1.
	if current == 0 {
		legacy, err := isLegacyDB(conn)
		if err != nil {
			return err
		}
		if legacy {
			log.Printf("detected legacy database, stamping as version 1")
			if _, err := conn.Exec("PRAGMA user_version = 1"); err != nil {
				return fmt.Errorf("stamping legacy version: %w", err)
			}
			current = 1
		}
	}

	latest := latestVersion()
	if current >= latest {
		return nil
	}

	for _, m := range migrations {
		if m.Version <= current {
			continue
		}

		log.Printf("applying migration %d: %s", m.Version, m.Description)

		tx, err := conn.Begin()
		if err != nil {
			return fmt.Errorf("begin migration %d: %w", m.Version, err)
		}

		if err := m.Up(tx); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration %d (%s): %w", m.Version, m.Description, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %d: %w", m.Version, err)
		}

		// Set user_version outside the transaction (modernc/sqlite requirement).
		// Safe: if we crash here, the idempotent DDL lets the migration re-run.
		if _, err := conn.Exec(fmt.Sprintf("PRAGMA user_version = %d", m.Version)); err != nil {
			return fmt.Errorf("setting version %d: %w", m.Version, err)
		}
	}

	return nil
}
