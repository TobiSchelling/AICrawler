package collect

import (
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
)

const maxPerFeed = 20

// FeedEntry represents a parsed feed entry.
type FeedEntry struct {
	URL           string
	Title         string
	PublishedDate string // YYYY-MM-DD or empty
	Content       string
	Source        string
}

// FeedConfig represents a single feed configuration.
type FeedConfig struct {
	URL  string
	Name string
}

// FeedParser parses RSS/Atom feeds.
type FeedParser struct {
	feeds []FeedConfig
}

// NewFeedParser creates a new FeedParser.
func NewFeedParser(feeds []FeedConfig) *FeedParser {
	return &FeedParser{feeds: feeds}
}

// ParseAll parses all configured feeds and returns entries within daysBack.
func (fp *FeedParser) ParseAll(daysBack int) []FeedEntry {
	cutoff := time.Now().AddDate(0, 0, -daysBack)
	var all []FeedEntry

	parser := gofeed.NewParser()
	for _, fc := range fp.feeds {
		name := fc.Name
		if name == "" {
			name = extractSourceName(fc.URL)
		}

		entries, err := parseFeed(parser, fc.URL, name, cutoff)
		if err != nil {
			log.Printf("Failed to parse feed %s: %v", fc.URL, err)
			continue
		}
		all = append(all, entries...)
		log.Printf("Parsed %d entries from %s (within %d days)", len(entries), name, daysBack)
	}

	return all
}

func parseFeed(parser *gofeed.Parser, feedURL, sourceName string, cutoff time.Time) ([]FeedEntry, error) {
	feed, err := parser.ParseURL(feedURL)
	if err != nil {
		return nil, err
	}

	var entries []FeedEntry
	for _, item := range feed.Items {
		if len(entries) >= maxPerFeed {
			break
		}

		entry := parseItem(item, sourceName)
		if entry == nil {
			continue
		}
		if isWithinWindow(entry.PublishedDate, cutoff) {
			entries = append(entries, *entry)
		}
	}

	return entries, nil
}

func parseItem(item *gofeed.Item, source string) *FeedEntry {
	itemURL := item.Link
	if itemURL == "" {
		itemURL = item.GUID
	}
	if itemURL == "" {
		return nil
	}

	title := strings.TrimSpace(item.Title)
	if title == "" {
		return nil
	}

	var publishedDate string
	if item.PublishedParsed != nil {
		publishedDate = item.PublishedParsed.Format("2006-01-02")
	} else if item.UpdatedParsed != nil {
		publishedDate = item.UpdatedParsed.Format("2006-01-02")
	}

	var content string
	if item.Content != "" {
		content = stripHTML(item.Content)
	} else if item.Description != "" {
		content = stripHTML(item.Description)
	}

	return &FeedEntry{
		URL:           itemURL,
		Title:         title,
		PublishedDate: publishedDate,
		Content:       content,
		Source:        source,
	}
}

func isWithinWindow(publishedDate string, cutoff time.Time) bool {
	if publishedDate == "" {
		return true // benefit of the doubt
	}
	pub, err := time.Parse("2006-01-02", publishedDate)
	if err != nil {
		return true
	}
	return !pub.Before(cutoff)
}

func stripHTML(text string) string {
	// Simple HTML tag removal
	var result strings.Builder
	inTag := false
	for _, r := range text {
		if r == '<' {
			inTag = true
			result.WriteRune(' ')
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteRune(r)
		}
	}

	s := result.String()
	// Decode common entities
	s = strings.ReplaceAll(s, "&nbsp;", " ")
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&quot;", `"`)
	s = strings.ReplaceAll(s, "&#39;", "'")

	// Normalize whitespace
	fields := strings.Fields(s)
	return strings.Join(fields, " ")
}

func extractSourceName(feedURL string) string {
	u, err := url.Parse(feedURL)
	if err != nil {
		return feedURL
	}
	host := strings.ToLower(u.Hostname())

	for _, prefix := range []string{"www.", "blog.", "blogs.", "rss.", "feeds."} {
		host = strings.TrimPrefix(host, prefix)
	}

	parts := strings.Split(host, ".")
	if len(parts) >= 2 {
		name := parts[len(parts)-2]
		return strings.ToUpper(name[:1]) + name[1:]
	}
	return strings.ToUpper(host[:1]) + host[1:]
}
