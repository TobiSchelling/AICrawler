package triage

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/TobiSchelling/AICrawler/internal/database"
)

// mockProvider implements llm.Provider for testing.
type mockProvider struct {
	response string
	err      error
}

func (m *mockProvider) Generate(_ context.Context, _ string, _ int) (string, error) {
	return m.response, m.err
}

func (m *mockProvider) IsConfigured() bool { return true }

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

func TestTriageRelevantArticle(t *testing.T) {
	db := openTestDB(t)
	aid, _ := db.InsertArticle("https://example.com/test", "How We Use Claude for Code Review",
		ptr("Blog"), nil, ptr("A detailed experience report..."), ptr("2026-02-06"))

	resp, _ := json.Marshal(map[string]any{
		"verdict":          "relevant",
		"article_type":     "experience_report",
		"key_points":       []string{"AI code review improves quality", "Reduced review time by 40%"},
		"relevance_reason": "Direct experience report on AI in development",
		"practical_score":  4,
	})

	triager := NewTriager(db, &mockProvider{response: string(resp)})
	result := triager.TriageArticles(context.Background(), "2026-02-06")

	if result.Processed != 1 {
		t.Errorf("expected 1 processed, got %d", result.Processed)
	}
	if result.Relevant != 1 {
		t.Errorf("expected 1 relevant, got %d", result.Relevant)
	}

	triage, _ := db.GetTriage(aid)
	if triage == nil || triage.Verdict != "relevant" {
		t.Error("expected relevant verdict")
	}
	if triage.PracticalScore != 4 {
		t.Errorf("expected score 4, got %d", triage.PracticalScore)
	}
}

func TestTriageSkipArticle(t *testing.T) {
	db := openTestDB(t)
	db.InsertArticle("https://example.com/funding", "AI Startup Raises $500M",
		nil, nil, ptr("Funding announcement..."), ptr("2026-02-06"))

	resp, _ := json.Marshal(map[string]any{
		"verdict":          "skip",
		"article_type":     "announcement",
		"key_points":       []string{},
		"relevance_reason": "Pure funding announcement",
		"practical_score":  0,
	})

	triager := NewTriager(db, &mockProvider{response: string(resp)})
	result := triager.TriageArticles(context.Background(), "2026-02-06")

	if result.Processed != 1 || result.Skipped != 1 {
		t.Errorf("expected 1 processed/skipped, got %d/%d", result.Processed, result.Skipped)
	}
}

func TestTriageHandlesUnparseableResponse(t *testing.T) {
	db := openTestDB(t)
	db.InsertArticle("https://example.com/test", "Test Article",
		nil, nil, ptr("Some content"), ptr("2026-02-06"))

	triager := NewTriager(db, &mockProvider{response: "This is not JSON at all"})
	result := triager.TriageArticles(context.Background(), "2026-02-06")

	if result.Processed != 1 {
		t.Errorf("expected 1 processed, got %d", result.Processed)
	}
	if result.Relevant != 1 {
		t.Errorf("expected 1 relevant (default), got %d", result.Relevant)
	}
}

func TestTriageSkipsAlreadyTriaged(t *testing.T) {
	db := openTestDB(t)
	aid, _ := db.InsertArticle("https://example.com/test", "Already Triaged",
		nil, nil, ptr("Content"), ptr("2026-02-06"))
	db.InsertTriage(aid, "relevant", nil, nil, nil, 3)

	mock := &mockProvider{}
	triager := NewTriager(db, mock)
	result := triager.TriageArticles(context.Background(), "2026-02-06")

	if result.Processed != 0 {
		t.Errorf("expected 0 processed, got %d", result.Processed)
	}
}

func TestTriageNoProvider(t *testing.T) {
	db := openTestDB(t)
	db.InsertArticle("https://example.com/test", "Test",
		nil, nil, ptr("C"), ptr("2026-02-06"))

	triager := NewTriager(db, nil)
	result := triager.TriageArticles(context.Background(), "2026-02-06")

	if result.Errors != 1 {
		t.Errorf("expected 1 error, got %d", result.Errors)
	}
}
