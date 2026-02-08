"""SQLite database operations for AICrawler."""

import json
import logging
import shutil
import sqlite3
from contextlib import contextmanager
from dataclasses import dataclass
from datetime import date, datetime
from pathlib import Path

logger = logging.getLogger(__name__)

# --- Dataclasses ---


@dataclass
class Article:
    """Represents a collected article."""

    id: int | None
    url: str
    title: str
    source: str | None
    published_date: str | None
    content: str | None
    content_fetched: bool
    period_id: str | None
    collected_at: str | None

    @classmethod
    def from_row(cls, row: sqlite3.Row) -> "Article":
        return cls(
            id=row["id"],
            url=row["url"],
            title=row["title"],
            source=row["source"],
            published_date=row["published_date"],
            content=row["content"],
            content_fetched=bool(row["content_fetched"]),
            period_id=row["period_id"],
            collected_at=row["collected_at"],
        )


@dataclass
class ArticleTriage:
    """Triage result for an article."""

    article_id: int
    verdict: str  # "relevant" or "skip"
    article_type: str | None
    key_points: list[str]
    relevance_reason: str | None
    practical_score: int
    triaged_at: str | None

    @classmethod
    def from_row(cls, row: sqlite3.Row) -> "ArticleTriage":
        key_points = json.loads(row["key_points"]) if row["key_points"] else []
        return cls(
            article_id=row["article_id"],
            verdict=row["verdict"],
            article_type=row["article_type"],
            key_points=key_points,
            relevance_reason=row["relevance_reason"],
            practical_score=row["practical_score"],
            triaged_at=row["triaged_at"],
        )


@dataclass
class Storyline:
    """A cluster of related articles forming a storyline."""

    id: int | None
    period_id: str
    label: str
    article_count: int
    created_at: str | None

    @classmethod
    def from_row(cls, row: sqlite3.Row) -> "Storyline":
        return cls(
            id=row["id"],
            period_id=row["period_id"],
            label=row["label"],
            article_count=row["article_count"],
            created_at=row["created_at"],
        )


@dataclass
class StorylineNarrative:
    """LLM-generated narrative for a storyline."""

    id: int | None
    storyline_id: int
    period_id: str
    title: str
    narrative_text: str
    source_references: list[dict]
    generated_at: str | None

    @classmethod
    def from_row(cls, row: sqlite3.Row) -> "StorylineNarrative":
        refs = json.loads(row["source_references"]) if row["source_references"] else []
        return cls(
            id=row["id"],
            storyline_id=row["storyline_id"],
            period_id=row["period_id"],
            title=row["title"],
            narrative_text=row["narrative_text"],
            source_references=refs,
            generated_at=row["generated_at"],
        )


@dataclass
class Briefing:
    """A complete briefing for a period."""

    id: int | None
    period_id: str
    tldr: str
    body_markdown: str
    storyline_count: int
    article_count: int
    generated_at: str | None

    @classmethod
    def from_row(cls, row: sqlite3.Row) -> "Briefing":
        return cls(
            id=row["id"],
            period_id=row["period_id"],
            tldr=row["tldr"],
            body_markdown=row["body_markdown"],
            storyline_count=row["storyline_count"],
            article_count=row["article_count"],
            generated_at=row["generated_at"],
        )


@dataclass
class ResearchPriority:
    """User-defined research priority."""

    id: int | None
    title: str
    description: str | None
    keywords: list[str]
    is_active: bool
    created_at: str | None
    updated_at: str | None

    @classmethod
    def from_row(cls, row: sqlite3.Row) -> "ResearchPriority":
        keywords = json.loads(row["keywords"]) if row["keywords"] else []
        return cls(
            id=row["id"],
            title=row["title"],
            description=row["description"],
            keywords=keywords,
            is_active=bool(row["is_active"]),
            created_at=row["created_at"],
            updated_at=row["updated_at"],
        )


