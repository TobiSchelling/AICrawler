"""Compose full briefing from storyline narratives."""

import logging

from .database import Briefing, Database, get_db
from .llm import LLMProvider, create_provider, parse_json_response

logger = logging.getLogger(__name__)

COMPOSE_PROMPT = """You are writing the TL;DR for a daily AI news briefing aimed at software practitioners.

Here are today's storylines and their narratives:

{storylines}

Write a TL;DR section (3-5 bullet points) that captures the most important takeaways from ALL storylines. Each bullet should be one sentence that tells the reader what happened and why it matters.

Respond with ONLY this JSON:
{{
    "tldr_bullets": [
        "First key takeaway",
        "Second key takeaway",
        "Third key takeaway"
    ]
}}"""

BRIEFLY_NOTED_LABEL = "Briefly Noted"


class BriefingComposer:
    """Composes the final briefing from storyline narratives."""

    def __init__(
        self,
        config: dict,
        db: Database | None = None,
        provider: LLMProvider | None = None,
    ):
        self.config = config
        self.db = db or get_db()
        self.provider = provider or create_provider(config)

    def compose_briefing(self, period_id: str) -> Briefing:
        """Compose a complete briefing for a period."""
        narratives = self.db.get_narratives_for_period(period_id)
        storylines = self.db.get_storylines_for_period(period_id)

        if not narratives:
            logger.warning("No narratives found for %s", period_id)
            return self._store_empty_briefing(period_id)

        # Generate TL;DR via LLM
        tldr = self._generate_tldr(narratives)

        # Assemble the body markdown
        body = self._assemble_body(narratives, storylines)

        # Count articles
        article_count = sum(s.article_count for s in storylines)

        self.db.insert_briefing(
            period_id=period_id,
            tldr=tldr,
            body_markdown=body,
            storyline_count=len(storylines),
            article_count=article_count,
        )

        # Also store a run report entry
        self.db.insert_report(
            period_id=period_id,
            article_count=article_count,
            storyline_count=len(storylines),
        )

        result = self.db.get_briefing(period_id)
        logger.info("Briefing composed for %s: %d storylines", period_id, len(storylines))
        return result

    def _generate_tldr(self, narratives: list) -> str:
        """Generate TL;DR bullets from storyline narratives."""
        if not self.provider:
            return self._fallback_tldr(narratives)

        storylines_text = []
        for n in narratives:
            if n.title != BRIEFLY_NOTED_LABEL:
                storylines_text.append(f"## {n.title}\n{n.narrative_text}")

        prompt = COMPOSE_PROMPT.format(storylines="\n\n".join(storylines_text))

        response_text = self.provider.generate(prompt, max_tokens=512)
        if not response_text:
            return self._fallback_tldr(narratives)

        result = parse_json_response(response_text)
        if result and "tldr_bullets" in result:
            bullets = result["tldr_bullets"]
            return "\n".join(f"- {b}" for b in bullets)

        # Try using raw response
        return response_text.strip()

    def _fallback_tldr(self, narratives: list) -> str:
        """Generate a simple TL;DR without LLM."""
        bullets = []
        for n in narratives:
            if n.title != BRIEFLY_NOTED_LABEL:
                bullets.append(f"- {n.title}")
        return "\n".join(bullets) if bullets else "- No significant storylines today."

    def _assemble_body(self, narratives: list, storylines: list) -> str:
        """Assemble the full briefing body as markdown."""
        sections = []

        # Main storylines first, then Briefly Noted last
        main_narratives = [n for n in narratives if n.title != BRIEFLY_NOTED_LABEL]
        briefly_noted = [n for n in narratives if n.title == BRIEFLY_NOTED_LABEL]

        for narrative in main_narratives:
            section = f"## {narrative.title}\n\n{narrative.narrative_text}"

            # Add source references as a collapsible section
            if narrative.source_references:
                refs = []
                for ref in narrative.source_references:
                    title = ref.get("title", "")
                    url = ref.get("url", "")
                    contribution = ref.get("contribution", "")
                    line = f"- [{title}]({url})"
                    if contribution:
                        line += f" — {contribution}"
                    refs.append(line)
                section += f"\n\n**Sources:**\n{''.join(chr(10) + r for r in refs)}"

            sections.append(section)

        # Briefly Noted at the end
        for narrative in briefly_noted:
            section = f"## {narrative.title}\n\n{narrative.narrative_text}"
            sections.append(section)

        return "\n\n---\n\n".join(sections)

    def _store_empty_briefing(self, period_id: str) -> Briefing:
        """Store an empty briefing when there are no narratives."""
        self.db.insert_briefing(
            period_id=period_id,
            tldr="- No articles collected today.",
            body_markdown="No briefing content available for this period.",
            storyline_count=0,
            article_count=0,
        )
        return self.db.get_briefing(period_id)
