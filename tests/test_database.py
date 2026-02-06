"""Tests for the database module."""

from src.database import format_period_display, get_today, make_period_id


def test_insert_article(temp_db):
    """Test inserting an article."""
    article_id = temp_db.insert_article(
        url="https://example.com/test",
        title="Test Article",
        source="Test Source",
        published_date="2026-01-27",
        content="Test content here",
        period_id="2026-02-06",
    )
    assert article_id is not None
    assert article_id > 0


def test_insert_duplicate_article(temp_db):
    """Test that duplicate URLs return None."""
    temp_db.insert_article(url="https://example.com/dup", title="First", period_id="2026-02-06")
    result = temp_db.insert_article(
        url="https://example.com/dup", title="Duplicate", period_id="2026-02-06"
    )
    assert result is None


def test_get_articles_for_period(temp_db):
    """Test fetching articles by period."""
    temp_db.insert_article(url="https://a.com", title="A", period_id="2026-02-06")
    temp_db.insert_article(url="https://b.com", title="B", period_id="2026-02-06")
    temp_db.insert_article(url="https://c.com", title="C", period_id="2026-02-05")

    articles = temp_db.get_articles_for_period("2026-02-06")
    assert len(articles) == 2


def test_articles_needing_fetch(temp_db):
    """Test fetching articles that need content."""
    temp_db.insert_article(url="https://a.com", title="No content", period_id="2026-02-06")
    temp_db.insert_article(
        url="https://b.com", title="Has content", content="Some text", period_id="2026-02-06"
    )

    needing = temp_db.get_articles_needing_fetch("2026-02-06")
    assert len(needing) == 1
    assert needing[0].title == "No content"


def test_update_article_content(temp_db):
    """Test updating article content after fetch."""
    aid = temp_db.insert_article(url="https://a.com", title="Test", period_id="2026-02-06")
    temp_db.update_article_content(aid, "Fetched content")

    article = temp_db.get_article_by_id(aid)
    assert article.content == "Fetched content"
    assert article.content_fetched is True


def test_triage_lifecycle(temp_db):
    """Test inserting and querying triage results."""
    aid = temp_db.insert_article(url="https://a.com", title="Test", period_id="2026-02-06")

    # Article should be untriaged
    untriaged = temp_db.get_untriaged_articles("2026-02-06")
    assert len(untriaged) == 1

    # Insert triage
    temp_db.insert_triage(
        article_id=aid,
        verdict="relevant",
        article_type="experience_report",
        key_points=["Point 1", "Point 2"],
        relevance_reason="Practical AI content",
        practical_score=4,
    )

    # Should no longer be untriaged
    untriaged = temp_db.get_untriaged_articles("2026-02-06")
    assert len(untriaged) == 0

    # Should appear in relevant articles
    relevant = temp_db.get_relevant_articles("2026-02-06")
    assert len(relevant) == 1

    # Check triage data
    triage = temp_db.get_triage(aid)
    assert triage.verdict == "relevant"
    assert triage.key_points == ["Point 1", "Point 2"]
    assert triage.practical_score == 4


def test_triage_stats(temp_db):
    """Test triage statistics."""
    a1 = temp_db.insert_article(url="https://a.com", title="A", period_id="2026-02-06")
    a2 = temp_db.insert_article(url="https://b.com", title="B", period_id="2026-02-06")

    temp_db.insert_triage(article_id=a1, verdict="relevant", practical_score=3)
    temp_db.insert_triage(article_id=a2, verdict="skip", practical_score=0)

    stats = temp_db.get_triage_stats("2026-02-06")
    assert stats["total"] == 2
    assert stats["relevant"] == 1
    assert stats["skipped"] == 1


def test_storyline_lifecycle(temp_db):
    """Test creating storylines with articles."""
    a1 = temp_db.insert_article(url="https://a.com", title="A", period_id="2026-02-06")
    a2 = temp_db.insert_article(url="https://b.com", title="B", period_id="2026-02-06")

    sid = temp_db.insert_storyline(
        period_id="2026-02-06",
        label="AI Testing",
        article_ids=[a1, a2],
    )
    assert sid is not None

    storylines = temp_db.get_storylines_for_period("2026-02-06")
    assert len(storylines) == 1
    assert storylines[0].label == "AI Testing"
    assert storylines[0].article_count == 2

    articles = temp_db.get_storyline_articles(sid)
    assert len(articles) == 2


