package collect

import (
	"log"

	"github.com/TobiSchelling/AICrawler/internal/config"
	"github.com/TobiSchelling/AICrawler/internal/database"
)

// Result holds the results of a collection run.
type Result struct {
	TotalFound  int
	NewArticles int
	Duplicates  int
	Sources     map[string]int
}

// Collector orchestrates article collection from RSS feeds and NewsAPI.
type Collector struct {
	db         *database.DB
	feedParser *FeedParser
	newsClient *NewsAPIClient
	newsQuery  string
	daysBack   int
}

// NewCollector creates a new article collector.
func NewCollector(cfg *config.Config, db *database.DB, daysBack int) *Collector {
	c := &Collector{
		db:       db,
		daysBack: daysBack,
	}

	// Set up feed parser
	if len(cfg.Sources.Feeds) > 0 {
		feeds := make([]FeedConfig, len(cfg.Sources.Feeds))
		for i, f := range cfg.Sources.Feeds {
			feeds[i] = FeedConfig{URL: f.URL, Name: f.Name}
		}
		c.feedParser = NewFeedParser(feeds)
	}

	// Set up NewsAPI client
	apiCfg := cfg.Sources.APIs.NewsAPI
	if apiCfg.Enabled {
		c.newsClient = NewNewsAPIClient(apiCfg.APIKeyEnv)
		c.newsQuery = apiCfg.Query
		if c.newsQuery == "" {
			c.newsQuery = "artificial intelligence software development"
		}
	}

	return c
}

// Collect collects articles from all configured sources.
func (c *Collector) Collect(periodID string) *Result {
	r := &Result{Sources: make(map[string]int)}

	// Collect from RSS feeds
	if c.feedParser != nil {
		log.Println("Collecting from RSS feeds...")
		entries := c.feedParser.ParseAll(c.daysBack)
		r.TotalFound += len(entries)

		for _, entry := range entries {
			var source, pubDate, content *string
			if entry.Source != "" {
				source = &entry.Source
			}
			if entry.PublishedDate != "" {
				pubDate = &entry.PublishedDate
			}
			if entry.Content != "" {
				content = &entry.Content
			}
			pid := periodID

			id, _ := c.db.InsertArticle(entry.URL, entry.Title, source, pubDate, content, &pid)
			if id > 0 {
				r.NewArticles++
				r.Sources[entry.Source]++
			} else {
				r.Duplicates++
			}
		}
	}

	// Collect from NewsAPI
	if c.newsClient != nil && c.newsClient.IsConfigured() {
		log.Println("Collecting from NewsAPI...")

		priorities, _ := c.db.GetActivePriorities()
		var priorityTitles []string
		for _, p := range priorities {
			priorityTitles = append(priorityTitles, p.Title)
		}

		var articles []NewsArticle
		if len(priorityTitles) > 0 {
			log.Printf("Using %d active priorities for search", len(priorityTitles))
			articles = c.newsClient.SearchWithPriorities(c.newsQuery, priorityTitles, c.daysBack)
		} else {
			articles = c.newsClient.Search(c.newsQuery, c.daysBack, 100)
		}

		r.TotalFound += len(articles)

		for _, article := range articles {
			var source, pubDate, content *string
			if article.Source != "" {
				source = &article.Source
			}
			if article.PublishedDate != "" {
				pubDate = &article.PublishedDate
			}
			if article.Content != "" {
				content = &article.Content
			}
			pid := periodID

			id, _ := c.db.InsertArticle(article.URL, article.Title, source, pubDate, content, &pid)
			if id > 0 {
				r.NewArticles++
				r.Sources[article.Source]++
			} else {
				r.Duplicates++
			}
		}
	}

	log.Printf("Collection complete: %d found, %d new, %d duplicates", r.TotalFound, r.NewArticles, r.Duplicates)
	return r
}
