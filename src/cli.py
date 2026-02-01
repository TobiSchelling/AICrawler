"""CLI interface for AICrawler."""

import logging
import sys
from pathlib import Path

import click
import yaml
from dotenv import load_dotenv

load_dotenv()


def setup_logging(verbose: bool = False) -> None:
    """Configure logging based on verbosity."""
    level = logging.DEBUG if verbose else logging.INFO
    logging.basicConfig(
        level=level,
        format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
        datefmt="%H:%M:%S",
    )


def load_config(config_path: str = "config.yaml") -> dict:
    """Load configuration from YAML file."""
    path = Path(config_path)
    if not path.exists():
        click.echo(f"Error: Config file not found: {config_path}", err=True)
        sys.exit(1)

    with open(path) as f:
        return yaml.safe_load(f)


@click.group()
@click.option("-v", "--verbose", is_flag=True, help="Enable verbose output")
@click.option("-c", "--config", default="config.yaml", help="Path to config file")
@click.pass_context
def main(ctx: click.Context, verbose: bool, config: str) -> None:
    """AICrawler - Weekly AI news briefings."""
    setup_logging(verbose)
    ctx.ensure_object(dict)
    ctx.obj["verbose"] = verbose
    ctx.obj["config_path"] = config
    ctx.obj["config"] = load_config(config)


@main.command()
@click.pass_context
def collect(ctx: click.Context) -> None:
    """Collect articles from configured sources."""
    from .collector import ArticleCollector

    config = ctx.obj["config"]
    click.echo("Collecting articles from sources...")

    collector = ArticleCollector(config)
    result = collector.collect()

    click.echo("\nCollection complete:")
    click.echo(f"  Total found: {result.total_found}")
    click.echo(f"  New articles: {result.new_articles}")
    click.echo(f"  Duplicates skipped: {result.duplicates}")

    if result.sources:
        click.echo("\nArticles by source:")
        for source, count in sorted(result.sources.items(), key=lambda x: -x[1]):
            click.echo(f"  {source}: {count}")


@main.command()
@click.option("--dry-run", is_flag=True, help="Show what would be done without executing")
@click.pass_context
def run(ctx: click.Context, dry_run: bool) -> None:
    """Run the full pipeline: collect -> fetch -> triage -> cluster -> synthesize -> compose."""
    from .collector import ArticleCollector
    from .database import get_current_week, get_db

    config = ctx.obj["config"]
    db = get_db()
    week_number = get_current_week()

    # Step 1: Collect
    click.echo("Step 1/6: Collecting articles...")
    if dry_run:
        articles = db.get_articles_for_week(week_number)
        click.echo(f"  [dry-run] {len(articles)} articles already in DB for {week_number}")
    else:
        collector = ArticleCollector(config)
        collect_result = collector.collect()
        click.echo(f"  Found {collect_result.new_articles} new articles")

    # Step 2: Fetch content
    click.echo("\nStep 2/6: Fetching article content...")
    if dry_run:
        needing_fetch = db.get_articles_needing_fetch(week_number)
        click.echo(f"  [dry-run] {len(needing_fetch)} articles need content fetching")
    else:
        from .content_fetcher import ContentFetcher

        fetcher = ContentFetcher(db=db)
        fetch_result = fetcher.fetch_missing_content(week_number)
        click.echo(f"  Fetched {fetch_result.fetched} articles, {fetch_result.failed} failed")

    # Step 3: Triage
    click.echo("\nStep 3/6: Triaging articles...")
    if dry_run:
        untriaged = db.get_untriaged_articles(week_number)
        click.echo(f"  [dry-run] {len(untriaged)} articles need triage")
    else:
        from .triage import ArticleTriager

        triager = ArticleTriager(config=config, db=db)
        triage_result = triager.triage_articles(week_number)
        click.echo(
            f"  Triaged {triage_result.processed} articles: "
            f"{triage_result.relevant} relevant, {triage_result.skipped} skipped"
        )

    # Step 4: Cluster into storylines
    click.echo("\nStep 4/6: Clustering into storylines...")
    if dry_run:
        relevant = db.get_relevant_articles(week_number)
        click.echo(f"  [dry-run] {len(relevant)} relevant articles to cluster")
    else:
        from .clusterer import ArticleClusterer

        clusterer = ArticleClusterer(db=db)
        cluster_result = clusterer.cluster_articles(week_number)
        click.echo(
            f"  Created {cluster_result.storyline_count} storylines "
            f"from {cluster_result.article_count} articles"
        )

    # Step 5: Synthesize storyline narratives
    click.echo("\nStep 5/6: Synthesizing narratives...")
    if dry_run:
        storylines = db.get_storylines_for_week(week_number)
        click.echo(f"  [dry-run] {len(storylines)} storylines need narratives")
    else:
        from .synthesizer import StorylineSynthesizer

        synthesizer = StorylineSynthesizer(config=config, db=db)
        synth_result = synthesizer.synthesize_week(week_number)
        click.echo(f"  Synthesized {synth_result.narratives_created} narratives")

    # Step 6: Compose weekly briefing
    click.echo("\nStep 6/6: Composing weekly briefing...")
    if dry_run:
        briefing = db.get_briefing(week_number)
        if briefing:
            click.echo(f"  [dry-run] Briefing already exists for {week_number}")
        else:
            click.echo(f"  [dry-run] Would compose briefing for {week_number}")
    else:
        from .composer import BriefingComposer

        composer = BriefingComposer(config=config, db=db)
        briefing = composer.compose_briefing(week_number)
        click.echo(f"  Briefing composed for {week_number}")
        click.echo(f"  {briefing.storyline_count} storylines, {briefing.article_count} articles")

    click.echo("\nPipeline complete! Run 'aicrawler serve' to view the briefing.")


