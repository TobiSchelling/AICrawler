"""Per-article LLM triage for AICrawler."""

import logging
from dataclasses import dataclass

from .database import Article, Database, get_db
from .llm import LLMProvider, create_provider, parse_json_response

logger = logging.getLogger(__name__)

TRIAGE_PROMPT = """You are triaging AI news articles for a daily briefing aimed at people who build software.

Decide whether this article is RELEVANT or should be SKIPPED.

RELEVANT means: practical AI developments, experience reports from using AI tools, new techniques you can try, architecture patterns, tool releases, significant model updates, or insightful commentary on AI's impact on software development.

SKIP means: pure academic research papers, funding/investment announcements, marketing fluff, product launches with no technical substance, celebrity AI opinions, or AI doom/hype pieces with no practical content.

Research priorities to give extra weight:
{priorities}

Article Title: {title}
Source: {source}
Content:
{content}

Respond with ONLY this JSON:
{{
    "verdict": "relevant" or "skip",
    "article_type": "experience_report" | "tool_release" | "technique" | "architecture" | "model_update" | "commentary" | "tutorial" | "announcement" | "other",
    "key_points": ["point 1", "point 2", "point 3"],
    "relevance_reason": "One sentence explaining your verdict",
    "practical_score": 1-5
}}

practical_score: 5 = immediately actionable, 1 = tangentially related. Skip articles get 0."""


@dataclass
class TriageResult:
    """Results from a triage run."""

    processed: int
    relevant: int
    skipped: int
    errors: int


class ArticleTriager:
    """Triages articles using LLM for relevance assessment."""

    def __init__(
        self,
        config: dict,
        db: Database | None = None,
        provider: LLMProvider | None = None,
    ):
        self.config = config
        self.db = db or get_db()
        self.provider = provider or create_provider(config)

    def triage_articles(self, period_id: str | None = None) -> TriageResult:
        """Triage all untriaged articles."""
        if not self.provider:
            logger.error("No LLM provider available for triage")
            return TriageResult(processed=0, relevant=0, skipped=0, errors=1)

        articles = self.db.get_untriaged_articles(period_id)
        if not articles:
            logger.info("No articles pending triage")
            return TriageResult(processed=0, relevant=0, skipped=0, errors=0)

        priorities = self.db.get_active_priorities()
        priorities_text = self._format_priorities(priorities)

        processed = 0
        relevant = 0
        skipped = 0
        errors = 0

        for article in articles:
            try:
                result = self._triage_article(article, priorities_text)
                if result:
                    self.db.insert_triage(
                        article_id=article.id,
                        verdict=result["verdict"],
                        article_type=result.get("article_type"),
                        key_points=result.get("key_points", []),
                        relevance_reason=result.get("relevance_reason"),
                        practical_score=result.get("practical_score", 0),
                    )
                    processed += 1
                    if result["verdict"] == "relevant":
                        relevant += 1
                    else:
                        skipped += 1
                    logger.debug(
                        "Triaged [%s]: %s", result["verdict"], article.title
                    )
                else:
                    errors += 1
            except Exception as e:
                logger.error("Error triaging article %s: %s", article.id, e)
                errors += 1

        logger.info(
            "Triage complete: %d processed (%d relevant, %d skipped), %d errors",
            processed,
            relevant,
            skipped,
            errors,
        )

        return TriageResult(
            processed=processed,
            relevant=relevant,
            skipped=skipped,
            errors=errors,
        )

    def _triage_article(self, article: Article, priorities_text: str) -> dict | None:
        """Triage a single article via LLM."""
        content = article.content or article.title
        if len(content) > 4000:
            content = content[:4000] + "..."

        prompt = TRIAGE_PROMPT.format(
            title=article.title,
            source=article.source or "Unknown",
            content=content,
            priorities=priorities_text or "None defined",
        )

        response_text = self.provider.generate(prompt, max_tokens=512)
        if not response_text:
            return None

        result = parse_json_response(response_text)
        if not result:
            # Default to relevant if we can't parse the response
            return {
                "verdict": "relevant",
                "article_type": "other",
                "key_points": [],
                "relevance_reason": "LLM response could not be parsed",
                "practical_score": 2,
            }

        # Validate and normalize
        verdict = result.get("verdict", "relevant").lower()
        if verdict not in ("relevant", "skip"):
            verdict = "relevant"

        practical_score = result.get("practical_score", 2)
        if verdict == "skip":
            practical_score = 0
        else:
            practical_score = max(1, min(5, int(practical_score)))

        return {
            "verdict": verdict,
            "article_type": result.get("article_type", "other"),
            "key_points": result.get("key_points", [])[:5],
            "relevance_reason": result.get("relevance_reason", ""),
            "practical_score": practical_score,
        }

    def _format_priorities(self, priorities: list) -> str:
        if not priorities:
            return "None defined"
        lines = []
        for p in priorities:
            line = f"- {p.title}"
            if p.description:
                line += f": {p.description[:100]}"
            lines.append(line)
        return "\n".join(lines)
