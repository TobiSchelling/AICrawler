"""RSS/Atom feed parser for AICrawler."""

import logging
from dataclasses import dataclass
from datetime import datetime
from time import mktime

import feedparser

logger = logging.getLogger(__name__)


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

    def parse_all(self) -> list[FeedEntry]:
        """Parse all configured feeds and return entries."""
        all_entries: list[FeedEntry] = []

        for feed_config in self.feeds:
            url = feed_config["url"]
            name = feed_config.get("name", self._extract_source_name(url))

            try:
                entries = self._parse_feed(url, name)
                all_entries.extend(entries)
                logger.info(f"Parsed {len(entries)} entries from {name}")
            except Exception as e:
                logger.error(f"Failed to parse feed {url}: {e}")

        return all_entries

    def _parse_feed(self, url: str, source_name: str) -> list[FeedEntry]:
        """Parse a single feed URL."""
        feed = feedparser.parse(url)

        if feed.bozo and feed.bozo_exception:
            logger.warning(f"Feed parsing warning for {url}: {feed.bozo_exception}")

        entries = []
        for entry in feed.entries:
            try:
                parsed = self._parse_entry(entry, source_name)
                if parsed:
                    entries.append(parsed)
            except Exception as e:
                logger.warning(f"Failed to parse entry: {e}")

        return entries

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
