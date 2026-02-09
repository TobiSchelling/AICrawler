package pipeline

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/TobiSchelling/AICrawler/internal/cluster"
	"github.com/TobiSchelling/AICrawler/internal/collect"
	"github.com/TobiSchelling/AICrawler/internal/compose"
	"github.com/TobiSchelling/AICrawler/internal/config"
	"github.com/TobiSchelling/AICrawler/internal/database"
	"github.com/TobiSchelling/AICrawler/internal/fetch"
	"github.com/TobiSchelling/AICrawler/internal/llm"
	"github.com/TobiSchelling/AICrawler/internal/synthesize"
	"github.com/TobiSchelling/AICrawler/internal/triage"
)

// StepResult holds the result of a single pipeline step.
type StepResult struct {
	Name    string
	Summary string
	Err     error
}

// Result holds the results of a full pipeline run.
type Result struct {
	PeriodID string
	Steps    []StepResult
}

// Pipeline orchestrates the 6-step briefing generation pipeline.
type Pipeline struct {
	cfg      *config.Config
	db       *database.DB
	provider llm.Provider
	embedder llm.Embedder
}

// New creates a new pipeline.
func New(cfg *config.Config, db *database.DB) *Pipeline {
	summ := cfg.Summarization
	provider := llm.CreateProvider(
		summ.Provider,
		summ.Model,
		summ.OllamaURL,
		summ.OpenAIModel,
		summ.APIKeyEnv,
	)

	var embedder llm.Embedder
	embModel := summ.EmbeddingModel
	if embModel == "" {
		embModel = "nomic-embed-text"
	}
	baseURL := summ.OllamaURL
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	embedder = llm.NewOllamaEmbedder(embModel, baseURL)

	return &Pipeline{
		cfg:      cfg,
		db:       db,
		provider: provider,
		embedder: embedder,
	}
}

// Run executes the full 6-step pipeline.
func (p *Pipeline) Run(ctx context.Context, periodID string, daysBack int) *Result {
	r := &Result{PeriodID: periodID}

	// Step 1: Collect
	step := p.runCollect(periodID, daysBack)
	r.Steps = append(r.Steps, step)
	if step.Err != nil {
		return r
	}

	// Step 2: Fetch content
	step = p.runFetch(periodID)
	r.Steps = append(r.Steps, step)

	// Step 3: Triage
	step = p.runTriage(ctx, periodID)
	r.Steps = append(r.Steps, step)

	// Step 4: Cluster
	step = p.runCluster(ctx, periodID)
	r.Steps = append(r.Steps, step)
	if step.Err != nil {
		return r
	}

	// Step 5: Synthesize
	step = p.runSynthesize(ctx, periodID)
	r.Steps = append(r.Steps, step)

	// Step 6: Compose
	step = p.runCompose(ctx, periodID)
	r.Steps = append(r.Steps, step)

	return r
}

// DryRun shows what would be done without executing.
func (p *Pipeline) DryRun(periodID string) *Result {
	r := &Result{PeriodID: periodID}

	articles, _ := p.db.GetArticlesForPeriod(periodID)
	r.Steps = append(r.Steps, StepResult{
		Name:    "Collect",
		Summary: fmt.Sprintf("[dry-run] %d articles already in DB for %s", len(articles), periodID),
	})

	needing, _ := p.db.GetArticlesNeedingFetch(&periodID)
	r.Steps = append(r.Steps, StepResult{
		Name:    "Fetch",
		Summary: fmt.Sprintf("[dry-run] %d articles need content fetching", len(needing)),
	})

	untriaged, _ := p.db.GetUntriagedArticles(&periodID)
	r.Steps = append(r.Steps, StepResult{
		Name:    "Triage",
		Summary: fmt.Sprintf("[dry-run] %d articles need triage", len(untriaged)),
	})

	relevant, _ := p.db.GetRelevantArticles(periodID)
	r.Steps = append(r.Steps, StepResult{
		Name:    "Cluster",
		Summary: fmt.Sprintf("[dry-run] %d relevant articles to cluster", len(relevant)),
	})

	storylines, _ := p.db.GetStorylinesForPeriod(periodID)
	r.Steps = append(r.Steps, StepResult{
		Name:    "Synthesize",
		Summary: fmt.Sprintf("[dry-run] %d storylines need narratives", len(storylines)),
	})

	briefing, _ := p.db.GetBriefing(periodID)
	if briefing != nil {
		r.Steps = append(r.Steps, StepResult{
			Name:    "Compose",
			Summary: fmt.Sprintf("[dry-run] Briefing already exists for %s", periodID),
		})
	} else {
		r.Steps = append(r.Steps, StepResult{
			Name:    "Compose",
			Summary: fmt.Sprintf("[dry-run] Would compose briefing for %s", periodID),
		})
	}

	return r
}

func (p *Pipeline) runCollect(periodID string, daysBack int) StepResult {
	log.Println("Step 1/6: Collecting articles...")
	collector := collect.NewCollector(p.cfg, p.db, daysBack)
	result := collector.Collect(periodID)
	return StepResult{
		Name:    "Collect",
		Summary: fmt.Sprintf("Found %d new articles (%d total, %d duplicates)", result.NewArticles, result.TotalFound, result.Duplicates),
	}
}

func (p *Pipeline) runFetch(periodID string) StepResult {
	log.Println("Step 2/6: Fetching article content...")
	fetcher := fetch.NewContentFetcher(p.db, 15*time.Second)
	result := fetcher.FetchMissingContent(&periodID)
	return StepResult{
		Name:    "Fetch",
		Summary: fmt.Sprintf("Fetched %d articles, %d failed", result.Fetched, result.Failed),
	}
}

func (p *Pipeline) runTriage(ctx context.Context, periodID string) StepResult {
	log.Println("Step 3/6: Triaging articles...")
	triager := triage.NewTriager(p.db, p.provider)
	result := triager.TriageArticles(ctx, periodID)
	return StepResult{
		Name:    "Triage",
		Summary: fmt.Sprintf("Triaged %d articles: %d relevant, %d skipped", result.Processed, result.Relevant, result.Skipped),
	}
}

func (p *Pipeline) runCluster(ctx context.Context, periodID string) StepResult {
	log.Println("Step 4/6: Clustering into storylines...")
	clusterer := cluster.NewClusterer(p.db, p.embedder, 0)
	result, err := clusterer.ClusterArticles(ctx, periodID)
	if err != nil {
		return StepResult{Name: "Cluster", Err: err}
	}
	return StepResult{
		Name:    "Cluster",
		Summary: fmt.Sprintf("Created %d storylines from %d articles", result.StorylineCount, result.ArticleCount),
	}
}

func (p *Pipeline) runSynthesize(ctx context.Context, periodID string) StepResult {
	log.Println("Step 5/6: Synthesizing narratives...")
	synth := synthesize.NewSynthesizer(p.db, p.provider)
	result := synth.SynthesizePeriod(ctx, periodID)
	return StepResult{
		Name:    "Synthesize",
		Summary: fmt.Sprintf("Synthesized %d narratives", result.NarrativesCreated),
	}
}

func (p *Pipeline) runCompose(ctx context.Context, periodID string) StepResult {
	log.Println("Step 6/6: Composing briefing...")
	comp := compose.NewComposer(p.db, p.provider)
	briefing, err := comp.ComposeBriefing(ctx, periodID)
	if err != nil {
		return StepResult{Name: "Compose", Err: err}
	}
	return StepResult{
		Name:    "Compose",
		Summary: fmt.Sprintf("Briefing composed: %d storylines, %d articles", briefing.StorylineCount, briefing.ArticleCount),
	}
}
