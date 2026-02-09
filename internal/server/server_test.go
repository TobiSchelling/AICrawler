package server

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/TobiSchelling/AICrawler/internal/database"
)

func openTestDB(t *testing.T) *database.DB {
	t.Helper()
	db, err := database.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func ptr(s string) *string { return &s }

func TestIndexRoute(t *testing.T) {
	db := openTestDB(t)
	srv, err := New(db)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Briefings") {
		t.Error("expected 'Briefings' in response body")
	}
}

func TestBriefingRoute(t *testing.T) {
	db := openTestDB(t)
	db.InsertBriefing("2026-02-06", "- Key point", "## Section\nContent", 1, 5)

	srv, err := New(db)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	req := httptest.NewRequest("GET", "/briefing/2026-02-06", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "AI Briefing") {
		t.Error("expected 'AI Briefing' in response")
	}
}

func TestStaticRoute(t *testing.T) {
	db := openTestDB(t)
	srv, err := New(db)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	req := httptest.NewRequest("GET", "/static/style.css", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "font-sans") {
		t.Error("expected CSS content")
	}
}
