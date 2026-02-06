"""Article collection orchestrator for AICrawler."""

import logging
from dataclasses import dataclass

from .api_client import NewsAPIClient
from .database import Database, get_db, get_today
from .feed_parser import FeedParser

logger = logging.getLogger(__name__)


@dataclass
class CollectionResult:
    """Results from a collection run."""

    total_found: int
    new_articles: int
    duplicates: int
    sources: dict[str, int]


class ArticleCollector:
    """Orchestrates article collection from RSS feeds and NewsAPI."""

    def __init__(self, config: dict, db: Database | None = None, days_back: int = 1):
        self.config = config
        self.db = db or get_db()
        self.days_back = days_back

        sources_config = config.get("sources", {})

        # Feed parser
        feeds = sources_config.get("feeds", [])
        self.feed_parser = FeedParser(feeds) if feeds else None

        # NewsAPI client
        api_config = sources_config.get("apis", {}).get("newsapi", {})
        if api_config.get("enabled", True):
            self.news_client = NewsAPIClient(
                api_key_env=api_config.get("api_key_env", "NEWSAPI_KEY")
            )
            self.news_query = api_config.get(
                "query", "artificial intelligence software development"
            )
        else:
            self.news_client = None
            self.news_query = ""

    def collect(self, period_id: str | None = None) -> CollectionResult:
        """Collect articles from all configured sources."""
        if period_id is None:
            period_id = get_today()
        total_found = 0
        new_articles = 0
        duplicates = 0
        sources: dict[str, int] = {}

        # Collect from RSS feeds
        if self.feed_parser:
            logger.info("Collecting from RSS feeds...")
            feed_entries = self.feed_parser.parse_all(days_back=self.days_back)
            total_found += len(feed_entries)

            for entry in feed_entries:
                result = self.db.insert_article(
                    url=entry.url,
                    title=entry.title,
                    source=entry.source,
                    published_date=entry.published_date,
                    content=entry.content,
                    period_id=period_id,
                )
                if result:
                    new_articles += 1
                    sources[entry.source] = sources.get(entry.source, 0) + 1
                else:
                    duplicates += 1

        # Collect from NewsAPI
        if self.news_client and self.news_client.is_configured():
            logger.info("Collecting from NewsAPI...")

            # Use research priorities for enhanced queries
            active_priorities = self.db.get_active_priorities()
            priorities = [p.title for p in active_priorities]

            if priorities:
                logger.info("Using %d active priorities for search", len(priorities))
                news_articles = self.news_client.search_with_priorities(
                    base_query=self.news_query,
                    priorities=priorities,
                    days_back=self.days_back,
                )
            else:
                news_articles = self.news_client.search(
                    self.news_query, days_back=self.days_back
                )

            total_found += len(news_articles)

            for article in news_articles:
                result = self.db.insert_article(
                    url=article.url,
                    title=article.title,
                    source=article.source,
                    published_date=article.published_date,
                    content=article.content,
                    period_id=period_id,
                )
                if result:
                    new_articles += 1
                    sources[article.source] = sources.get(article.source, 0) + 1
                else:
                    duplicates += 1

        logger.info(
            "Collection complete: %d found, %d new, %d duplicates",
            total_found,
            new_articles,
            duplicates,
        )

        return CollectionResult(
            total_found=total_found,
            new_articles=new_articles,
            duplicates=duplicates,
            sources=sources,
        )
