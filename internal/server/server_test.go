package server

import (
	"fmt"
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

func TestStorylineFeedbackRoute(t *testing.T) {
	db := openTestDB(t)
	a1, _ := db.InsertArticle("https://a.com", "A", nil, nil, nil, ptr("2026-02-06"))
	sid, _ := db.InsertStoryline("2026-02-06", "AI Testing", []int64{a1})
	db.InsertStorylineNarrative(sid, "2026-02-06", "AI Testing", "Narrative text.", nil)
	db.InsertBriefing("2026-02-06", "TL;DR", "Body", 1, 1)

	srv, err := New(db)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// POST feedback
	body := strings.NewReader("period_id=2026-02-06")
	req := httptest.NewRequest("POST", fmt.Sprintf("/feedback/storyline/%d/useful", sid), body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", rec.Code)
	}
	loc := rec.Header().Get("Location")
	if !strings.Contains(loc, "#storyline-") {
		t.Errorf("expected anchor in redirect, got %q", loc)
	}

	// Verify feedback stored
	fb, _ := db.GetStorylineFeedback(sid)
	if fb == nil || fb.Rating != "useful" {
		t.Error("expected 'useful' feedback stored")
	}

	// Toggle off: POST same rating again
	body = strings.NewReader("period_id=2026-02-06")
	req = httptest.NewRequest("POST", fmt.Sprintf("/feedback/storyline/%d/useful", sid), body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	fb, _ = db.GetStorylineFeedback(sid)
	if fb != nil {
		t.Error("expected nil feedback after toggle off")
	}
}

func TestArticleFeedbackRoute(t *testing.T) {
	db := openTestDB(t)
	aid, _ := db.InsertArticle("https://a.com", "A", nil, nil, nil, ptr("2026-02-06"))
	sid, _ := db.InsertStoryline("2026-02-06", "Test", []int64{aid})
	db.InsertStorylineNarrative(sid, "2026-02-06", "Test", "Narrative.", nil)
	db.InsertBriefing("2026-02-06", "TL;DR", "Body", 1, 1)

	srv, err := New(db)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	body := strings.NewReader(fmt.Sprintf("period_id=2026-02-06&storyline_id=%d", sid))
	req := httptest.NewRequest("POST", fmt.Sprintf("/feedback/article/%d/positive", aid), body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", rec.Code)
	}

	fb, _ := db.GetArticleFeedback(aid)
	if fb == nil || fb.Rating != "positive" {
		t.Error("expected 'positive' feedback stored")
	}
}

func TestBriefingStructured(t *testing.T) {
	db := openTestDB(t)
	a1, _ := db.InsertArticle("https://a.com", "Article One", ptr("TestSource"), nil, nil, ptr("2026-02-06"))
	sid, _ := db.InsertStoryline("2026-02-06", "AI Testing", []int64{a1})
	db.InsertStorylineNarrative(sid, "2026-02-06", "AI Testing Tools", "A narrative about AI testing.", nil)
	at := "experience_report"
	db.InsertTriage(a1, "relevant", &at, nil, nil, 4)
	db.InsertBriefing("2026-02-06", "- Key point", "## Section\nContent", 1, 1)

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

	// Should contain storyline title
	if !strings.Contains(body, "AI Testing Tools") {
		t.Error("expected storyline title 'AI Testing Tools' in response")
	}
	// Should contain article title
	if !strings.Contains(body, "Article One") {
		t.Error("expected article title 'Article One' in response")
	}
	// Should contain feedback form action
	if !strings.Contains(body, "/feedback/storyline/") {
		t.Error("expected storyline feedback form in response")
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
