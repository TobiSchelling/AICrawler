"""Tests for the synthesizer module."""

import json
from unittest.mock import MagicMock

from src.synthesizer import StorylineSynthesizer, BRIEFLY_NOTED_LABEL


def test_synthesize_storyline(temp_db):
    """Test synthesizing a narrative for a storyline."""
    a1 = temp_db.insert_article(
        url="https://a.com", title="AI Testing Part 1", content="Content 1",
        week_number="2026-W05",
    )
    a2 = temp_db.insert_article(
        url="https://b.com", title="AI Testing Part 2", content="Content 2",
        week_number="2026-W05",
    )
    temp_db.insert_triage(article_id=a1, verdict="relevant", key_points=["Point 1"])
    temp_db.insert_triage(article_id=a2, verdict="relevant", key_points=["Point 2"])

    sid = temp_db.insert_storyline(
        week_number="2026-W05", label="AI Testing", article_ids=[a1, a2]
    )

    mock_provider = MagicMock()
    mock_provider.generate.return_value = json.dumps({
        "title": "AI Transforms Software Testing",
        "narrative": "This week saw significant progress in AI-powered testing...",
        "source_references": [
            {"title": "AI Testing Part 1", "url": "https://a.com", "contribution": "Foundation"},
            {"title": "AI Testing Part 2", "url": "https://b.com", "contribution": "Extensions"},
        ],
    })

    synthesizer = StorylineSynthesizer(config={}, db=temp_db, provider=mock_provider)
    result = synthesizer.synthesize_week("2026-W05")

    assert result.narratives_created == 1
    assert result.errors == 0

    narrative = temp_db.get_narrative_for_storyline(sid)
    assert narrative is not None
    assert narrative.title == "AI Transforms Software Testing"


def test_synthesize_briefly_noted(temp_db):
    """Test that Briefly Noted gets bullet-point treatment."""
    a1 = temp_db.insert_article(
        url="https://a.com", title="Random Article", content="Content",
        source="Source A", week_number="2026-W05",
    )
    temp_db.insert_triage(article_id=a1, verdict="relevant", key_points=["A key point"])

    sid = temp_db.insert_storyline(
        week_number="2026-W05", label=BRIEFLY_NOTED_LABEL, article_ids=[a1]
    )

    mock_provider = MagicMock()  # Should NOT be called for briefly noted
    synthesizer = StorylineSynthesizer(config={}, db=temp_db, provider=mock_provider)
    result = synthesizer.synthesize_week("2026-W05")

    assert result.narratives_created == 1
    mock_provider.generate.assert_not_called()  # No LLM call for briefly noted

    narrative = temp_db.get_narrative_for_storyline(sid)
    assert narrative.title == BRIEFLY_NOTED_LABEL
    assert "Random Article" in narrative.narrative_text


def test_synthesize_skips_existing(temp_db):
    """Test that existing narratives are not re-generated."""
    a1 = temp_db.insert_article(
        url="https://a.com", title="A", content="C", week_number="2026-W05"
    )
    sid = temp_db.insert_storyline(
        week_number="2026-W05", label="Test", article_ids=[a1]
    )
    temp_db.insert_storyline_narrative(
        storyline_id=sid, week_number="2026-W05", title="Existing", narrative_text="Already done"
    )

    mock_provider = MagicMock()
    synthesizer = StorylineSynthesizer(config={}, db=temp_db, provider=mock_provider)
    result = synthesizer.synthesize_week("2026-W05")

    assert result.narratives_created == 1  # counted as created (existing)
    mock_provider.generate.assert_not_called()
