package synthesize

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

func TestSynthesizeStoryline(t *testing.T) {
	db := openTestDB(t)
	a1, _ := db.InsertArticle("https://a.com", "AI Testing Part 1", nil, nil, ptr("Content 1"), ptr("2026-02-06"))
	a2, _ := db.InsertArticle("https://b.com", "AI Testing Part 2", nil, nil, ptr("Content 2"), ptr("2026-02-06"))
	db.InsertTriage(a1, "relevant", nil, []string{"Point 1"}, nil, 3)
	db.InsertTriage(a2, "relevant", nil, []string{"Point 2"}, nil, 3)
	sid, _ := db.InsertStoryline("2026-02-06", "AI Testing", []int64{a1, a2})

	resp, _ := json.Marshal(map[string]any{
		"title":     "AI Transforms Software Testing",
		"narrative": "Today saw significant progress...",
		"source_references": []map[string]string{
			{"title": "AI Testing Part 1", "url": "https://a.com", "contribution": "Foundation"},
			{"title": "AI Testing Part 2", "url": "https://b.com", "contribution": "Extensions"},
		},
	})

	synth := NewSynthesizer(db, &mockProvider{response: string(resp)})
	result := synth.SynthesizePeriod(context.Background(), "2026-02-06")

	if result.NarrativesCreated != 1 {
		t.Errorf("expected 1 narrative, got %d", result.NarrativesCreated)
	}

	narrative, _ := db.GetNarrativeForStoryline(sid)
	if narrative == nil || narrative.Title != "AI Transforms Software Testing" {
		t.Error("expected narrative title 'AI Transforms Software Testing'")
	}
}

func TestSynthesizeBrieflyNoted(t *testing.T) {
	db := openTestDB(t)
	a1, _ := db.InsertArticle("https://a.com", "Random Article", ptr("Source A"), nil, ptr("Content"), ptr("2026-02-06"))
	db.InsertTriage(a1, "relevant", nil, []string{"A key point"}, nil, 3)
	sid, _ := db.InsertStoryline("2026-02-06", brieflyNotedLabel, []int64{a1})

	mock := &mockProvider{} // Should NOT be called for briefly noted
	synth := NewSynthesizer(db, mock)
	result := synth.SynthesizePeriod(context.Background(), "2026-02-06")

	if result.NarrativesCreated != 1 {
		t.Errorf("expected 1 narrative, got %d", result.NarrativesCreated)
	}

	narrative, _ := db.GetNarrativeForStoryline(sid)
	if narrative == nil {
		t.Fatal("expected narrative")
	}
	if narrative.Title != brieflyNotedLabel {
		t.Errorf("expected title %q, got %q", brieflyNotedLabel, narrative.Title)
	}
	if !strings.Contains(narrative.NarrativeText, "Random Article") {
		t.Error("expected narrative to contain article title")
	}
}

func TestSynthesizeSkipsExisting(t *testing.T) {
	db := openTestDB(t)
	a1, _ := db.InsertArticle("https://a.com", "A", nil, nil, ptr("C"), ptr("2026-02-06"))
	sid, _ := db.InsertStoryline("2026-02-06", "Test", []int64{a1})
	db.InsertStorylineNarrative(sid, "2026-02-06", "Existing", "Already done", nil)

	mock := &mockProvider{}
	synth := NewSynthesizer(db, mock)
	result := synth.SynthesizePeriod(context.Background(), "2026-02-06")

	if result.NarrativesCreated != 1 {
		t.Errorf("expected 1 (existing counted), got %d", result.NarrativesCreated)
	}
}
