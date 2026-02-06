"""Sentence-transformer embeddings + agglomerative clustering into storylines."""

import logging
from dataclasses import dataclass

import numpy as np
from scipy.cluster.hierarchy import fcluster, linkage
from sentence_transformers import SentenceTransformer

from .database import Article, Database, get_db

logger = logging.getLogger(__name__)

BRIEFLY_NOTED_LABEL = "Briefly Noted"
DEFAULT_MODEL = "all-MiniLM-L6-v2"
DEFAULT_DISTANCE_THRESHOLD = 1.2


@dataclass
class ClusterResult:
    """Results from a clustering run."""

    storyline_count: int
    article_count: int
    briefly_noted_count: int


class ArticleClusterer:
    """Clusters relevant articles into storylines using embeddings."""

    def __init__(
        self,
        db: Database | None = None,
        model_name: str = DEFAULT_MODEL,
        distance_threshold: float = DEFAULT_DISTANCE_THRESHOLD,
    ):
        self.db = db or get_db()
        self.model_name = model_name
        self.distance_threshold = distance_threshold
        self._model: SentenceTransformer | None = None

    @property
    def model(self) -> SentenceTransformer:
        if self._model is None:
            self._model = SentenceTransformer(self.model_name)
        return self._model

    def cluster_articles(self, period_id: str) -> ClusterResult:
        """Cluster relevant articles for a period into storylines."""
        articles = self.db.get_relevant_articles(period_id)

        if not articles:
            logger.info("No relevant articles to cluster for %s", period_id)
            return ClusterResult(storyline_count=0, article_count=0, briefly_noted_count=0)

        # Clear existing storylines for re-clustering
        self.db.clear_storylines_for_period(period_id)

        if len(articles) < 2:
            # Only one article — put it in Briefly Noted
            self.db.insert_storyline(
                period_id=period_id,
                label=BRIEFLY_NOTED_LABEL,
                article_ids=[a.id for a in articles],
            )
            return ClusterResult(
                storyline_count=1,
                article_count=len(articles),
                briefly_noted_count=len(articles),
            )

        # Build text representations for embedding
        texts = [self._article_text(a) for a in articles]

        # Generate embeddings
        logger.info("Generating embeddings for %d articles...", len(articles))
        embeddings = self.model.encode(texts, show_progress_bar=False)

        # Agglomerative clustering
        clusters = self._cluster_embeddings(embeddings)

        # Group articles by cluster
        cluster_groups: dict[int, list[Article]] = {}
        for article, cluster_id in zip(articles, clusters):
            cluster_groups.setdefault(cluster_id, []).append(article)

        # Separate real storylines from singletons
        storylines = []
        briefly_noted = []

        for cluster_id, group in cluster_groups.items():
            if len(group) >= 2:
                storylines.append(group)
            else:
                briefly_noted.extend(group)

        # Store storylines in DB
        for group in storylines:
            label = self._generate_label(group)
            self.db.insert_storyline(
                period_id=period_id,
                label=label,
                article_ids=[a.id for a in group],
            )

        # Store Briefly Noted group
        briefly_noted_count = 0
        if briefly_noted:
            self.db.insert_storyline(
                period_id=period_id,
                label=BRIEFLY_NOTED_LABEL,
                article_ids=[a.id for a in briefly_noted],
            )
            briefly_noted_count = len(briefly_noted)

        total_storylines = len(storylines) + (1 if briefly_noted else 0)

        logger.info(
            "Clustering complete: %d storylines + %d briefly noted from %d articles",
            len(storylines),
            briefly_noted_count,
            len(articles),
        )

        return ClusterResult(
            storyline_count=total_storylines,
            article_count=len(articles),
            briefly_noted_count=briefly_noted_count,
        )

    def _article_text(self, article: Article) -> str:
        """Build text representation for embedding."""
        parts = [article.title]

        # Add triage key points if available
        triage = self.db.get_triage(article.id)
        if triage and triage.key_points:
            parts.extend(triage.key_points)

        # Add truncated content
        if article.content:
            parts.append(article.content[:500])

        return " ".join(parts)

    def _cluster_embeddings(self, embeddings: np.ndarray) -> list[int]:
        """Perform agglomerative clustering on embeddings."""
        # Compute linkage matrix using Ward's method
        linkage_matrix = linkage(embeddings, method="ward", metric="euclidean")

        # Cut dendrogram at threshold
        clusters = fcluster(linkage_matrix, t=self.distance_threshold, criterion="distance")

        return clusters.tolist()

    def _generate_label(self, articles: list[Article]) -> str:
        """Generate a short label for a storyline from its articles."""
        # Use the most common significant words across titles
        from collections import Counter

        stop_words = {
            "the", "a", "an", "is", "are", "was", "were", "be", "been", "being",
            "have", "has", "had", "do", "does", "did", "will", "would", "could",
            "should", "may", "might", "can", "shall", "to", "of", "in", "for",
            "on", "with", "at", "by", "from", "as", "into", "through", "during",
            "before", "after", "above", "below", "and", "but", "or", "nor", "not",
            "so", "yet", "both", "either", "neither", "each", "every", "all",
            "any", "few", "more", "most", "other", "some", "such", "no", "only",
            "own", "same", "than", "too", "very", "just", "how", "what", "which",
            "who", "whom", "this", "that", "these", "those", "it", "its", "new",
            "about", "up", "out", "one", "two", "also", "like", "get", "use",
        }

        word_counts: Counter = Counter()
        for article in articles:
            words = article.title.lower().split()
            for word in words:
                word = word.strip(".,!?:;\"'()-[]")
                if len(word) > 2 and word not in stop_words:
                    word_counts[word] += 1

        # Take top 3 most common words
        top_words = [word for word, _ in word_counts.most_common(3)]

        if top_words:
            return " ".join(w.title() for w in top_words)

        # Fallback: use first article's title truncated
        return articles[0].title[:50]
