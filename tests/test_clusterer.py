"""Tests for the clustering module."""

from unittest.mock import MagicMock

import numpy as np

from src.clusterer import BRIEFLY_NOTED_LABEL, ArticleClusterer


def test_cluster_no_articles(temp_db):
    """Test clustering with no relevant articles."""
    clusterer = ArticleClusterer(db=temp_db)
    result = clusterer.cluster_articles("2026-02-06")

    assert result.storyline_count == 0
    assert result.article_count == 0


def test_cluster_single_article_goes_to_briefly_noted(temp_db):
    """Test that a single article goes to Briefly Noted."""
    aid = temp_db.insert_article(
        url="https://a.com", title="Solo Article", content="Content", period_id="2026-02-06"
    )
    temp_db.insert_triage(article_id=aid, verdict="relevant", practical_score=3)

    clusterer = ArticleClusterer(db=temp_db)
    result = clusterer.cluster_articles("2026-02-06")

    assert result.storyline_count == 1
    assert result.briefly_noted_count == 1

    storylines = temp_db.get_storylines_for_period("2026-02-06")
    assert storylines[0].label == BRIEFLY_NOTED_LABEL


def test_cluster_similar_articles_grouped(temp_db):
    """Test that similar articles are grouped into a storyline."""
    # Create articles about similar topic
    for i in range(3):
        aid = temp_db.insert_article(
            url=f"https://example.com/ai-testing-{i}",
            title=f"AI-Powered Testing Framework {i}: Revolution in QA",
            content=f"How AI is transforming software testing and QA processes part {i}",
            period_id="2026-02-06",
        )
        temp_db.insert_triage(article_id=aid, verdict="relevant", practical_score=4)

    # Create one unrelated article
    aid = temp_db.insert_article(
        url="https://example.com/crypto",
        title="New Cryptocurrency Market Analysis",
        content="Analysis of cryptocurrency markets and blockchain technology trends",
        period_id="2026-02-06",
    )
    temp_db.insert_triage(article_id=aid, verdict="relevant", practical_score=2)

    # Mock the model to return controllable embeddings
    mock_model = MagicMock()
    # 3 similar embeddings + 1 different
    mock_model.encode.return_value = np.array([
        [1.0, 0.0, 0.0],
        [0.95, 0.05, 0.0],
        [0.9, 0.1, 0.0],
        [0.0, 0.0, 1.0],
    ])

    clusterer = ArticleClusterer(db=temp_db, distance_threshold=1.0)
    clusterer._model = mock_model

    result = clusterer.cluster_articles("2026-02-06")

    assert result.article_count == 4
    # Should have at least one storyline (the 3 similar) + briefly noted (the outlier)
    assert result.storyline_count >= 1


def test_re_clustering_clears_old_data(temp_db):
    """Test that re-clustering clears previous storylines."""
    aid = temp_db.insert_article(
        url="https://a.com", title="A", content="Content", period_id="2026-02-06"
    )
    temp_db.insert_triage(article_id=aid, verdict="relevant", practical_score=3)

    # First clustering
    clusterer = ArticleClusterer(db=temp_db)
    clusterer.cluster_articles("2026-02-06")

    assert len(temp_db.get_storylines_for_period("2026-02-06")) == 1

    # Re-cluster
    clusterer.cluster_articles("2026-02-06")

    # Should still be exactly 1 (not 2)
    assert len(temp_db.get_storylines_for_period("2026-02-06")) == 1
