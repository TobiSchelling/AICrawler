package fetch

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	readability "github.com/go-shiori/go-readability"

	"github.com/TobiSchelling/AICrawler/internal/database"
)

// Result holds the results of a content fetch run.
type Result struct {
	Fetched          int
	AlreadyHadContent int
	Failed           int
}

// ContentFetcher fetches full article text via HTTP + readability extraction.
type ContentFetcher struct {
	db     *database.DB
	client *http.Client
}

// NewContentFetcher creates a new content fetcher.
func NewContentFetcher(db *database.DB, timeout time.Duration) *ContentFetcher {
	if timeout == 0 {
		timeout = 15 * time.Second
	}
	return &ContentFetcher{
		db: db,
		client: &http.Client{
			Timeout: timeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return http.ErrUseLastResponse
				}
				return nil
			},
		},
	}
}

// FetchMissingContent fetches content for articles that have empty content.
func (f *ContentFetcher) FetchMissingContent(periodID *string) *Result {
	articles, err := f.db.GetArticlesNeedingFetch(periodID)
	if err != nil {
		log.Printf("Error getting articles needing fetch: %v", err)
		return &Result{}
	}

	if len(articles) == 0 {
		log.Println("No articles need content fetching")
		return &Result{}
	}

	result := &Result{}
	failedDomains := make(map[string]struct{})

	for _, article := range articles {
		u, _ := url.Parse(article.URL)
		domain := ""
		if u != nil {
			domain = strings.ToLower(u.Host)
		}

		if _, failed := failedDomains[domain]; failed {
			f.db.MarkArticleFetchAttempted(article.ID)
			result.Failed++
			continue
		}

		content, httpErr := f.fetchArticleContent(article.URL)
		if httpErr != nil {
			f.db.MarkArticleFetchAttempted(article.ID)
			result.Failed++
			if domain != "" {
				failedDomains[domain] = struct{}{}
			}
			log.Printf("HTTP error for %s — skipping remaining from %s", article.URL, domain)
			continue
		}

		if content != "" {
			f.db.UpdateArticleContent(article.ID, &content)
			result.Fetched++
			log.Printf("Fetched content for: %s", article.Title)
		} else {
			f.db.MarkArticleFetchAttempted(article.ID)
			result.Failed++
			log.Printf("No extractable content from: %s", article.URL)
		}
	}

	log.Printf("Content fetch complete: %d fetched, %d failed", result.Fetched, result.Failed)
	return result
}

func (f *ContentFetcher) fetchArticleContent(articleURL string) (string, error) {
	req, err := http.NewRequest("GET", articleURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "AICrawler/1.0 (news aggregator)")

	resp, err := f.client.Do(req)
	if err != nil {
		return "", nil // connection error, not HTTP error
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", &httpError{code: resp.StatusCode}
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil
	}

	parsedURL, _ := url.Parse(articleURL)
	article, err := readability.FromReader(strings.NewReader(string(bodyBytes)), parsedURL)
	if err != nil {
		return "", nil
	}

	text := strings.TrimSpace(article.TextContent)
	if len(text) > 100 {
		return text, nil
	}
	return "", nil
}

type httpError struct {
	code int
}

func (e *httpError) Error() string {
	return http.StatusText(e.code)
}
