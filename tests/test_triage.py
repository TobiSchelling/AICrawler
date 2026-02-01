"""Tests for the triage module."""

from unittest.mock import MagicMock

import json

from src.triage import ArticleTriager


def test_triage_relevant_article(temp_db):
    """Test triaging an article that's relevant."""
    # Insert a test article
    aid = temp_db.insert_article(
        url="https://example.com/test",
        title="How We Use Claude for Code Review",
        content="A detailed experience report about using AI for code review...",
        week_number="2026-W05",
    )

    # Mock the LLM provider
    mock_provider = MagicMock()
    mock_provider.generate.return_value = json.dumps({
        "verdict": "relevant",
        "article_type": "experience_report",
        "key_points": ["AI code review improves quality", "Reduced review time by 40%"],
        "relevance_reason": "Direct experience report on AI in development",
        "practical_score": 4,
    })

    triager = ArticleTriager(config={}, db=temp_db, provider=mock_provider)
    result = triager.triage_articles("2026-W05")

    assert result.processed == 1
    assert result.relevant == 1
    assert result.skipped == 0

    triage = temp_db.get_triage(aid)
    assert triage.verdict == "relevant"
    assert triage.practical_score == 4


def test_triage_skip_article(temp_db):
    """Test triaging an article that should be skipped."""
    temp_db.insert_article(
        url="https://example.com/funding",
        title="AI Startup Raises $500M",
        content="Funding announcement for yet another AI company...",
        week_number="2026-W05",
    )

    mock_provider = MagicMock()
    mock_provider.generate.return_value = json.dumps({
        "verdict": "skip",
        "article_type": "announcement",
        "key_points": [],
        "relevance_reason": "Pure funding announcement, no technical substance",
        "practical_score": 0,
    })

    triager = ArticleTriager(config={}, db=temp_db, provider=mock_provider)
    result = triager.triage_articles("2026-W05")

    assert result.processed == 1
    assert result.relevant == 0
    assert result.skipped == 1


def test_triage_handles_unparseable_response(temp_db):
    """Test that unparseable LLM responses default to relevant."""
    temp_db.insert_article(
        url="https://example.com/test",
        title="Test Article",
        content="Some content",
        week_number="2026-W05",
    )

    mock_provider = MagicMock()
    mock_provider.generate.return_value = "This is not JSON at all"

    triager = ArticleTriager(config={}, db=temp_db, provider=mock_provider)
    result = triager.triage_articles("2026-W05")

    assert result.processed == 1
    assert result.relevant == 1  # defaults to relevant


def test_triage_skips_already_triaged(temp_db):
    """Test that already-triaged articles are skipped."""
    aid = temp_db.insert_article(
        url="https://example.com/test",
        title="Already Triaged",
        content="Content",
        week_number="2026-W05",
    )
    temp_db.insert_triage(article_id=aid, verdict="relevant", practical_score=3)

    mock_provider = MagicMock()
    triager = ArticleTriager(config={}, db=temp_db, provider=mock_provider)
    result = triager.triage_articles("2026-W05")

    assert result.processed == 0
    mock_provider.generate.assert_not_called()


def test_triage_no_provider(temp_db):
    """Test triage with no LLM provider available."""
    triager = ArticleTriager(config={}, db=temp_db)
    # Force provider to None (simulates no Ollama/OpenAI available)
    triager.provider = None

    temp_db.insert_article(
        url="https://example.com/test", title="Test", content="C", week_number="2026-W05"
    )

    result = triager.triage_articles("2026-W05")
    assert result.errors == 1