@dataclass
class RunReport:
    """Metadata about a pipeline run."""

    id: int | None
    period_id: str
    generated_at: str | None
    article_count: int
    storyline_count: int

    @classmethod
    def from_row(cls, row: sqlite3.Row) -> "RunReport":
        return cls(
            id=row["id"],
            period_id=row["period_id"],
            generated_at=row["generated_at"],
            article_count=row["article_count"],
            storyline_count=row["storyline_count"],
        )


# --- Schema ---

SCHEMA_SQL = """
CREATE TABLE IF NOT EXISTS articles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    url TEXT UNIQUE NOT NULL,
    title TEXT NOT NULL,
    source TEXT,
    published_date TEXT,
    content TEXT,
    content_fetched INTEGER DEFAULT 0,
    period_id TEXT,
    collected_at TEXT DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS article_triage (
    article_id INTEGER PRIMARY KEY REFERENCES articles(id),
    verdict TEXT NOT NULL,
    article_type TEXT,
    key_points TEXT,
    relevance_reason TEXT,
    practical_score INTEGER DEFAULT 0,
    triaged_at TEXT DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS storylines (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    period_id TEXT NOT NULL,
    label TEXT NOT NULL,
    article_count INTEGER DEFAULT 0,
    created_at TEXT DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS storyline_articles (
    storyline_id INTEGER NOT NULL REFERENCES storylines(id),
    article_id INTEGER NOT NULL REFERENCES articles(id),
    PRIMARY KEY (storyline_id, article_id)
);

CREATE TABLE IF NOT EXISTS storyline_narratives (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    storyline_id INTEGER NOT NULL REFERENCES storylines(id),
    period_id TEXT NOT NULL,
    title TEXT NOT NULL,
    narrative_text TEXT NOT NULL,
    source_references TEXT,
    generated_at TEXT DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS briefings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    period_id TEXT UNIQUE NOT NULL,
    tldr TEXT NOT NULL,
    body_markdown TEXT NOT NULL,
    storyline_count INTEGER DEFAULT 0,
    article_count INTEGER DEFAULT 0,
    generated_at TEXT DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS research_priorities (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    description TEXT,
    keywords TEXT,
    is_active INTEGER DEFAULT 1,
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS run_reports (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    period_id TEXT UNIQUE NOT NULL,
    generated_at TEXT DEFAULT (datetime('now')),
    article_count INTEGER DEFAULT 0,
    storyline_count INTEGER DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_articles_period ON articles(period_id);
CREATE INDEX IF NOT EXISTS idx_articles_url ON articles(url);
CREATE INDEX IF NOT EXISTS idx_storylines_period ON storylines(period_id);
CREATE INDEX IF NOT EXISTS idx_storyline_narratives_period ON storyline_narratives(period_id);
CREATE INDEX IF NOT EXISTS idx_briefings_period ON briefings(period_id);
"""


# --- Database class ---


