package triage

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/TobiSchelling/AICrawler/internal/database"
	"github.com/TobiSchelling/AICrawler/internal/llm"
)

const triagePrompt = `You are triaging AI news articles for a daily briefing aimed at people who build software.

Decide whether this article is RELEVANT or should be SKIPPED.

RELEVANT means: practical AI developments, experience reports from using AI tools, new techniques you can try, architecture patterns, tool releases, significant model updates, or insightful commentary on AI's impact on software development.

SKIP means: pure academic research papers, funding/investment announcements, marketing fluff, product launches with no technical substance, celebrity AI opinions, or AI doom/hype pieces with no practical content.

Research priorities to give extra weight:
%s

Article Title: %s
Source: %s
Content:
%s

Respond with ONLY this JSON:
{
    "verdict": "relevant" or "skip",
    "article_type": "experience_report" | "tool_release" | "technique" | "architecture" | "model_update" | "commentary" | "tutorial" | "announcement" | "other",
    "key_points": ["point 1", "point 2", "point 3"],
    "relevance_reason": "One sentence explaining your verdict",
    "practical_score": 1-5
}

practical_score: 5 = immediately actionable, 1 = tangentially related. Skip articles get 0.`

// Result holds the results of a triage run.
type Result struct {
	Processed int
	Relevant  int
	Skipped   int
	Errors    int
}

// Triager triages articles using LLM for relevance assessment.
type Triager struct {
	db       *database.DB
	provider llm.Provider
}

// NewTriager creates a new article triager.
func NewTriager(db *database.DB, provider llm.Provider) *Triager {
	return &Triager{db: db, provider: provider}
}

// TriageArticles triages all untriaged articles for a period.
func (t *Triager) TriageArticles(ctx context.Context, periodID string) *Result {
	if t.provider == nil {
		log.Println("No LLM provider available for triage")
		return &Result{Errors: 1}
	}

	articles, err := t.db.GetUntriagedArticles(&periodID)
	if err != nil {
		log.Printf("Error getting untriaged articles: %v", err)
		return &Result{Errors: 1}
	}

	if len(articles) == 0 {
		log.Println("No articles pending triage")
		return &Result{}
	}

	priorities, _ := t.db.GetActivePriorities()
	prioritiesText := formatPriorities(priorities)

	r := &Result{}
	for _, article := range articles {
		result, err := t.triageArticle(ctx, article, prioritiesText)
		if err != nil {
			log.Printf("Error triaging article %d: %v", article.ID, err)
			r.Errors++
			continue
		}

		if result == nil {
			r.Errors++
			continue
		}

		t.db.InsertTriage(article.ID, result.verdict, result.articleType, result.keyPoints, result.reason, result.practicalScore)
		r.Processed++
		if result.verdict == "relevant" {
			r.Relevant++
		} else {
			r.Skipped++
		}
		log.Printf("Triaged [%s]: %s", result.verdict, article.Title)
	}

	log.Printf("Triage complete: %d processed (%d relevant, %d skipped), %d errors",
		r.Processed, r.Relevant, r.Skipped, r.Errors)
	return r
}

type triageResult struct {
	verdict        string
	articleType    *string
	keyPoints      []string
	reason         *string
	practicalScore int
}

func (t *Triager) triageArticle(ctx context.Context, article database.Article, prioritiesText string) (*triageResult, error) {
	content := ""
	if article.Content != nil {
		content = *article.Content
	}
	if content == "" {
		content = article.Title
	}
	if len(content) > 4000 {
		content = content[:4000] + "..."
	}

	source := "Unknown"
	if article.Source != nil {
		source = *article.Source
	}

	prompt := fmt.Sprintf(triagePrompt, prioritiesText, article.Title, source, content)

	responseText, err := t.provider.Generate(ctx, prompt, 512)
	if err != nil {
		return nil, err
	}

	parsed := llm.ParseJSONResponse(responseText)
	if parsed == nil {
		// Default to relevant if we can't parse
		at := "other"
		reason := "LLM response could not be parsed"
		return &triageResult{
			verdict:        "relevant",
			articleType:    &at,
			keyPoints:      nil,
			reason:         &reason,
			practicalScore: 2,
		}, nil
	}

	verdict := strings.ToLower(getString(parsed, "verdict", "relevant"))
	if verdict != "relevant" && verdict != "skip" {
		verdict = "relevant"
	}

	at := getString(parsed, "article_type", "other")
	reason := getString(parsed, "relevance_reason", "")

	var keyPoints []string
	if kp, ok := parsed["key_points"]; ok {
		if arr, ok := kp.([]any); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					keyPoints = append(keyPoints, s)
				}
			}
			if len(keyPoints) > 5 {
				keyPoints = keyPoints[:5]
			}
		}
	}

	score := getInt(parsed, "practical_score", 2)
	if verdict == "skip" {
		score = 0
	} else if score < 1 {
		score = 1
	} else if score > 5 {
		score = 5
	}

	return &triageResult{
		verdict:        verdict,
		articleType:    &at,
		keyPoints:      keyPoints,
		reason:         &reason,
		practicalScore: score,
	}, nil
}

func formatPriorities(priorities []database.ResearchPriority) string {
	if len(priorities) == 0 {
		return "None defined"
	}
	var lines []string
	for _, p := range priorities {
		line := "- " + p.Title
		if p.Description != nil && *p.Description != "" {
			desc := *p.Description
			if len(desc) > 100 {
				desc = desc[:100]
			}
			line += ": " + desc
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func getString(m map[string]any, key, fallback string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return fallback
}

func getInt(m map[string]any, key string, fallback int) int {
	if v, ok := m[key]; ok {
		switch n := v.(type) {
		case float64:
			return int(n)
		case json.Number:
			if i, err := n.Int64(); err == nil {
				return int(i)
			}
		}
	}
	return fallback
}
