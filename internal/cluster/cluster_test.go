package cluster

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/TobiSchelling/AICrawler/internal/database"
)

// mockEmbedder implements llm.Embedder for testing.
type mockEmbedder struct {
	embeddings [][]float64
}

func (m *mockEmbedder) Embed(_ context.Context, _ []string) ([][]float64, error) {
	return m.embeddings, nil
}

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

func TestClusterNoArticles(t *testing.T) {
	db := openTestDB(t)
	clusterer := NewClusterer(db, nil, DefaultDistanceThreshold)
	result, err := clusterer.ClusterArticles(context.Background(), "2026-02-06")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.StorylineCount != 0 || result.ArticleCount != 0 {
		t.Errorf("expected 0/0, got %d/%d", result.StorylineCount, result.ArticleCount)
	}
}

func TestClusterSingleArticleGoesToBrieflyNoted(t *testing.T) {
	db := openTestDB(t)
	aid, _ := db.InsertArticle("https://a.com", "Solo Article", nil, nil, ptr("Content"), ptr("2026-02-06"))
	db.InsertTriage(aid, "relevant", nil, nil, nil, 3)

	clusterer := NewClusterer(db, nil, DefaultDistanceThreshold)
	result, err := clusterer.ClusterArticles(context.Background(), "2026-02-06")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.StorylineCount != 1 {
		t.Errorf("expected 1 storyline, got %d", result.StorylineCount)
	}
	if result.BrieflyNotedCount != 1 {
		t.Errorf("expected 1 briefly noted, got %d", result.BrieflyNotedCount)
	}

	storylines, _ := db.GetStorylinesForPeriod("2026-02-06")
	if len(storylines) == 0 || storylines[0].Label != BrieflyNotedLabel {
		t.Error("expected Briefly Noted label")
	}
}

func TestClusterSimilarArticlesGrouped(t *testing.T) {
	db := openTestDB(t)
	for i := 0; i < 3; i++ {
		aid, _ := db.InsertArticle(
			"https://example.com/ai-testing-"+string(rune('0'+i)),
			"AI-Powered Testing Framework: Revolution in QA",
			nil, nil, ptr("How AI is transforming testing"), ptr("2026-02-06"))
		db.InsertTriage(aid, "relevant", nil, nil, nil, 4)
	}
	aid, _ := db.InsertArticle("https://example.com/crypto", "New Cryptocurrency Market Analysis",
		nil, nil, ptr("Analysis of cryptocurrency markets"), ptr("2026-02-06"))
	db.InsertTriage(aid, "relevant", nil, nil, nil, 2)

	embeddings := [][]float64{
		{1.0, 0.0, 0.0},
		{0.95, 0.05, 0.0},
		{0.9, 0.1, 0.0},
		{0.0, 0.0, 1.0},
	}

	clusterer := NewClusterer(db, &mockEmbedder{embeddings: embeddings}, 1.0)
	result, err := clusterer.ClusterArticles(context.Background(), "2026-02-06")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ArticleCount != 4 {
		t.Errorf("expected 4 articles, got %d", result.ArticleCount)
	}
	if result.StorylineCount < 1 {
		t.Errorf("expected at least 1 storyline, got %d", result.StorylineCount)
	}
}

func TestReClusteringClearsOldData(t *testing.T) {
	db := openTestDB(t)
	aid, _ := db.InsertArticle("https://a.com", "A", nil, nil, ptr("Content"), ptr("2026-02-06"))
	db.InsertTriage(aid, "relevant", nil, nil, nil, 3)

	clusterer := NewClusterer(db, nil, DefaultDistanceThreshold)
	clusterer.ClusterArticles(context.Background(), "2026-02-06")

	storylines, _ := db.GetStorylinesForPeriod("2026-02-06")
	if len(storylines) != 1 {
		t.Fatalf("expected 1 storyline, got %d", len(storylines))
	}

	// Re-cluster
	clusterer.ClusterArticles(context.Background(), "2026-02-06")

	storylines, _ = db.GetStorylinesForPeriod("2026-02-06")
	if len(storylines) != 1 {
		t.Errorf("expected 1 storyline after re-cluster, got %d", len(storylines))
	}
}
