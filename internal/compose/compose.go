package compose

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/TobiSchelling/AICrawler/internal/database"
	"github.com/TobiSchelling/AICrawler/internal/llm"
)

const brieflyNotedLabel = "Briefly Noted"

const composePrompt = `You are writing the TL;DR for a daily AI news briefing aimed at software practitioners.

Here are today's storylines and their narratives:

%s

Write a TL;DR section (3-5 bullet points) that captures the most important takeaways from ALL storylines. Each bullet should be one sentence that tells the reader what happened and why it matters.

Respond with ONLY this JSON:
{
    "tldr_bullets": [
        "First key takeaway",
        "Second key takeaway",
        "Third key takeaway"
    ]
}`

// Composer composes the final briefing from storyline narratives.
type Composer struct {
	db       *database.DB
	provider llm.Provider
}

// NewComposer creates a new briefing composer.
func NewComposer(db *database.DB, provider llm.Provider) *Composer {
	return &Composer{db: db, provider: provider}
}

// ComposeBriefing composes a complete briefing for a period.
func (c *Composer) ComposeBriefing(ctx context.Context, periodID string) (*database.Briefing, error) {
	narratives, err := c.db.GetNarrativesForPeriod(periodID)
	if err != nil {
		return nil, err
	}
	storylines, err := c.db.GetStorylinesForPeriod(periodID)
	if err != nil {
		return nil, err
	}

	if len(narratives) == 0 {
		log.Printf("No narratives found for %s", periodID)
		return c.storeEmptyBriefing(periodID)
	}

	tldr := c.generateTLDR(ctx, narratives)
	body := assembleBody(narratives)

	var articleCount int
	for _, s := range storylines {
		articleCount += s.ArticleCount
	}

	c.db.InsertBriefing(periodID, tldr, body, len(storylines), articleCount)
	c.db.InsertReport(periodID, articleCount, len(storylines))

	briefing, err := c.db.GetBriefing(periodID)
	if err != nil {
		return nil, err
	}
	log.Printf("Briefing composed for %s: %d storylines", periodID, len(storylines))
	return briefing, nil
}

func (c *Composer) generateTLDR(ctx context.Context, narratives []database.StorylineNarrative) string {
	if c.provider == nil {
		return fallbackTLDR(narratives)
	}

	var parts []string
	for _, n := range narratives {
		if n.Title != brieflyNotedLabel {
			parts = append(parts, fmt.Sprintf("## %s\n%s", n.Title, n.NarrativeText))
		}
	}

	prompt := fmt.Sprintf(composePrompt, strings.Join(parts, "\n\n"))
	responseText, err := c.provider.Generate(ctx, prompt, 512)
	if err != nil || responseText == "" {
		return fallbackTLDR(narratives)
	}

	parsed := llm.ParseJSONResponse(responseText)
	if parsed != nil {
		if bullets, ok := parsed["tldr_bullets"]; ok {
			if arr, ok := bullets.([]any); ok {
				var lines []string
				for _, b := range arr {
					if s, ok := b.(string); ok {
						lines = append(lines, "- "+s)
					}
				}
				return strings.Join(lines, "\n")
			}
		}
	}

	return strings.TrimSpace(responseText)
}

func fallbackTLDR(narratives []database.StorylineNarrative) string {
	var bullets []string
	for _, n := range narratives {
		if n.Title != brieflyNotedLabel {
			bullets = append(bullets, "- "+n.Title)
		}
	}
	if len(bullets) == 0 {
		return "- No significant storylines today."
	}
	return strings.Join(bullets, "\n")
}

func assembleBody(narratives []database.StorylineNarrative) string {
	var mainNarratives, brieflyNoted []database.StorylineNarrative
	for _, n := range narratives {
		if n.Title == brieflyNotedLabel {
			brieflyNoted = append(brieflyNoted, n)
		} else {
			mainNarratives = append(mainNarratives, n)
		}
	}

	var sections []string
	for _, n := range mainNarratives {
		section := fmt.Sprintf("## %s\n\n%s", n.Title, n.NarrativeText)
		if len(n.SourceReferences) > 0 {
			var refs []string
			for _, ref := range n.SourceReferences {
				line := fmt.Sprintf("- [%s](%s)", ref.Title, ref.URL)
				if ref.Contribution != "" {
					line += " — " + ref.Contribution
				}
				refs = append(refs, line)
			}
			section += "\n\n**Sources:**\n" + strings.Join(refs, "\n")
		}
		sections = append(sections, section)
	}

	for _, n := range brieflyNoted {
		sections = append(sections, fmt.Sprintf("## %s\n\n%s", n.Title, n.NarrativeText))
	}

	return strings.Join(sections, "\n\n---\n\n")
}

func (c *Composer) storeEmptyBriefing(periodID string) (*database.Briefing, error) {
	c.db.InsertBriefing(periodID, "- No articles collected today.", "No briefing content available for this period.", 0, 0)
	return c.db.GetBriefing(periodID)
}