class Database:
    """SQLite database operations."""

    def __init__(self, db_path: str = "data/aicrawler.db"):
        self.db_path = db_path
        Path(db_path).parent.mkdir(parents=True, exist_ok=True)
        self._init_schema()

    def _init_schema(self) -> None:
        with self.connection() as conn:
            conn.executescript(SCHEMA_SQL)

    @contextmanager
    def connection(self):
        conn = sqlite3.connect(self.db_path)
        conn.row_factory = sqlite3.Row
        conn.execute("PRAGMA journal_mode=WAL")
        conn.execute("PRAGMA foreign_keys=ON")
        try:
            yield conn
            conn.commit()
        except Exception:
            conn.rollback()
            raise
        finally:
            conn.close()

    # --- Articles ---

    def insert_article(
        self,
        url: str,
        title: str,
        source: str | None = None,
        published_date: str | None = None,
        content: str | None = None,
        period_id: str | None = None,
    ) -> int | None:
        """Insert an article. Returns ID on success, None if duplicate."""
        with self.connection() as conn:
            try:
                cursor = conn.execute(
                    """INSERT INTO articles (url, title, source, published_date, content, period_id)
                    VALUES (?, ?, ?, ?, ?, ?)""",
                    (url, title, source, published_date, content, period_id),
                )
                return cursor.lastrowid
            except sqlite3.IntegrityError:
                return None

    def get_articles_for_period(self, period_id: str) -> list[Article]:
        with self.connection() as conn:
            rows = conn.execute(
                "SELECT * FROM articles WHERE period_id = ? ORDER BY collected_at DESC",
                (period_id,),
            ).fetchall()
            return [Article.from_row(r) for r in rows]

    def get_articles_needing_fetch(self, period_id: str | None = None) -> list[Article]:
        """Get articles with empty content that haven't been fetched yet."""
        query = """SELECT * FROM articles
            WHERE (content IS NULL OR content = '') AND content_fetched = 0"""
        params: list = []
        if period_id:
            query += " AND period_id = ?"
            params.append(period_id)
        query += " ORDER BY collected_at DESC"

        with self.connection() as conn:
            rows = conn.execute(query, params).fetchall()
            return [Article.from_row(r) for r in rows]

    def update_article_content(self, article_id: int, content: str | None) -> None:
        """Update article content after fetching."""
        with self.connection() as conn:
            conn.execute(
                "UPDATE articles SET content = ?, content_fetched = 1 WHERE id = ?",
                (content, article_id),
            )

    def mark_article_fetch_attempted(self, article_id: int) -> None:
        """Mark that we tried to fetch content (even if it failed)."""
        with self.connection() as conn:
            conn.execute(
                "UPDATE articles SET content_fetched = 1 WHERE id = ?",
                (article_id,),
            )

    def get_untriaged_articles(self, period_id: str | None = None) -> list[Article]:
        """Get articles that haven't been triaged yet."""
        query = """SELECT a.* FROM articles a
            LEFT JOIN article_triage t ON a.id = t.article_id
            WHERE t.article_id IS NULL"""
        params: list = []
        if period_id:
            query += " AND a.period_id = ?"
            params.append(period_id)
        query += " ORDER BY a.collected_at DESC"

        with self.connection() as conn:
            rows = conn.execute(query, params).fetchall()
            return [Article.from_row(r) for r in rows]

    def get_relevant_articles(self, period_id: str) -> list[Article]:
        """Get articles triaged as relevant for a given period."""
        with self.connection() as conn:
            rows = conn.execute(
                """SELECT a.* FROM articles a
                JOIN article_triage t ON a.id = t.article_id
                WHERE a.period_id = ? AND t.verdict = 'relevant'
                ORDER BY t.practical_score DESC""",
                (period_id,),
            ).fetchall()
            return [Article.from_row(r) for r in rows]

    def get_article_by_id(self, article_id: int) -> Article | None:
        with self.connection() as conn:
            row = conn.execute(
                "SELECT * FROM articles WHERE id = ?", (article_id,)
            ).fetchone()
            return Article.from_row(row) if row else None

    # --- Triage ---

    def insert_triage(
        self,
        article_id: int,
        verdict: str,
        article_type: str | None = None,
        key_points: list[str] | None = None,
        relevance_reason: str | None = None,
        practical_score: int = 0,
    ) -> None:
        with self.connection() as conn:
            conn.execute(
                """INSERT OR REPLACE INTO article_triage
                (article_id, verdict, article_type, key_points, relevance_reason, practical_score)
                VALUES (?, ?, ?, ?, ?, ?)""",
                (
                    article_id,
                    verdict,
                    article_type,
                    json.dumps(key_points) if key_points else None,
                    relevance_reason,
                    practical_score,
                ),
            )

    def get_triage(self, article_id: int) -> ArticleTriage | None:
        with self.connection() as conn:
            row = conn.execute(
                "SELECT * FROM article_triage WHERE article_id = ?", (article_id,)
            ).fetchone()
            return ArticleTriage.from_row(row) if row else None

    def get_triage_stats(self, period_id: str) -> dict:
        with self.connection() as conn:
            row = conn.execute(
                """SELECT
                    COUNT(*) as total,
                    SUM(CASE WHEN verdict = 'relevant' THEN 1 ELSE 0 END) as relevant,
                    SUM(CASE WHEN verdict = 'skip' THEN 1 ELSE 0 END) as skipped
                FROM article_triage t
                JOIN articles a ON a.id = t.article_id
                WHERE a.period_id = ?""",
                (period_id,),
            ).fetchone()
            return {
                "total": row["total"] or 0,
                "relevant": row["relevant"] or 0,
                "skipped": row["skipped"] or 0,
            }

    # --- Storylines ---

    def insert_storyline(
        self,
        period_id: str,
        label: str,
        article_ids: list[int],
    ) -> int:
        with self.connection() as conn:
            cursor = conn.execute(
                """INSERT INTO storylines (period_id, label, article_count)
                VALUES (?, ?, ?)""",
                (period_id, label, len(article_ids)),
            )
            storyline_id = cursor.lastrowid
            for aid in article_ids:
                conn.execute(
                    "INSERT INTO storyline_articles (storyline_id, article_id) VALUES (?, ?)",
                    (storyline_id, aid),
                )
            return storyline_id

    def get_storylines_for_period(self, period_id: str) -> list[Storyline]:
        with self.connection() as conn:
            rows = conn.execute(
                "SELECT * FROM storylines WHERE period_id = ? ORDER BY article_count DESC",
                (period_id,),
            ).fetchall()
            return [Storyline.from_row(r) for r in rows]

    def get_storyline_article_ids(self, storyline_id: int) -> list[int]:
        with self.connection() as conn:
            rows = conn.execute(
                "SELECT article_id FROM storyline_articles WHERE storyline_id = ?",
                (storyline_id,),
            ).fetchall()
            return [r["article_id"] for r in rows]

    def get_storyline_articles(self, storyline_id: int) -> list[Article]:
        with self.connection() as conn:
            rows = conn.execute(
                """SELECT a.* FROM articles a
                JOIN storyline_articles sa ON a.id = sa.article_id
                WHERE sa.storyline_id = ?""",
                (storyline_id,),
            ).fetchall()
            return [Article.from_row(r) for r in rows]

    def clear_storylines_for_period(self, period_id: str) -> None:
        """Remove existing storylines for a period (for re-clustering)."""
        with self.connection() as conn:
            storyline_ids = conn.execute(
                "SELECT id FROM storylines WHERE period_id = ?", (period_id,)
            ).fetchall()
            for row in storyline_ids:
                conn.execute(
                    "DELETE FROM storyline_articles WHERE storyline_id = ?", (row["id"],)
                )
                conn.execute(
                    "DELETE FROM storyline_narratives WHERE storyline_id = ?", (row["id"],)
                )
            conn.execute("DELETE FROM storylines WHERE period_id = ?", (period_id,))

    # --- Storyline Narratives ---

    def insert_storyline_narrative(
        self,
        storyline_id: int,
        period_id: str,
        title: str,
        narrative_text: str,
        source_references: list[dict] | None = None,
    ) -> int:
        with self.connection() as conn:
            cursor = conn.execute(
                """INSERT INTO storyline_narratives
                (storyline_id, period_id, title, narrative_text, source_references)
                VALUES (?, ?, ?, ?, ?)""",
                (
                    storyline_id,
                    period_id,
                    title,
                    narrative_text,
                    json.dumps(source_references) if source_references else None,
                ),
            )
            return cursor.lastrowid

    def get_narratives_for_period(self, period_id: str) -> list[StorylineNarrative]:
        with self.connection() as conn:
            rows = conn.execute(
                """SELECT sn.* FROM storyline_narratives sn
                JOIN storylines s ON s.id = sn.storyline_id
                WHERE sn.period_id = ?
                ORDER BY s.article_count DESC""",
                (period_id,),
            ).fetchall()
            return [StorylineNarrative.from_row(r) for r in rows]

    def get_narrative_for_storyline(self, storyline_id: int) -> StorylineNarrative | None:
        with self.connection() as conn:
            row = conn.execute(
                "SELECT * FROM storyline_narratives WHERE storyline_id = ?",
                (storyline_id,),
            ).fetchone()
            return StorylineNarrative.from_row(row) if row else None

    # --- Briefings ---

    def insert_briefing(
        self,
        period_id: str,
        tldr: str,
        body_markdown: str,
        storyline_count: int,
        article_count: int,
    ) -> int:
        with self.connection() as conn:
            cursor = conn.execute(
                """INSERT OR REPLACE INTO briefings
                (period_id, tldr, body_markdown, storyline_count, article_count)
                VALUES (?, ?, ?, ?, ?)""",
                (period_id, tldr, body_markdown, storyline_count, article_count),
            )
            return cursor.lastrowid

    def get_briefing(self, period_id: str) -> Briefing | None:
        with self.connection() as conn:
            row = conn.execute(
                "SELECT * FROM briefings WHERE period_id = ?",
                (period_id,),
            ).fetchone()
            return Briefing.from_row(row) if row else None

    def get_all_briefings(self) -> list[Briefing]:
        with self.connection() as conn:
            rows = conn.execute(
                "SELECT * FROM briefings ORDER BY period_id DESC"
            ).fetchall()
            return [Briefing.from_row(r) for r in rows]

    # --- Run Reports (metadata) ---

    def insert_report(
        self, period_id: str, article_count: int, storyline_count: int
    ) -> int:
        with self.connection() as conn:
            cursor = conn.execute(
                """INSERT OR REPLACE INTO run_reports
                (period_id, article_count, storyline_count)
                VALUES (?, ?, ?)""",
                (period_id, article_count, storyline_count),
            )
            return cursor.lastrowid

    def get_last_run_date(self) -> str | None:
        """Get the end date from the most recent run report's period_id.

        Returns the date string (YYYY-MM-DD) or None if no runs exist.
        For range periods (YYYY-MM-DD..YYYY-MM-DD), returns the end date.
        """
        with self.connection() as conn:
            row = conn.execute(
                "SELECT period_id FROM run_reports ORDER BY period_id DESC LIMIT 1"
            ).fetchone()
            if not row:
                return None
            period_id = row["period_id"]
            # Range format: "YYYY-MM-DD..YYYY-MM-DD" — return end date
            if ".." in period_id:
                return period_id.split("..")[1]
            return period_id

    # --- Research Priorities ---

    def insert_priority(
        self,
        title: str,
        description: str = "",
        keywords: list[str] | None = None,
    ) -> int:
        with self.connection() as conn:
            cursor = conn.execute(
                """INSERT INTO research_priorities (title, description, keywords)
                VALUES (?, ?, ?)""",
                (title, description, json.dumps(keywords) if keywords else None),
            )
            return cursor.lastrowid

    def get_all_priorities(self) -> list[ResearchPriority]:
        with self.connection() as conn:
            rows = conn.execute(
                "SELECT * FROM research_priorities ORDER BY created_at DESC"
            ).fetchall()
            return [ResearchPriority.from_row(r) for r in rows]

    def get_active_priorities(self) -> list[ResearchPriority]:
        with self.connection() as conn:
            rows = conn.execute(
                "SELECT * FROM research_priorities WHERE is_active = 1 ORDER BY created_at DESC"
            ).fetchall()
            return [ResearchPriority.from_row(r) for r in rows]

    def get_priority(self, priority_id: int) -> ResearchPriority | None:
        with self.connection() as conn:
            row = conn.execute(
                "SELECT * FROM research_priorities WHERE id = ?", (priority_id,)
            ).fetchone()
            return ResearchPriority.from_row(row) if row else None

    def update_priority(
        self,
        priority_id: int,
        title: str | None = None,
        description: str | None = None,
        keywords: list[str] | None = None,
    ) -> None:
        updates = []
        params: list = []
        if title is not None:
            updates.append("title = ?")
            params.append(title)
        if description is not None:
            updates.append("description = ?")
            params.append(description)
        if keywords is not None:
            updates.append("keywords = ?")
            params.append(json.dumps(keywords))
        if not updates:
            return
        updates.append("updated_at = datetime('now')")
        params.append(priority_id)

        with self.connection() as conn:
            conn.execute(
                f"UPDATE research_priorities SET {', '.join(updates)} WHERE id = ?",
                params,
            )

    def toggle_priority(self, priority_id: int) -> None:
        with self.connection() as conn:
            conn.execute(
                """UPDATE research_priorities
                SET is_active = NOT is_active, updated_at = datetime('now')
                WHERE id = ?""",
                (priority_id,),
            )

    def delete_priority(self, priority_id: int) -> None:
        with self.connection() as conn:
            conn.execute(
                "DELETE FROM research_priorities WHERE id = ?", (priority_id,)
            )

    # --- Stats ---

    def get_stats(self) -> dict:
        with self.connection() as conn:
            total = conn.execute("SELECT COUNT(*) as c FROM articles").fetchone()["c"]
            triaged = conn.execute(
                "SELECT COUNT(*) as c FROM article_triage"
            ).fetchone()["c"]
            relevant = conn.execute(
                "SELECT COUNT(*) as c FROM article_triage WHERE verdict = 'relevant'"
            ).fetchone()["c"]
            periods = conn.execute(
                "SELECT COUNT(DISTINCT period_id) as c FROM articles"
            ).fetchone()["c"]
            briefings = conn.execute(
                "SELECT COUNT(*) as c FROM briefings"
            ).fetchone()["c"]
            storylines = conn.execute(
                "SELECT COUNT(*) as c FROM storylines"
            ).fetchone()["c"]
            priorities = conn.execute(
                "SELECT COUNT(*) as c FROM research_priorities"
            ).fetchone()["c"]
            active_priorities = conn.execute(
                "SELECT COUNT(*) as c FROM research_priorities WHERE is_active = 1"
            ).fetchone()["c"]

            return {
                "total_articles": total,
                "triaged_articles": triaged,
                "relevant_articles": relevant,
                "periods_with_articles": periods,
                "briefings": briefings,
                "storylines": storylines,
                "total_priorities": priorities,
                "active_priorities": active_priorities,
            }


