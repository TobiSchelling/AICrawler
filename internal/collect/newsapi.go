package collect

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const newsAPIBaseURL = "https://newsapi.org/v2/everything"

// NewsArticle represents an article from NewsAPI.
type NewsArticle struct {
	URL           string
	Title         string
	PublishedDate string
	Content       string
	Source        string
}

// NewsAPIClient fetches articles from NewsAPI.
type NewsAPIClient struct {
	apiKey string
	client *http.Client
}

// NewNewsAPIClient creates a new NewsAPI client.
func NewNewsAPIClient(apiKeyEnv string) *NewsAPIClient {
	return &NewsAPIClient{
		apiKey: os.Getenv(apiKeyEnv),
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// IsConfigured returns whether the API key is available.
func (c *NewsAPIClient) IsConfigured() bool {
	return c.apiKey != ""
}

// Search searches for articles matching a query.
func (c *NewsAPIClient) Search(query string, daysBack, pageSize int) []NewsArticle {
	if c.apiKey == "" {
		log.Println("NewsAPI not configured, skipping search")
		return nil
	}

	fromDate := time.Now().AddDate(0, 0, -daysBack).Format("2006-01-02")
	toDate := time.Now().Format("2006-01-02")

	if pageSize > 100 {
		pageSize = 100
	}

	params := url.Values{
		"q":        {query},
		"from":     {fromDate},
		"to":       {toDate},
		"language": {"en"},
		"pageSize": {fmt.Sprintf("%d", pageSize)},
		"sortBy":   {"relevancy"},
	}

	req, err := http.NewRequest("GET", newsAPIBaseURL+"?"+params.Encode(), nil)
	if err != nil {
		log.Printf("NewsAPI request error: %v", err)
		return nil
	}
	req.Header.Set("X-Api-Key", c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		log.Printf("NewsAPI error: %v", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("NewsAPI HTTP error: %d", resp.StatusCode)
		return nil
	}

	var result struct {
		Status   string `json:"status"`
		Articles []struct {
			URL         string `json:"url"`
			Title       string `json:"title"`
			PublishedAt string `json:"publishedAt"`
			Content     string `json:"content"`
			Description string `json:"description"`
			Source      struct {
				Name string `json:"name"`
			} `json:"source"`
		} `json:"articles"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("NewsAPI decode error: %v", err)
		return nil
	}

	if result.Status != "ok" {
		log.Printf("NewsAPI status: %s", result.Status)
		return nil
	}

	var articles []NewsArticle
	for _, a := range result.Articles {
		if a.URL == "" || a.Title == "" {
			continue
		}
		if a.Title == "[Removed]" || a.URL == "https://removed.com" {
			continue
		}

		var pubDate string
		if a.PublishedAt != "" {
			t, err := time.Parse(time.RFC3339, a.PublishedAt)
			if err == nil {
				pubDate = t.Format("2006-01-02")
			}
		}

		content := a.Content
		if content == "" {
			content = a.Description
		}
		content = strings.TrimSpace(content)

		source := "NewsAPI"
		if a.Source.Name != "" {
			source = a.Source.Name
		}

		articles = append(articles, NewsArticle{
			URL:           a.URL,
			Title:         strings.TrimSpace(a.Title),
			PublishedDate: pubDate,
			Content:       content,
			Source:        source,
		})
	}

	log.Printf("Fetched %d articles from NewsAPI for query: %s", len(articles), query)
	return articles
}

// SearchWithPriorities searches with base query and priority-enhanced queries.
func (c *NewsAPIClient) SearchWithPriorities(baseQuery string, priorities []string, daysBack int) []NewsArticle {
	seen := make(map[string]struct{})
	var all []NewsArticle

	for _, a := range c.Search(baseQuery, daysBack, 100) {
		if _, ok := seen[a.URL]; !ok {
			seen[a.URL] = struct{}{}
			all = append(all, a)
		}
	}

	for _, priority := range priorities {
		q := baseQuery + " " + priority
		for _, a := range c.Search(q, daysBack, 50) {
			if _, ok := seen[a.URL]; !ok {
				seen[a.URL] = struct{}{}
				all = append(all, a)
			}
		}
	}

	return all
}
