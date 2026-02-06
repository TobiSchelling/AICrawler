"""Fetch full article content for feeds with empty RSS entries."""

import logging
from dataclasses import dataclass

import httpx
import trafilatura

from .database import Database, get_db

logger = logging.getLogger(__name__)


@dataclass
class FetchResult:
    """Results from a content fetch run."""

    fetched: int
    already_had_content: int
    failed: int


class ContentFetcher:
    """Fetches full article text via HTTP + trafilatura extraction."""

    def __init__(self, db: Database | None = None, timeout: float = 15.0):
        self.db = db or get_db()
        self.timeout = timeout

    def fetch_missing_content(self, period_id: str | None = None) -> FetchResult:
        """Fetch content for articles that have empty content."""
        articles = self.db.get_articles_needing_fetch(period_id)

        if not articles:
            logger.info("No articles need content fetching")
            return FetchResult(fetched=0, already_had_content=0, failed=0)

        fetched = 0
        failed = 0

        for article in articles:
            try:
                content = self._fetch_article_content(article.url)
                if content:
                    self.db.update_article_content(article.id, content)
                    fetched += 1
                    logger.debug("Fetched content for: %s", article.title)
                else:
                    self.db.mark_article_fetch_attempted(article.id)
                    failed += 1
                    logger.debug("No extractable content from: %s", article.url)
            except Exception as e:
                self.db.mark_article_fetch_attempted(article.id)
                failed += 1
                logger.warning("Failed to fetch %s: %s", article.url, e)

        logger.info(
            "Content fetch complete: %d fetched, %d failed", fetched, failed
        )

        return FetchResult(fetched=fetched, already_had_content=0, failed=failed)

    def _fetch_article_content(self, url: str) -> str | None:
        """Fetch and extract article text from a URL."""
        try:
            with httpx.Client(
                timeout=self.timeout,
                follow_redirects=True,
                headers={"User-Agent": "AICrawler/0.2 (news aggregator)"},
            ) as client:
                response = client.get(url)
                response.raise_for_status()
                html = response.text
        except httpx.HTTPStatusError as e:
            logger.debug("HTTP %d for %s", e.response.status_code, url)
            return None
        except httpx.RequestError as e:
            logger.debug("Request error for %s: %s", url, e)
            return None

        # Extract readable text using trafilatura
        text = trafilatura.extract(html, include_comments=False, include_tables=False)

        if text and len(text.strip()) > 100:
            return text.strip()

        return None