# --- Utility functions ---


def get_today() -> str:
    """Get today's date as YYYY-MM-DD string."""
    return date.today().isoformat()


def make_period_id(start: str, end: str) -> str:
    """Create a period_id from start and end dates.

    If start == end, returns just the date (e.g., "2026-02-06").
    Otherwise returns a range (e.g., "2026-02-01..2026-02-06").
    """
    if start == end:
        return start
    return f"{start}..{end}"


def format_period_display(period_id: str) -> str:
    """Format a period_id for human-readable display.

    Single day: "Feb 06, 2026"
    Range: "Feb 01 - Feb 06, 2026"
    """
    try:
        if ".." in period_id:
            start_str, end_str = period_id.split("..")
            start = date.fromisoformat(start_str)
            end = date.fromisoformat(end_str)
            return f"{start.strftime('%b %d')} - {end.strftime('%b %d, %Y')}"
        d = date.fromisoformat(period_id)
        return d.strftime("%b %d, %Y")
    except (ValueError, IndexError):
        return period_id


def backup_database(db_path: str) -> str | None:
    """Back up existing database before migration."""
    path = Path(db_path)
    if not path.exists():
        return None
    backup_path = str(path.with_suffix(f".backup-{datetime.now():%Y%m%d-%H%M%S}.db"))
    shutil.copy2(db_path, backup_path)
    logger.info("Database backed up to %s", backup_path)
    return backup_path


# --- Singleton ---

_db_instance: Database | None = None


def get_db(db_path: str | None = None, data_dir: str | None = None) -> Database:
    """Get or create the singleton Database instance.

    Priority: explicit db_path > data_dir/aicrawler.db > ~/.local/share/aicrawler/aicrawler.db
    """
    global _db_instance
    if _db_instance is None:
        if db_path:
            path = db_path
        elif data_dir:
            path = str(Path(data_dir) / "aicrawler.db")
        else:
            default_dir = Path.home() / ".local" / "share" / "aicrawler"
            path = str(default_dir / "aicrawler.db")
        _db_instance = Database(path)
    return _db_instance


def reset_db() -> None:
    """Reset the singleton (for testing)."""
    global _db_instance
    _db_instance = None
