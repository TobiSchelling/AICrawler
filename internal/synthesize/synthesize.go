package synthesize

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/TobiSchelling/AICrawler/internal/database"
	"github.com/TobiSchelling/AICrawler/internal/llm"
)

const brieflyNotedLabel = "Briefly Noted"

const synthesisPrompt = `You are writing one section of a daily AI news briefing for software practitioners.

This section covers a storyline about: %s

Write a cohesive 2-3 paragraph narrative that weaves these articles together. Write as if you're a well-informed colleague explaining what happened recently. Be specific about tools, techniques, and outcomes. Avoid marketing language.

Articles in this storyline:
%s

Respond with ONLY this JSON:
{
    "title": "A compelling 5-8 word section title",
    "narrative": "Your 2-3 paragraph narrative here. Use markdown for emphasis.",
    "source_references": [
        {"title": "Article Title", "url": "https://...", "contribution": "What this article added to the story"}
    ]
}`

// Result holds the results of a synthesis run.
type Result struct {
	NarrativesCreated int
	Errors            int
}

// Synthesizer synthesizes narratives for each storyline using LLM.
type Synthesizer struct {
	db       *database.DB
	provider llm.Provider
}

// NewSynthesizer creates a new storyline synthesizer.
func NewSynthesizer(db *database.DB, provider llm.Provider) *Synthesizer {
	return &Synthesizer{db: db, provider: provider}
}

// SynthesizePeriod synthesizes narratives for all storylines in a period.
func (s *Synthesizer) SynthesizePeriod(ctx context.Context, periodID string) *Result {
	if s.provider == nil {
		log.Println("No LLM provider available for synthesis")
		return &Result{Errors: 1}
	}

	storylines, err := s.db.GetStorylinesForPeriod(periodID)
	if err != nil {
		log.Printf("Error getting storylines: %v", err)
		return &Result{Errors: 1}
	}
	if len(storylines) == 0 {
		log.Printf("No storylines to synthesize for %s", periodID)
		return &Result{}
	}

	r := &Result{}
	for _, storyline := range storylines {
		existing, _ := s.db.GetNarrativeForStoryline(storyline.ID)
		if existing != nil {
			r.NarrativesCreated++
			continue
		}

		articles, _ := s.db.GetStorylineArticles(storyline.ID)
		if len(articles) == 0 {
			continue
		}

		var synthErr error
		if storyline.Label == brieflyNotedLabel {
			synthErr = s.synthesizeBrieflyNoted(storyline, articles, periodID)
		} else {
			synthErr = s.synthesizeStoryline(ctx, storyline, articles, periodID)
		}

		if synthErr != nil {
			log.Printf("Error synthesizing storyline %d: %v", storyline.ID, synthErr)
			r.Errors++
		} else {
			r.NarrativesCreated++
		}
	}

	log.Printf("Synthesis complete: %d narratives created, %d errors", r.NarrativesCreated, r.Errors)
	return r
}

func (s *Synthesizer) synthesizeStoryline(ctx context.Context, storyline database.Storyline, articles []database.Article, periodID string) error {
	articlesText := s.formatArticles(articles)
	prompt := fmt.Sprintf(synthesisPrompt, storyline.Label, articlesText)

	responseText, err := s.provider.Generate(ctx, prompt, 1024)
	if err != nil {
		return err
	}

	parsed := llm.ParseJSONResponse(responseText)

	var title, narrative string
	var refs []database.SourceReference

	if parsed != nil {
		title = getStr(parsed, "title", storyline.Label)
		narrative = getStr(parsed, "narrative", "")
		refs = parseSourceRefs(parsed)
	} else {
		title = storyline.Label
		narrative = strings.TrimSpace(responseText)
		for _, a := range articles {
			refs = append(refs, database.SourceReference{Title: a.Title, URL: a.URL})
		}
	}

	_, err = s.db.InsertStorylineNarrative(storyline.ID, periodID, title, narrative, refs)
	return err
}

func (s *Synthesizer) synthesizeBrieflyNoted(storyline database.Storyline, articles []database.Article, periodID string) error {
	var bullets []string
	var refs []database.SourceReference

	for _, article := range articles {
		triage, _ := s.db.GetTriage(article.ID)
		point := article.Title
		if triage != nil && len(triage.KeyPoints) > 0 {
			point = triage.KeyPoints[0]
		}

		source := "Unknown"
		if article.Source != nil {
			source = *article.Source
		}
		bullets = append(bullets, fmt.Sprintf("- **%s** (%s): %s", article.Title, source, point))
		refs = append(refs, database.SourceReference{Title: article.Title, URL: article.URL})
	}

	narrative := strings.Join(bullets, "\n")
	_, err := s.db.InsertStorylineNarrative(storyline.ID, periodID, brieflyNotedLabel, narrative, refs)
	return err
}

func (s *Synthesizer) formatArticles(articles []database.Article) string {
	var parts []string
	for i, article := range articles {
		triage, _ := s.db.GetTriage(article.ID)
		var keyPoints string
		if triage != nil && len(triage.KeyPoints) > 0 {
			keyPoints = "\n  Key points: " + strings.Join(triage.KeyPoints, "; ")
		}

		var contentPreview string
		if article.Content != nil {
			content := *article.Content
			if len(content) > 300 {
				content = content[:300]
			}
			contentPreview = fmt.Sprintf("\n  Content: %s...", content)
		}

		source := "Unknown"
		if article.Source != nil {
			source = *article.Source
		}

		parts = append(parts, fmt.Sprintf("[%d] %s\n  Source: %s\n  URL: %s%s%s",
			i+1, article.Title, source, article.URL, keyPoints, contentPreview))
	}
	return strings.Join(parts, "\n\n")
}

func getStr(m map[string]any, key, fallback string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return fallback
}

func parseSourceRefs(m map[string]any) []database.SourceReference {
	refsRaw, ok := m["source_references"]
	if !ok {
		return nil
	}
	arr, ok := refsRaw.([]any)
	if !ok {
		return nil
	}

	var refs []database.SourceReference
	for _, item := range arr {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		ref := database.SourceReference{
			Title:        getStr(obj, "title", ""),
			URL:          getStr(obj, "url", ""),
			Contribution: getStr(obj, "contribution", ""),
		}
		refs = append(refs, ref)
	}
	return refs
}
