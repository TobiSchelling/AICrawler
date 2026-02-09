package cluster

import (
	"context"
	"log"
	"strings"

	"github.com/TobiSchelling/AICrawler/internal/database"
	"github.com/TobiSchelling/AICrawler/internal/llm"
)

const (
	BrieflyNotedLabel         = "Briefly Noted"
	DefaultDistanceThreshold = 1.2
)

// Result holds the results of a clustering run.
type Result struct {
	StorylineCount    int
	ArticleCount      int
	BrieflyNotedCount int
}

// Clusterer clusters relevant articles into storylines using embeddings.
type Clusterer struct {
	db                *database.DB
	embedder          llm.Embedder
	distanceThreshold float64
}

// NewClusterer creates a new article clusterer.
func NewClusterer(db *database.DB, embedder llm.Embedder, distanceThreshold float64) *Clusterer {
	if distanceThreshold <= 0 {
		distanceThreshold = DefaultDistanceThreshold
	}
	return &Clusterer{
		db:                db,
		embedder:          embedder,
		distanceThreshold: distanceThreshold,
	}
}

// ClusterArticles clusters relevant articles for a period into storylines.
func (c *Clusterer) ClusterArticles(ctx context.Context, periodID string) (*Result, error) {
	articles, err := c.db.GetRelevantArticles(periodID)
	if err != nil {
		return nil, err
	}

	if len(articles) == 0 {
		log.Printf("No relevant articles to cluster for %s", periodID)
		return &Result{}, nil
	}

	// Clear existing storylines for re-clustering
	if err := c.db.ClearStorylinesForPeriod(periodID); err != nil {
		return nil, err
	}

	if len(articles) < 2 {
		// Single article -> Briefly Noted
		ids := make([]int64, len(articles))
		for i, a := range articles {
			ids[i] = a.ID
		}
		c.db.InsertStoryline(periodID, BrieflyNotedLabel, ids)
		return &Result{
			StorylineCount:    1,
			ArticleCount:      len(articles),
			BrieflyNotedCount: len(articles),
		}, nil
	}

	// Build text representations for embedding
	texts := make([]string, len(articles))
	for i, a := range articles {
		texts[i] = c.articleText(a)
	}

	// Generate embeddings
	log.Printf("Generating embeddings for %d articles...", len(articles))
	embeddings, err := c.embedder.Embed(ctx, texts)
	if err != nil {
		return nil, err
	}

	// Cluster using Ward's linkage
	clusterLabels := c.clusterEmbeddings(embeddings)

	// Group articles by cluster
	groups := make(map[int][]database.Article)
	for i, label := range clusterLabels {
		groups[label] = append(groups[label], articles[i])
	}

	// Separate real storylines from singletons
	var storylines [][]database.Article
	var brieflyNoted []database.Article

	for _, group := range groups {
		if len(group) >= 2 {
			storylines = append(storylines, group)
		} else {
			brieflyNoted = append(brieflyNoted, group...)
		}
	}

	// Store storylines
	for _, group := range storylines {
		label := generateLabel(group)
		ids := make([]int64, len(group))
		for i, a := range group {
			ids[i] = a.ID
		}
		c.db.InsertStoryline(periodID, label, ids)
	}

	// Store Briefly Noted
	brieflyNotedCount := 0
	if len(brieflyNoted) > 0 {
		ids := make([]int64, len(brieflyNoted))
		for i, a := range brieflyNoted {
			ids[i] = a.ID
		}
		c.db.InsertStoryline(periodID, BrieflyNotedLabel, ids)
		brieflyNotedCount = len(brieflyNoted)
	}

	totalStorylines := len(storylines)
	if brieflyNotedCount > 0 {
		totalStorylines++
	}

	log.Printf("Clustering complete: %d storylines + %d briefly noted from %d articles",
		len(storylines), brieflyNotedCount, len(articles))

	return &Result{
		StorylineCount:    totalStorylines,
		ArticleCount:      len(articles),
		BrieflyNotedCount: brieflyNotedCount,
	}, nil
}

func (c *Clusterer) articleText(article database.Article) string {
	parts := []string{article.Title}

	triage, _ := c.db.GetTriage(article.ID)
	if triage != nil && len(triage.KeyPoints) > 0 {
		parts = append(parts, triage.KeyPoints...)
	}

	if article.Content != nil {
		content := *article.Content
		if len(content) > 500 {
			content = content[:500]
		}
		parts = append(parts, content)
	}

	return strings.Join(parts, " ")
}

func (c *Clusterer) clusterEmbeddings(embeddings [][]float64) []int {
	dist := pairwiseDistances(embeddings)
	merges := wardLinkage(dist, len(embeddings))
	return cutDendrogram(merges, len(embeddings), c.distanceThreshold)
}

func generateLabel(articles []database.Article) string {
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "is": true, "are": true, "was": true,
		"were": true, "be": true, "been": true, "being": true, "have": true, "has": true,
		"had": true, "do": true, "does": true, "did": true, "will": true, "would": true,
		"could": true, "should": true, "may": true, "might": true, "can": true, "shall": true,
		"to": true, "of": true, "in": true, "for": true, "on": true, "with": true, "at": true,
		"by": true, "from": true, "as": true, "into": true, "through": true, "during": true,
		"before": true, "after": true, "above": true, "below": true, "and": true, "but": true,
		"or": true, "nor": true, "not": true, "so": true, "yet": true, "both": true,
		"either": true, "neither": true, "each": true, "every": true, "all": true, "any": true,
		"few": true, "more": true, "most": true, "other": true, "some": true, "such": true,
		"no": true, "only": true, "own": true, "same": true, "than": true, "too": true,
		"very": true, "just": true, "how": true, "what": true, "which": true, "who": true,
		"whom": true, "this": true, "that": true, "these": true, "those": true, "it": true,
		"its": true, "new": true, "about": true, "up": true, "out": true, "one": true,
		"two": true, "also": true, "like": true, "get": true, "use": true,
	}

	wordCounts := make(map[string]int)
	for _, article := range articles {
		words := strings.Fields(strings.ToLower(article.Title))
		for _, word := range words {
			word = strings.Trim(word, ".,!?:;\"'()-[]")
			if len(word) > 2 && !stopWords[word] {
				wordCounts[word]++
			}
		}
	}

	// Find top 3 words
	var topWords []string
	for i := 0; i < 3; i++ {
		maxCount := 0
		maxWord := ""
		for word, count := range wordCounts {
			if count > maxCount {
				maxCount = count
				maxWord = word
			}
		}
		if maxWord != "" {
			topWords = append(topWords, strings.Title(maxWord)) //nolint: staticcheck
			delete(wordCounts, maxWord)
		}
	}

	if len(topWords) > 0 {
		return strings.Join(topWords, " ")
	}

	// Fallback: first article title truncated
	title := articles[0].Title
	if len(title) > 50 {
		title = title[:50]
	}
	return title
}