def test_clear_storylines(temp_db):
    """Test clearing storylines for re-clustering."""
    a1 = temp_db.insert_article(url="https://a.com", title="A", period_id="2026-02-06")
    sid = temp_db.insert_storyline(
        period_id="2026-02-06", label="Test", article_ids=[a1]
    )
    temp_db.insert_storyline_narrative(
        storyline_id=sid, period_id="2026-02-06", title="T", narrative_text="N"
    )

    temp_db.clear_storylines_for_period("2026-02-06")

    assert len(temp_db.get_storylines_for_period("2026-02-06")) == 0
    assert len(temp_db.get_narratives_for_period("2026-02-06")) == 0


def test_briefing_lifecycle(temp_db):
    """Test creating and retrieving briefings."""
    temp_db.insert_briefing(
        period_id="2026-02-06",
        tldr="- Key point 1\n- Key point 2",
        body_markdown="## Section\nNarrative here.",
        storyline_count=3,
        article_count=15,
    )

    briefing = temp_db.get_briefing("2026-02-06")
    assert briefing is not None
    assert briefing.storyline_count == 3
    assert briefing.article_count == 15
    assert "Key point 1" in briefing.tldr

    all_briefings = temp_db.get_all_briefings()
    assert len(all_briefings) == 1


def test_priority_lifecycle(temp_db):
    """Test full priority CRUD."""
    pid = temp_db.insert_priority(title="AI Agents", description="Agent frameworks")
    assert pid is not None

    priority = temp_db.get_priority(pid)
    assert priority.title == "AI Agents"
    assert priority.is_active is True

    temp_db.toggle_priority(pid)
    priority = temp_db.get_priority(pid)
    assert priority.is_active is False

    temp_db.update_priority(pid, title="AI Agent Frameworks")
    priority = temp_db.get_priority(pid)
    assert priority.title == "AI Agent Frameworks"

    temp_db.delete_priority(pid)
    assert temp_db.get_priority(pid) is None


def test_get_stats(temp_db):
    """Test database statistics."""
    stats = temp_db.get_stats()
    assert stats["total_articles"] == 0
    assert stats["briefings"] == 0

    temp_db.insert_article(url="https://a.com", title="A", period_id="2026-02-06")
    temp_db.insert_priority(title="Test Priority")

    stats = temp_db.get_stats()
    assert stats["total_articles"] == 1
    assert stats["total_priorities"] == 1


def test_get_today():
    """Test today's date format."""
    today = get_today()
    assert today.startswith("20")
    assert len(today) == 10
    assert today.count("-") == 2


def test_format_period_display_single_day():
    """Test formatting a single-day period."""
    result = format_period_display("2026-02-06")
    assert "Feb" in result
    assert "2026" in result


def test_format_period_display_range():
    """Test formatting a date range period."""
    result = format_period_display("2026-02-01..2026-02-06")
    assert "Feb 01" in result
    assert "Feb 06" in result
    assert "-" in result


def test_make_period_id_single_day():
    """Test make_period_id for a single day."""
    assert make_period_id("2026-02-06", "2026-02-06") == "2026-02-06"


def test_make_period_id_range():
    """Test make_period_id for a date range."""
    assert make_period_id("2026-02-01", "2026-02-06") == "2026-02-01..2026-02-06"


def test_get_last_run_date(temp_db):
    """Test getting the last run date from reports."""
    assert temp_db.get_last_run_date() is None

    temp_db.insert_report(period_id="2026-02-05", article_count=10, storyline_count=3)
    assert temp_db.get_last_run_date() == "2026-02-05"


def test_get_last_run_date_range(temp_db):
    """Test getting the last run date from a range period."""
    temp_db.insert_report(period_id="2026-02-01..2026-02-05", article_count=10, storyline_count=3)
    assert temp_db.get_last_run_date() == "2026-02-05"
