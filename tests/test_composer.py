"""Tests for the composer module."""

import json
from unittest.mock import MagicMock

from src.composer import BriefingComposer


def test_compose_briefing(temp_db):
    """Test composing a full briefing."""
    # Set up storyline with narrative
    a1 = temp_db.insert_article(
        url="https://a.com", title="A", content="C", period_id="2026-02-06"
    )
    a2 = temp_db.insert_article(
        url="https://b.com", title="B", content="C", period_id="2026-02-06"
    )
    sid = temp_db.insert_storyline(
        period_id="2026-02-06", label="AI Testing", article_ids=[a1, a2]
    )
    temp_db.insert_storyline_narrative(
        storyline_id=sid,
        period_id="2026-02-06",
        title="AI Transforms Testing",
        narrative_text="Today saw major changes in testing...",
        source_references=[{"title": "A", "url": "https://a.com"}],
    )

    mock_provider = MagicMock()
    mock_provider.generate.return_value = json.dumps({
        "tldr_bullets": [
            "AI testing tools gained significant traction",
            "New frameworks emerged for LLM-based QA",
        ]
    })

    composer = BriefingComposer(config={}, db=temp_db, provider=mock_provider)
    briefing = composer.compose_briefing("2026-02-06")

    assert briefing is not None
    assert briefing.period_id == "2026-02-06"
    assert briefing.storyline_count == 1
    assert briefing.article_count == 2
    assert "AI testing tools" in briefing.tldr
    assert "AI Transforms Testing" in briefing.body_markdown


def test_compose_empty_period(temp_db):
    """Test composing when no narratives exist."""
    mock_provider = MagicMock()
    composer = BriefingComposer(config={}, db=temp_db, provider=mock_provider)
    briefing = composer.compose_briefing("2026-02-06")

    assert briefing is not None
    assert briefing.article_count == 0
    mock_provider.generate.assert_not_called()


def test_compose_fallback_without_provider(temp_db):
    """Test composing with a provider that returns None falls back gracefully."""
    a1 = temp_db.insert_article(
        url="https://a.com", title="A", content="C", period_id="2026-02-06"
    )
    sid = temp_db.insert_storyline(
        period_id="2026-02-06", label="AI Testing", article_ids=[a1]
    )
    temp_db.insert_storyline_narrative(
        storyline_id=sid,
        period_id="2026-02-06",
        title="AI Testing Narrative",
        narrative_text="Content here.",
    )

    # Provider that always returns None (simulates LLM unavailable)
    mock_provider = MagicMock()
    mock_provider.generate.return_value = None

    composer = BriefingComposer(config={}, db=temp_db, provider=mock_provider)
    briefing = composer.compose_briefing("2026-02-06")

    assert briefing is not None
    # Fallback TL;DR uses storyline titles as bullets
    assert "AI Testing Narrative" in briefing.tldr