@main.command()
@click.option("--port", "-p", default=8000, help="Port to run server on")
@click.pass_context
def serve(ctx: click.Context, port: int) -> None:
    """Start the local web server."""
    from .server import serve as start_server

    click.echo(f"Starting server at http://localhost:{port}")
    click.echo("Press Ctrl+C to stop")
    start_server(port=port)


@main.command()
@click.pass_context
def status(ctx: click.Context) -> None:
    """Show database and system status."""
    from .database import get_current_week, get_db

    db = get_db()
    stats = db.get_stats()
    week = get_current_week()

    click.echo(f"Current week: {week}\n")
    click.echo("Articles:")
    click.echo(f"  Total collected: {stats['total_articles']}")
    click.echo(f"  Triaged: {stats['triaged_articles']}")
    click.echo(f"  Relevant: {stats['relevant_articles']}")
    click.echo("\nOutput:")
    click.echo(f"  Storylines: {stats['storylines']}")
    click.echo(f"  Briefings: {stats['briefings']}")
    click.echo(f"  Weeks with data: {stats['weeks_with_articles']}")
    click.echo("\nResearch Priorities:")
    click.echo(f"  Total: {stats['total_priorities']}")
    click.echo(f"  Active: {stats['active_priorities']}")


# --- Priorities subcommand group ---


@main.group()
def priorities() -> None:
    """Manage research priorities."""
    pass


@priorities.command("list")
def priorities_list() -> None:
    """List all research priorities."""
    from .database import get_db

    db = get_db()
    items = db.get_all_priorities()

    if not items:
        click.echo("No priorities defined. Add one with: aicrawler priorities add")
        return

    click.echo("Research Priorities:\n")
    for p in items:
        status_icon = "*" if p.is_active else " "
        click.echo(f"  [{p.id}] {status_icon} {p.title}")
        if p.description:
            desc = p.description[:60] + "..." if len(p.description) > 60 else p.description
            click.echo(f"        {desc}")


@priorities.command("add")
@click.argument("title")
@click.argument("description", required=False, default="")
def priorities_add(title: str, description: str) -> None:
    """Add a new research priority."""
    from .database import get_db

    db = get_db()
    priority_id = db.insert_priority(title=title, description=description)
    click.echo(f"Added priority [{priority_id}]: {title}")


@priorities.command("remove")
@click.argument("priority_id", type=int)
def priorities_remove(priority_id: int) -> None:
    """Remove a research priority."""
    from .database import get_db

    db = get_db()
    priority = db.get_priority(priority_id)

    if not priority:
        click.echo(f"Error: Priority {priority_id} not found", err=True)
        sys.exit(1)

    db.delete_priority(priority_id)
    click.echo(f"Removed priority [{priority_id}]: {priority.title}")


@priorities.command("toggle")
@click.argument("priority_id", type=int)
def priorities_toggle(priority_id: int) -> None:
    """Toggle a priority's active state."""
    from .database import get_db

    db = get_db()
    priority = db.get_priority(priority_id)

    if not priority:
        click.echo(f"Error: Priority {priority_id} not found", err=True)
        sys.exit(1)

    db.toggle_priority(priority_id)
    new_state = "disabled" if priority.is_active else "enabled"
    click.echo(f"Priority [{priority_id}] {priority.title}: {new_state}")


if __name__ == "__main__":
    main()
