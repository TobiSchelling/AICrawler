package database

// Article represents a collected article.
type Article struct {
	ID             int64
	URL            string
	Title          string
	Source         *string
	PublishedDate  *string
	Content        *string
	ContentFetched bool
	PeriodID       *string
	CollectedAt    *string
}

// ArticleTriage holds triage results for an article.
type ArticleTriage struct {
	ArticleID       int64
	Verdict         string // "relevant" or "skip"
	ArticleType     *string
	KeyPoints       []string
	RelevanceReason *string
	PracticalScore  int
	TriagedAt       *string
}

// Storyline represents a cluster of related articles.
type Storyline struct {
	ID           int64
	PeriodID     string
	Label        string
	ArticleCount int
	CreatedAt    *string
}

// StorylineNarrative holds the LLM-generated narrative for a storyline.
type StorylineNarrative struct {
	ID               int64
	StorylineID      int64
	PeriodID         string
	Title            string
	NarrativeText    string
	SourceReferences []SourceReference
	GeneratedAt      *string
}

// SourceReference is a reference to an article in a narrative.
type SourceReference struct {
	Title        string `json:"title"`
	URL          string `json:"url"`
	Contribution string `json:"contribution,omitempty"`
}

// Briefing represents a complete briefing for a period.
type Briefing struct {
	ID             int64
	PeriodID       string
	TLDR           string
	BodyMarkdown   string
	StorylineCount int
	ArticleCount   int
	GeneratedAt    *string
}

// ResearchPriority is a user-defined research priority.
type ResearchPriority struct {
	ID          int64
	Title       string
	Description *string
	Keywords    []string
	IsActive    bool
	CreatedAt   *string
	UpdatedAt   *string
}

// RunReport holds metadata about a pipeline run.
type RunReport struct {
	ID             int64
	PeriodID       string
	GeneratedAt    *string
	ArticleCount   int
	StorylineCount int
}

// Stats contains aggregate database statistics.
type Stats struct {
	TotalArticles      int
	TriagedArticles    int
	RelevantArticles   int
	PeriodsWithArticles int
	Briefings          int
	Storylines         int
	TotalPriorities    int
	ActivePriorities   int
}

// TriageStats contains triage statistics for a period.
type TriageStats struct {
	Total    int
	Relevant int
	Skipped  int
}
