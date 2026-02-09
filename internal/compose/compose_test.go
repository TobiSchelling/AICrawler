package compose

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/TobiSchelling/AICrawler/internal/database"
)

type mockProvider struct {
	response string
}

func (m *mockProvider) Generate(_ context.Context, _ string, _ int) (string, error) {
	return m.response, nil
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

func TestComposeBriefing(t *testing.T) {
	db := openTestDB(t)
	a1, _ := db.InsertArticle("https://a.com", "A", nil, nil, ptr("C"), ptr("2026-02-06"))
	a2, _ := db.InsertArticle("https://b.com", "B", nil, nil, ptr("C"), ptr("2026-02-06"))
	sid, _ := db.InsertStoryline("2026-02-06", "AI Testing", []int64{a1, a2})
	db.InsertStorylineNarrative(sid, "2026-02-06", "AI Transforms Testing",
		"Today saw major changes in testing...",
		[]database.SourceReference{{Title: "A", URL: "https://a.com"}})

	resp, _ := json.Marshal(map[string]any{
		"tldr_bullets": []string{
			"AI testing tools gained significant traction",
			"New frameworks emerged for LLM-based QA",
		},
	})

	composer := NewComposer(db, &mockProvider{response: string(resp)})
	briefing, err := composer.ComposeBriefing(context.Background(), "2026-02-06")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if briefing == nil {
		t.Fatal("expected briefing")
	}
	if briefing.PeriodID != "2026-02-06" {
		t.Errorf("expected period_id '2026-02-06', got %q", briefing.PeriodID)
	}
	if briefing.StorylineCount != 1 {
		t.Errorf("expected 1 storyline, got %d", briefing.StorylineCount)
	}
	if briefing.ArticleCount != 2 {
		t.Errorf("expected 2 articles, got %d", briefing.ArticleCount)
	}
	if !strings.Contains(briefing.TLDR, "AI testing tools") {
		t.Error("expected TL;DR to contain 'AI testing tools'")
	}
	if !strings.Contains(briefing.BodyMarkdown, "AI Transforms Testing") {
		t.Error("expected body to contain 'AI Transforms Testing'")
	}
}

func TestComposeEmptyPeriod(t *testing.T) {
	db := openTestDB(t)
	composer := NewComposer(db, &mockProvider{})
	briefing, err := composer.ComposeBriefing(context.Background(), "2026-02-06")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if briefing == nil {
		t.Fatal("expected briefing")
	}
	if briefing.ArticleCount != 0 {
		t.Errorf("expected 0 articles, got %d", briefing.ArticleCount)
	}
}

func TestComposeFallbackWithoutProvider(t *testing.T) {
	db := openTestDB(t)
	a1, _ := db.InsertArticle("https://a.com", "A", nil, nil, ptr("C"), ptr("2026-02-06"))
	sid, _ := db.InsertStoryline("2026-02-06", "AI Testing", []int64{a1})
	db.InsertStorylineNarrative(sid, "2026-02-06", "AI Testing Narrative", "Content here.", nil)

	// Provider returns empty (simulates unavailable)
	composer := NewComposer(db, &mockProvider{response: ""})
	briefing, err := composer.ComposeBriefing(context.Background(), "2026-02-06")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if briefing == nil {
		t.Fatal("expected briefing")
	}
	if !strings.Contains(briefing.TLDR, "AI Testing Narrative") {
		t.Errorf("expected fallback TL;DR with storyline title, got %q", briefing.TLDR)
	}
}
