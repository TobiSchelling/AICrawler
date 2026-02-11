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

func TestTriageWithFeedbackSummary(t *testing.T) {
	db := openTestDB(t)
	// Create articles with sources
	a1, _ := db.InsertArticle("https://a.com", "Good Article", ptr("SwissTesting"), nil, ptr("Content A"), ptr("2026-02-06"))
	a2, _ := db.InsertArticle("https://b.com", "Bad Article", ptr("SpamBlog"), nil, ptr("Content B"), ptr("2026-02-06"))

	// Add triage for article types
	at := "experience_report"
	db.InsertTriage(a1, "relevant", &at, nil, nil, 4)
	at2 := "announcement"
	db.InsertTriage(a2, "relevant", &at2, nil, nil, 2)

	// Add feedback
	db.UpsertArticleFeedback(a1, "positive")
	db.UpsertArticleFeedback(a2, "negative")

	// Now create a new untriaged article to triage
	db.InsertArticle("https://c.com", "New AI Testing Article", ptr("SwissTesting"), nil, ptr("New content about testing"), ptr("2026-02-06"))

	resp, _ := json.Marshal(map[string]any{
		"verdict":          "relevant",
		"article_type":     "experience_report",
		"key_points":       []string{"Testing insight"},
		"relevance_reason": "Matches preferred source",
		"practical_score":  4,
	})

	// Capture the prompt sent to the LLM to verify feedback injection
	var capturedPrompt string
	provider := &mockProvider{response: string(resp)}
	captureProvider := &promptCapture{inner: provider}

	triager := NewTriager(db, captureProvider)
	result := triager.TriageArticles(context.Background(), "2026-02-06")

	if result.Processed != 1 {
		t.Errorf("expected 1 processed, got %d", result.Processed)
	}

	capturedPrompt = captureProvider.lastPrompt
	// Verify feedback patterns are in the prompt
	if !containsStr(capturedPrompt, "SwissTesting") {
		t.Error("expected 'SwissTesting' in triage prompt from feedback")
	}
	if !containsStr(capturedPrompt, "Preferred sources") {
		t.Error("expected 'Preferred sources' in triage prompt")
	}
}

// promptCapture wraps a provider and captures the last prompt.
type promptCapture struct {
	inner      *mockProvider
	lastPrompt string
}

func (p *promptCapture) Generate(ctx context.Context, prompt string, maxTokens int) (string, error) {
	p.lastPrompt = prompt
	return p.inner.Generate(ctx, prompt, maxTokens)
}

func (p *promptCapture) IsConfigured() bool { return true }

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
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
