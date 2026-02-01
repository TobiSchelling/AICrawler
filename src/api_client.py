"""NewsAPI client for AICrawler."""

import logging
import os
from dataclasses import dataclass
from datetime import datetime, timedelta

import httpx

logger = logging.getLogger(__name__)

NEWSAPI_BASE_URL = "https://newsapi.org/v2/everything"


@dataclass
class NewsArticle:
    """Represents an article from NewsAPI."""

    url: str
    title: str
    published_date: str | None
    content: str | None
    source: str


class NewsAPIClient:
    """Client for fetching articles from NewsAPI."""

    def __init__(
        self,
        api_key: str | None = None,
        api_key_env: str = "NEWSAPI_KEY",
    ):
        """
        Initialize NewsAPI client.

        Args:
            api_key: API key directly, or None to read from environment.
            api_key_env: Environment variable name for API key.
        """
        self.api_key = api_key or os.environ.get(api_key_env)
        if not self.api_key:
            logger.warning(
                f"NewsAPI key not found. Set {api_key_env} environment variable."
            )

    def is_configured(self) -> bool:
        """Check if API key is available."""
        return bool(self.api_key)

    def search(
        self,
        query: str,
        days_back: int = 7,
        page_size: int = 100,
        language: str = "en",
    ) -> list[NewsArticle]:
        """
        Search for articles matching query.

        Args:
            query: Search query string.
            days_back: Number of days to search back.
            page_size: Number of results per page (max 100).
            language: Language filter.

        Returns:
            List of NewsArticle objects.
        """
        if not self.api_key:
            logger.warning("NewsAPI not configured, skipping search")
            return []

        from_date = (datetime.now() - timedelta(days=days_back)).strftime("%Y-%m-%d")
        to_date = datetime.now().strftime("%Y-%m-%d")

        params = {
            "q": query,
            "from": from_date,
            "to": to_date,
            "language": language,
            "pageSize": min(page_size, 100),
            "sortBy": "relevancy",
        }

        headers = {"X-Api-Key": self.api_key}

        try:
            with httpx.Client(timeout=30.0) as client:
                response = client.get(NEWSAPI_BASE_URL, params=params, headers=headers)
                response.raise_for_status()
                data = response.json()

            if data.get("status") != "ok":
                logger.error(f"NewsAPI error: {data.get('message', 'Unknown error')}")
                return []

            articles = []
            for article in data.get("articles", []):
                parsed = self._parse_article(article)
                if parsed:
                    articles.append(parsed)

            logger.info(f"Fetched {len(articles)} articles from NewsAPI for query: {query}")
            return articles

        except httpx.HTTPStatusError as e:
            logger.error(f"NewsAPI HTTP error: {e.response.status_code}")
            return []
        except httpx.RequestError as e:
            logger.error(f"NewsAPI request error: {e}")
            return []
        except Exception as e:
            logger.error(f"NewsAPI unexpected error: {e}")
            return []

    def search_with_priorities(
        self,
        base_query: str,
        priorities: list[str],
        days_back: int = 7,
    ) -> list[NewsArticle]:
        """
        Search with base query and additional priority-based queries.

        Args:
            base_query: Main search query.
            priorities: List of priority titles/keywords to search for.
            days_back: Number of days to search back.

        Returns:
            Combined list of articles (deduplicated by URL).
        """
        all_articles: dict[str, NewsArticle] = {}

        # Base query
        for article in self.search(base_query, days_back=days_back):
            all_articles[article.url] = article

        # Priority-specific queries
        for priority in priorities:
            priority_query = f"{base_query} {priority}"
            for article in self.search(priority_query, days_back=days_back, page_size=50):
                if article.url not in all_articles:
                    all_articles[article.url] = article

        return list(all_articles.values())

    def _parse_article(self, data: dict) -> NewsArticle | None:
        """Parse NewsAPI article response into NewsArticle."""
        url = data.get("url")
        title = data.get("title")

        if not url or not title:
            return None

        # Skip removed articles
        if title == "[Removed]" or url == "https://removed.com":
            return None

        # Get published date
        published_date = None
        if pub := data.get("publishedAt"):
            try:
                published_date = datetime.fromisoformat(
                    pub.replace("Z", "+00:00")
                ).strftime("%Y-%m-%d")
            except ValueError:
                pass

        # Get content (NewsAPI provides truncated content)
        content = data.get("content") or data.get("description")
        if content:
            # NewsAPI truncates content with "[+N chars]" - note this in content
            content = content.strip()

        # Get source name
        source = "NewsAPI"
        if source_data := data.get("source"):
            source = source_data.get("name", "NewsAPI")

        return NewsArticle(
            url=url,
            title=title.strip(),
            published_date=published_date,
            content=content,
            source=source,
        )
