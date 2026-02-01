"""Per-storyline LLM narrative synthesis for AICrawler."""

import logging
from dataclasses import dataclass

from .database import Article, Database, Storyline, get_db
from .llm import LLMProvider, create_provider, parse_json_response

logger = logging.getLogger(__name__)

BRIEFLY_NOTED_LABEL = "Briefly Noted"

SYNTHESIS_PROMPT = """You are writing one section of a weekly AI news briefing for software practitioners.

This section covers a storyline about: {label}

Write a cohesive 2-3 paragraph narrative that weaves these articles together. Write as if you're a well-informed colleague explaining what happened this week. Be specific about tools, techniques, and outcomes. Avoid marketing language.

Articles in this storyline:
{articles}

Respond with ONLY this JSON:
{{
    "title": "A compelling 5-8 word section title",
    "narrative": "Your 2-3 paragraph narrative here. Use markdown for emphasis.",
    "source_references": [
        {{"title": "Article Title", "url": "https://...", "contribution": "What this article added to the story"}}
    ]
}}"""


@dataclass
class SynthesisResult:
    """Results from a synthesis run."""

    narratives_created: int
    errors: int


class StorylineSynthesizer:
    """Synthesizes narratives for each storyline using LLM."""

    def __init__(
        self,
        config: dict,
        db: Database | None = None,
        provider: LLMProvider | None = None,
    ):
        self.config = config
        self.db = db or get_db()
        self.provider = provider or create_provider(config)

    def synthesize_week(self, week_number: str) -> SynthesisResult:
        """Synthesize narratives for all storylines in a week."""
        if not self.provider:
            logger.error("No LLM provider available for synthesis")
            return SynthesisResult(narratives_created=0, errors=1)

        storylines = self.db.get_storylines_for_week(week_number)
        if not storylines:
            logger.info("No storylines to synthesize for %s", week_number)
            return SynthesisResult(narratives_created=0, errors=0)

        created = 0
        errors = 0

        for storyline in storylines:
            # Skip if narrative already exists (resumable)
            existing = self.db.get_narrative_for_storyline(storyline.id)
            if existing:
                logger.debug("Narrative already exists for storyline %d", storyline.id)
                created += 1
                continue

            articles = self.db.get_storyline_articles(storyline.id)
            if not articles:
                continue

            try:
                if storyline.label == BRIEFLY_NOTED_LABEL:
                    self._synthesize_briefly_noted(storyline, articles, week_number)
                else:
                    self._synthesize_storyline(storyline, articles, week_number)
                created += 1
            except Exception as e:
                logger.error("Error synthesizing storyline %d: %s", storyline.id, e)
                errors += 1

        logger.info(
            "Synthesis complete: %d narratives created, %d errors", created, errors
        )

        return SynthesisResult(narratives_created=created, errors=errors)

    def _synthesize_storyline(
        self,
        storyline: Storyline,
        articles: list[Article],
        week_number: str,
    ) -> None:
        """Synthesize a narrative for a regular storyline."""
        articles_text = self._format_articles(articles)

        prompt = SYNTHESIS_PROMPT.format(
            label=storyline.label,
            articles=articles_text,
        )

        response_text = self.provider.generate(prompt, max_tokens=1024)
        if not response_text:
            raise RuntimeError("LLM returned empty response")

        result = parse_json_response(response_text)

        if result:
            title = result.get("title", storyline.label)
            narrative = result.get("narrative", "")
            refs = result.get("source_references", [])
        else:
            # Use raw response as narrative
            title = storyline.label
            narrative = response_text.strip()
            refs = [{"title": a.title, "url": a.url} for a in articles]

        self.db.insert_storyline_narrative(
            storyline_id=storyline.id,
            week_number=week_number,
            title=title,
            narrative_text=narrative,
            source_references=refs,
        )

    def _synthesize_briefly_noted(
        self,
        storyline: Storyline,
        articles: list[Article],
        week_number: str,
    ) -> None:
        """Create a bullet-point summary for unclustered articles."""
        bullets = []
        refs = []

        for article in articles:
            triage = self.db.get_triage(article.id)
            if triage and triage.key_points:
                point = triage.key_points[0]
            else:
                point = article.title

            bullets.append(f"- **{article.title}** ({article.source}): {point}")
            refs.append({"title": article.title, "url": article.url})

        narrative = "\n".join(bullets)

        self.db.insert_storyline_narrative(
            storyline_id=storyline.id,
            week_number=week_number,
            title=BRIEFLY_NOTED_LABEL,
            narrative_text=narrative,
            source_references=refs,
        )

    def _format_articles(self, articles: list[Article]) -> str:
        """Format articles for the synthesis prompt."""
        parts = []
        for i, article in enumerate(articles, 1):
            triage = self.db.get_triage(article.id)
            key_points = ""
            if triage and triage.key_points:
                key_points = "\n  Key points: " + "; ".join(triage.key_points)

            content_preview = ""
            if article.content:
                content_preview = f"\n  Content: {article.content[:300]}..."

            parts.append(
                f"[{i}] {article.title}\n"
                f"  Source: {article.source}\n"
                f"  URL: {article.url}"
                f"{key_points}"
                f"{content_preview}"
            )

        return "\n\n".join(parts)
