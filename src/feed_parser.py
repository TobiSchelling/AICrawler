"""RSS/Atom feed parser for AICrawler."""

import logging
from dataclasses import dataclass
from datetime import datetime, timedelta
from time import mktime

import feedparser

logger = logging.getLogger(__name__)

MAX_PER_FEED = 20


@dataclass
class FeedEntry:
    """Represents a parsed feed entry."""

    url: str
    title: str
    published_date: str | None
    content: str | None
    source: str


class FeedParser:
    """Parses RSS/Atom feeds and extracts articles."""

    def __init__(self, feeds: list[dict]):
        """
        Initialize with feed configurations.

        Args:
            feeds: List of dicts with 'url' and optional 'name' keys.
        """
        self.feeds = feeds

    def parse_all(self, days_back: int = 1) -> list[FeedEntry]:
        """Parse all configured feeds and return entries.

        Args:
            days_back: Only include entries published within this many days.
        """
        all_entries: list[FeedEntry] = []
        cutoff = datetime.now() - timedelta(days=days_back)

        for feed_config in self.feeds:
            url = feed_config["url"]
            name = feed_config.get("name", self._extract_source_name(url))

            try:
                entries = self._parse_feed(url, name, cutoff)
                all_entries.extend(entries)
                logger.info("Parsed %d entries from %s (within %d days)", len(entries), name, days_back)
            except Exception as e:
                logger.error("Failed to parse feed %s: %s", url, e)

        return all_entries

    def _parse_feed(self, url: str, source_name: str, cutoff: datetime) -> list[FeedEntry]:
        """Parse a single feed URL, filtering by date and capping at MAX_PER_FEED."""
        feed = feedparser.parse(url)

        if feed.bozo and feed.bozo_exception:
            logger.warning("Feed parsing warning for %s: %s", url, feed.bozo_exception)

        entries = []
        for entry in feed.entries:
            if len(entries) >= MAX_PER_FEED:
                break

            try:
                parsed = self._parse_entry(entry, source_name)
                if parsed and self._is_within_window(parsed.published_date, cutoff):
                    entries.append(parsed)
            except Exception as e:
                logger.warning("Failed to parse entry: %s", e)

        return entries

    def _is_within_window(self, published_date: str | None, cutoff: datetime) -> bool:
        """Check if an article's published date is within the collection window."""
        if not published_date:
            # No date available — include it (benefit of the doubt)
            return True
        try:
            pub = datetime.strptime(published_date, "%Y-%m-%d")
            return pub >= cutoff
        except ValueError:
            return True

    def _parse_entry(self, entry: feedparser.FeedParserDict, source: str) -> FeedEntry | None:
        """Parse a single feed entry."""
        # Get URL (required)
        url = entry.get("link") or entry.get("id")
        if not url:
            return None

        # Get title (required)
        title = entry.get("title", "").strip()
        if not title:
            return None

        # Get published date
        published_date = None
        if hasattr(entry, "published_parsed") and entry.published_parsed:
            published_date = datetime.fromtimestamp(
                mktime(entry.published_parsed)
            ).strftime("%Y-%m-%d")
        elif hasattr(entry, "updated_parsed") and entry.updated_parsed:
            published_date = datetime.fromtimestamp(
                mktime(entry.updated_parsed)
            ).strftime("%Y-%m-%d")

        # Get content (try multiple fields)
        content = None
        if hasattr(entry, "content") and entry.content:
            content = entry.content[0].get("value", "")
        elif hasattr(entry, "summary"):
            content = entry.summary
        elif hasattr(entry, "description"):
            content = entry.description

        # Clean HTML tags from content (basic)
        if content:
            content = self._strip_html(content)

        return FeedEntry(
            url=url,
            title=title,
            published_date=published_date,
            content=content,
            source=source,
        )

    def _strip_html(self, text: str) -> str:
        """Remove HTML tags from text (basic implementation)."""
        import re

        # Remove script and style elements
        text = re.sub(r"<(script|style)[^>]*>.*?</\1>", "", text, flags=re.DOTALL | re.IGNORECASE)
        # Remove HTML tags
        text = re.sub(r"<[^>]+>", " ", text)
        # Decode common HTML entities
        text = text.replace("&nbsp;", " ")
        text = text.replace("&amp;", "&")
        text = text.replace("&lt;", "<")
        text = text.replace("&gt;", ">")
        text = text.replace("&quot;", '"')
        text = text.replace("&#39;", "'")
        # Normalize whitespace
        text = re.sub(r"\s+", " ", text)
        return text.strip()

    def _extract_source_name(self, url: str) -> str:
        """Extract a readable source name from URL."""
        from urllib.parse import urlparse

        parsed = urlparse(url)
        domain = parsed.netloc.lower()

        # Remove common prefixes
        for prefix in ["www.", "blog.", "blogs.", "rss.", "feeds."]:
            if domain.startswith(prefix):
                domain = domain[len(prefix) :]

        # Extract main domain name
        parts = domain.split(".")
        if len(parts) >= 2:
            return parts[-2].title()

        return domain.title()
