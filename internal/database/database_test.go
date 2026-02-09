package database

import (
	"path/filepath"
	"testing"
)

func openTestDB(t *testing.T) *DB {
	t.Helper()
	db, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func ptr(s string) *string { return &s }

func TestInsertArticle(t *testing.T) {
	db := openTestDB(t)
	id, err := db.InsertArticle("https://example.com/test", "Test Article", ptr("Test Source"), ptr("2026-01-27"), ptr("Test content here"), ptr("2026-02-06"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id == 0 {
		t.Error("expected non-zero article ID")
	}
}

func TestInsertDuplicateArticle(t *testing.T) {
	db := openTestDB(t)
	_, _ = db.InsertArticle("https://example.com/dup", "First", nil, nil, nil, ptr("2026-02-06"))
	id, err := db.InsertArticle("https://example.com/dup", "Duplicate", nil, nil, nil, ptr("2026-02-06"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 0 {
		t.Error("expected 0 for duplicate article")
	}
}

func TestGetArticlesForPeriod(t *testing.T) {
	db := openTestDB(t)
	db.InsertArticle("https://a.com", "A", nil, nil, nil, ptr("2026-02-06"))
	db.InsertArticle("https://b.com", "B", nil, nil, nil, ptr("2026-02-06"))
	db.InsertArticle("https://c.com", "C", nil, nil, nil, ptr("2026-02-05"))

	articles, err := db.GetArticlesForPeriod("2026-02-06")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(articles) != 2 {
		t.Errorf("expected 2 articles, got %d", len(articles))
	}
}

func TestArticlesNeedingFetch(t *testing.T) {
	db := openTestDB(t)
	db.InsertArticle("https://a.com", "No content", nil, nil, nil, ptr("2026-02-06"))
	db.InsertArticle("https://b.com", "Has content", nil, nil, ptr("Some text"), ptr("2026-02-06"))

	period := "2026-02-06"
	needing, err := db.GetArticlesNeedingFetch(&period)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(needing) != 1 {
		t.Errorf("expected 1 article needing fetch, got %d", len(needing))
	}
	if needing[0].Title != "No content" {
		t.Errorf("expected 'No content', got %q", needing[0].Title)
	}
}

func TestUpdateArticleContent(t *testing.T) {
	db := openTestDB(t)
	id, _ := db.InsertArticle("https://a.com", "Test", nil, nil, nil, ptr("2026-02-06"))
	content := "Fetched content"
	if err := db.UpdateArticleContent(id, &content); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	article, err := db.GetArticleByID(id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if article.Content == nil || *article.Content != "Fetched content" {
		t.Error("expected content to be updated")
	}
	if !article.ContentFetched {
		t.Error("expected content_fetched to be true")
	}
}

func TestTriageLifecycle(t *testing.T) {
	db := openTestDB(t)
	id, _ := db.InsertArticle("https://a.com", "Test", nil, nil, nil, ptr("2026-02-06"))

	period := "2026-02-06"
	untriaged, _ := db.GetUntriagedArticles(&period)
	if len(untriaged) != 1 {
		t.Fatalf("expected 1 untriaged, got %d", len(untriaged))
	}

	at := "experience_report"
	reason := "Practical AI content"
	if err := db.InsertTriage(id, "relevant", &at, []string{"Point 1", "Point 2"}, &reason, 4); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	untriaged, _ = db.GetUntriagedArticles(&period)
	if len(untriaged) != 0 {
		t.Error("expected 0 untriaged after triage")
	}

	relevant, _ := db.GetRelevantArticles("2026-02-06")
	if len(relevant) != 1 {
		t.Error("expected 1 relevant article")
	}

	triage, _ := db.GetTriage(id)
	if triage == nil {
		t.Fatal("expected triage result")
	}
	if triage.Verdict != "relevant" {
		t.Errorf("expected verdict 'relevant', got %q", triage.Verdict)
	}
	if len(triage.KeyPoints) != 2 {
		t.Errorf("expected 2 key points, got %d", len(triage.KeyPoints))
	}
	if triage.PracticalScore != 4 {
		t.Errorf("expected practical_score 4, got %d", triage.PracticalScore)
	}
}

func TestTriageStats(t *testing.T) {
	db := openTestDB(t)
	a1, _ := db.InsertArticle("https://a.com", "A", nil, nil, nil, ptr("2026-02-06"))
	a2, _ := db.InsertArticle("https://b.com", "B", nil, nil, nil, ptr("2026-02-06"))

	db.InsertTriage(a1, "relevant", nil, nil, nil, 3)
	db.InsertTriage(a2, "skip", nil, nil, nil, 0)

	stats, err := db.GetTriageStats("2026-02-06")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.Total != 2 {
		t.Errorf("expected total 2, got %d", stats.Total)
	}
	if stats.Relevant != 1 {
		t.Errorf("expected relevant 1, got %d", stats.Relevant)
	}
	if stats.Skipped != 1 {
		t.Errorf("expected skipped 1, got %d", stats.Skipped)
	}
}

func TestStorylineLifecycle(t *testing.T) {
	db := openTestDB(t)
	a1, _ := db.InsertArticle("https://a.com", "A", nil, nil, nil, ptr("2026-02-06"))
	a2, _ := db.InsertArticle("https://b.com", "B", nil, nil, nil, ptr("2026-02-06"))

	sid, err := db.InsertStoryline("2026-02-06", "AI Testing", []int64{a1, a2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sid == 0 {
		t.Error("expected non-zero storyline ID")
	}

	storylines, _ := db.GetStorylinesForPeriod("2026-02-06")
	if len(storylines) != 1 {
		t.Fatalf("expected 1 storyline, got %d", len(storylines))
	}
	if storylines[0].Label != "AI Testing" {
		t.Errorf("expected label 'AI Testing', got %q", storylines[0].Label)
	}
	if storylines[0].ArticleCount != 2 {
		t.Errorf("expected article_count 2, got %d", storylines[0].ArticleCount)
	}

	articles, _ := db.GetStorylineArticles(sid)
	if len(articles) != 2 {
		t.Errorf("expected 2 storyline articles, got %d", len(articles))
	}
}

func TestClearStorylines(t *testing.T) {
	db := openTestDB(t)
	a1, _ := db.InsertArticle("https://a.com", "A", nil, nil, nil, ptr("2026-02-06"))
	sid, _ := db.InsertStoryline("2026-02-06", "Test", []int64{a1})
	db.InsertStorylineNarrative(sid, "2026-02-06", "T", "N", nil)

	if err := db.ClearStorylinesForPeriod("2026-02-06"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	storylines, _ := db.GetStorylinesForPeriod("2026-02-06")
	if len(storylines) != 0 {
		t.Errorf("expected 0 storylines after clear, got %d", len(storylines))
	}
	narratives, _ := db.GetNarrativesForPeriod("2026-02-06")
	if len(narratives) != 0 {
		t.Errorf("expected 0 narratives after clear, got %d", len(narratives))
	}
}

func TestBriefingLifecycle(t *testing.T) {
	db := openTestDB(t)
	_, err := db.InsertBriefing("2026-02-06", "- Key point 1\n- Key point 2", "## Section\nNarrative here.", 3, 15)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	briefing, _ := db.GetBriefing("2026-02-06")
	if briefing == nil {
		t.Fatal("expected briefing")
	}
	if briefing.StorylineCount != 3 {
		t.Errorf("expected 3 storylines, got %d", briefing.StorylineCount)
	}
	if briefing.ArticleCount != 15 {
		t.Errorf("expected 15 articles, got %d", briefing.ArticleCount)
	}

	all, _ := db.GetAllBriefings()
	if len(all) != 1 {
		t.Errorf("expected 1 briefing, got %d", len(all))
	}
}

func TestPriorityLifecycle(t *testing.T) {
	db := openTestDB(t)
	pid, err := db.InsertPriority("AI Agents", "Agent frameworks", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pid == 0 {
		t.Error("expected non-zero priority ID")
	}

	priority, _ := db.GetPriority(pid)
	if priority == nil {
		t.Fatal("expected priority")
	}
	if priority.Title != "AI Agents" {
		t.Errorf("expected title 'AI Agents', got %q", priority.Title)
	}
	if !priority.IsActive {
		t.Error("expected priority to be active")
	}

	db.TogglePriority(pid)
	priority, _ = db.GetPriority(pid)
	if priority.IsActive {
		t.Error("expected priority to be inactive after toggle")
	}

	newTitle := "AI Agent Frameworks"
	db.UpdatePriority(pid, &newTitle, nil, nil)
	priority, _ = db.GetPriority(pid)
	if priority.Title != "AI Agent Frameworks" {
		t.Errorf("expected updated title, got %q", priority.Title)
	}

	db.DeletePriority(pid)
	priority, _ = db.GetPriority(pid)
	if priority != nil {
		t.Error("expected nil after delete")
	}
}

func TestGetStats(t *testing.T) {
	db := openTestDB(t)
	stats, err := db.GetStats()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.TotalArticles != 0 {
		t.Errorf("expected 0 articles, got %d", stats.TotalArticles)
	}

	db.InsertArticle("https://a.com", "A", nil, nil, nil, ptr("2026-02-06"))
	db.InsertPriority("Test Priority", "", nil)

	stats, _ = db.GetStats()
	if stats.TotalArticles != 1 {
		t.Errorf("expected 1 article, got %d", stats.TotalArticles)
	}
	if stats.TotalPriorities != 1 {
		t.Errorf("expected 1 priority, got %d", stats.TotalPriorities)
	}
}

func TestGetToday(t *testing.T) {
	today := GetToday()
	if len(today) != 10 {
		t.Errorf("expected 10-char date, got %q", today)
	}
	if today[4] != '-' || today[7] != '-' {
		t.Errorf("expected YYYY-MM-DD format, got %q", today)
	}
}

func TestFormatPeriodDisplaySingleDay(t *testing.T) {
	result := FormatPeriodDisplay("2026-02-06")
	if result == "" || result == "2026-02-06" {
		t.Errorf("expected formatted date, got %q", result)
	}
	// Should contain "Feb" and "2026"
	if !contains(result, "Feb") || !contains(result, "2026") {
		t.Errorf("expected 'Feb' and '2026' in %q", result)
	}
}

func TestFormatPeriodDisplayRange(t *testing.T) {
	result := FormatPeriodDisplay("2026-02-01..2026-02-06")
	if !contains(result, "Feb 01") {
		t.Errorf("expected 'Feb 01' in %q", result)
	}
	if !contains(result, "Feb 06") {
		t.Errorf("expected 'Feb 06' in %q", result)
	}
	if !contains(result, "-") {
		t.Errorf("expected '-' separator in %q", result)
	}
}

func TestMakePeriodIDSingleDay(t *testing.T) {
	result := MakePeriodID("2026-02-06", "2026-02-06")
	if result != "2026-02-06" {
		t.Errorf("expected '2026-02-06', got %q", result)
	}
}

func TestMakePeriodIDRange(t *testing.T) {
	result := MakePeriodID("2026-02-01", "2026-02-06")
	if result != "2026-02-01..2026-02-06" {
		t.Errorf("expected '2026-02-01..2026-02-06', got %q", result)
	}
}

func TestGetLastRunDate(t *testing.T) {
	db := openTestDB(t)

	last, err := db.GetLastRunDate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if last != "" {
		t.Errorf("expected empty string, got %q", last)
	}

	db.InsertReport("2026-02-05", 10, 3)
	last, _ = db.GetLastRunDate()
	if last != "2026-02-05" {
		t.Errorf("expected '2026-02-05', got %q", last)
	}
}

func TestGetLastRunDateRange(t *testing.T) {
	db := openTestDB(t)
	db.InsertReport("2026-02-01..2026-02-05", 10, 3)

	last, _ := db.GetLastRunDate()
	if last != "2026-02-05" {
		t.Errorf("expected '2026-02-05', got %q", last)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
