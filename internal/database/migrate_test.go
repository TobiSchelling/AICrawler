package database

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func TestMigrateNewDB(t *testing.T) {
	db := openTestDB(t)

	version, err := getSchemaVersion(db.conn)
	if err != nil {
		t.Fatalf("getSchemaVersion: %v", err)
	}
	if version != latestVersion() {
		t.Errorf("expected version %d, got %d", latestVersion(), version)
	}
}

func TestMigrateLegacyDB(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "legacy.db")

	// Simulate a pre-migration database: create tables without setting user_version.
	raw, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open raw db: %v", err)
	}
	_, err = raw.Exec(`CREATE TABLE articles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		url TEXT UNIQUE NOT NULL,
		title TEXT NOT NULL
	)`)
	if err != nil {
		t.Fatalf("create legacy table: %v", err)
	}
	raw.Close()

	// Now open via the migration system.
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	version, err := getSchemaVersion(db.conn)
	if err != nil {
		t.Fatalf("getSchemaVersion: %v", err)
	}
	if version != latestVersion() {
		t.Errorf("expected version %d after legacy migration, got %d", latestVersion(), version)
	}
}

func TestMigrateIdempotent(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "idem.db")

	db1, err := Open(dbPath)
	if err != nil {
		t.Fatalf("first Open: %v", err)
	}
	db1.Close()

	db2, err := Open(dbPath)
	if err != nil {
		t.Fatalf("second Open: %v", err)
	}
	defer db2.Close()

	version, err := getSchemaVersion(db2.conn)
	if err != nil {
		t.Fatalf("getSchemaVersion: %v", err)
	}
	if version != latestVersion() {
		t.Errorf("expected version %d, got %d", latestVersion(), version)
	}
}

func TestGetSchemaVersionNewDB(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "empty.db")
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer conn.Close()

	version, err := getSchemaVersion(conn)
	if err != nil {
		t.Fatalf("getSchemaVersion: %v", err)
	}
	if version != 0 {
		t.Errorf("expected version 0 on new db, got %d", version)
	}
}

func TestIsLegacyDBFalseOnNew(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "fresh.db")
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer conn.Close()

	legacy, err := isLegacyDB(conn)
	if err != nil {
		t.Fatalf("isLegacyDB: %v", err)
	}
	if legacy {
		t.Error("expected isLegacyDB=false on empty database")
	}
}
