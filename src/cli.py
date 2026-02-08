"""CLI interface for AICrawler."""

import importlib.resources
import logging
import os
import shutil
import signal
import socket
import subprocess
import sys
from datetime import date, timedelta
from pathlib import Path

import click
import yaml
from dotenv import load_dotenv

load_dotenv()

CONFIG_DIR = Path.home() / ".config" / "aicrawler"
DATA_DIR = Path.home() / ".local" / "share" / "aicrawler"


def setup_logging(verbose: bool = False) -> None:
    """Configure logging based on verbosity."""
    level = logging.DEBUG if verbose else logging.INFO
    logging.basicConfig(
        level=level,
        format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
        datefmt="%H:%M:%S",
    )


def resolve_config_path(explicit: str | None) -> Path:
    """Resolve the config file path.

    Priority: --config flag > ~/.config/aicrawler/config.yaml > ./config.yaml
    """
    if explicit:
        path = Path(explicit)
        if not path.exists():
            click.echo(f"Error: Config file not found: {explicit}", err=True)
            sys.exit(1)
        return path

    xdg_config = CONFIG_DIR / "config.yaml"
    if xdg_config.exists():
        return xdg_config

    cwd_config = Path("config.yaml")
    if cwd_config.exists():
        return cwd_config

    click.echo(
        "Error: No config file found. Searched:\n"
        f"  {xdg_config}\n"
        "  ./config.yaml\n"
        "\nRun 'aicrawler init' to create a default config.",
        err=True,
    )
    sys.exit(1)


def load_config(config_path: Path) -> dict:
    """Load configuration from YAML file."""
    with open(config_path) as f:
        return yaml.safe_load(f)


def get_data_dir(config: dict) -> str:
    """Get the data directory path from config or XDG default."""
    configured = config.get("output", {}).get("data_dir")
    if configured:
        return configured
    return str(DATA_DIR)


@click.group()
@click.option("-v", "--verbose", is_flag=True, help="Enable verbose output")
@click.option("-c", "--config", default=None, help="Path to config file")
@click.pass_context
def main(ctx: click.Context, verbose: bool, config: str | None) -> None:
    """AICrawler - Daily AI news briefings."""
    setup_logging(verbose)
    ctx.ensure_object(dict)
    ctx.obj["verbose"] = verbose

    if ctx.invoked_subcommand == "init":
        return

    config_path = resolve_config_path(config)
    ctx.obj["config_path"] = str(config_path)
    ctx.obj["config"] = load_config(config_path)
    ctx.obj["data_dir"] = get_data_dir(ctx.obj["config"])


@main.command()
def init() -> None:
    """Initialize configuration in ~/.config/aicrawler/."""
    target = CONFIG_DIR / "config.yaml"
    if target.exists():
        click.echo(f"Config already exists: {target}")
        return

    CONFIG_DIR.mkdir(parents=True, exist_ok=True)
    default_config = importlib.resources.files("src").joinpath("default_config.yaml")
    shutil.copy2(str(default_config), str(target))
    click.echo(f"Created config: {target}")
    click.echo("Edit it to configure feeds, API keys, and LLM provider.")


@main.command()
@click.pass_context
def collect(ctx: click.Context) -> None:
    """Collect articles from configured sources."""
    from .collector import ArticleCollector
    from .database import get_db, get_today

    config = ctx.obj["config"]
    get_db(data_dir=ctx.obj["data_dir"])
    period_id = get_today()

    click.echo("Collecting articles from sources...")

    collector = ArticleCollector(config)
    result = collector.collect(period_id=period_id)

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
@click.option("--days-back", type=int, default=None, help="Override lookback window (days)")
@click.pass_context
def run(ctx: click.Context, dry_run: bool, days_back: int | None) -> None:
    """Run the full pipeline: collect -> fetch -> triage -> cluster -> synthesize -> compose."""
    from .collector import ArticleCollector
    from .database import get_db, get_today, make_period_id

    config = ctx.obj["config"]
    db = get_db(data_dir=ctx.obj["data_dir"])
    today = get_today()

    # Explicit --days-back override
    if days_back is not None:
        if days_back == 1:
            period_id = today
        else:
            start = (date.fromisoformat(today) - timedelta(days=days_back - 1)).isoformat()
            period_id = make_period_id(start, today)
        click.echo(f"Collecting {days_back} day(s) of articles ({period_id}).")
    elif (last_run := db.get_last_run_date()) is None:
        click.echo("First run detected — collecting today's articles.")
        period_id = today
        days_back = 1
    else:
        last_date = date.fromisoformat(last_run)
        today_date = date.fromisoformat(today)
        missed_days = (today_date - last_date).days

        if missed_days <= 0:
            click.echo(f"Already ran today ({today}). Re-running pipeline.")
            period_id = today
            days_back = 1
        elif missed_days == 1:
            click.echo(f"Daily run for {today}.")
            period_id = today
            days_back = 1
        else:
            # Catch-up: missed days
            start_date = (last_date + timedelta(days=1)).isoformat()
            period_id = make_period_id(start_date, today)
            days_back = missed_days

            if missed_days > 5:
                click.echo(f"Last run was {missed_days} days ago ({last_run}).")
                if not click.confirm(
                    f"Catch up {missed_days} days ({period_id})? This will use more API calls"
                ):
                    click.echo("Aborted.")
                    return
            else:
                click.echo(f"Catching up {missed_days} days ({period_id}).")

    # Step 1: Collect
    click.echo("\nStep 1/6: Collecting articles...")
    if dry_run:
        articles = db.get_articles_for_period(period_id)
        click.echo(f"  [dry-run] {len(articles)} articles already in DB for {period_id}")
    else:
        collector = ArticleCollector(config, days_back=days_back)
        collect_result = collector.collect(period_id=period_id)
        click.echo(f"  Found {collect_result.new_articles} new articles")

    # Step 2: Fetch content
    click.echo("\nStep 2/6: Fetching article content...")
    if dry_run:
        needing_fetch = db.get_articles_needing_fetch(period_id)
        click.echo(f"  [dry-run] {len(needing_fetch)} articles need content fetching")
    else:
        from .content_fetcher import ContentFetcher

        fetcher = ContentFetcher(db=db)
        fetch_result = fetcher.fetch_missing_content(period_id)
        click.echo(f"  Fetched {fetch_result.fetched} articles, {fetch_result.failed} failed")

    # Step 3: Triage
    click.echo("\nStep 3/6: Triaging articles...")
    if dry_run:
        untriaged = db.get_untriaged_articles(period_id)
        click.echo(f"  [dry-run] {len(untriaged)} articles need triage")
    else:
        from .triage import ArticleTriager

        triager = ArticleTriager(config=config, db=db)
        triage_result = triager.triage_articles(period_id)
        click.echo(
            f"  Triaged {triage_result.processed} articles: "
            f"{triage_result.relevant} relevant, {triage_result.skipped} skipped"
        )

    # Step 4: Cluster into storylines
    click.echo("\nStep 4/6: Clustering into storylines...")
    if dry_run:
        relevant = db.get_relevant_articles(period_id)
        click.echo(f"  [dry-run] {len(relevant)} relevant articles to cluster")
    else:
        from .clusterer import ArticleClusterer

        clusterer = ArticleClusterer(db=db)
        cluster_result = clusterer.cluster_articles(period_id)
        click.echo(
            f"  Created {cluster_result.storyline_count} storylines "
            f"from {cluster_result.article_count} articles"
        )

    # Step 5: Synthesize storyline narratives
    click.echo("\nStep 5/6: Synthesizing narratives...")
    if dry_run:
        storylines = db.get_storylines_for_period(period_id)
        click.echo(f"  [dry-run] {len(storylines)} storylines need narratives")
    else:
        from .synthesizer import StorylineSynthesizer

        synthesizer = StorylineSynthesizer(config=config, db=db)
        synth_result = synthesizer.synthesize_period(period_id)
        click.echo(f"  Synthesized {synth_result.narratives_created} narratives")

    # Step 6: Compose briefing
    click.echo("\nStep 6/6: Composing briefing...")
    if dry_run:
        briefing = db.get_briefing(period_id)
        if briefing:
            click.echo(f"  [dry-run] Briefing already exists for {period_id}")
        else:
            click.echo(f"  [dry-run] Would compose briefing for {period_id}")
    else:
        from .composer import BriefingComposer

        composer = BriefingComposer(config=config, db=db)
        briefing = composer.compose_briefing(period_id)
        click.echo(f"  Briefing composed for {period_id}")
        click.echo(f"  {briefing.storyline_count} storylines, {briefing.article_count} articles")

    click.echo("\nPipeline complete! Run 'aicrawler serve' to view the briefing.")


def _stop_existing_server(port: int) -> bool:
    """Stop a previous aicrawler server on the given port, if any.

    Checks the port, identifies the PID via lsof, verifies it belongs to an
    aicrawler process, and only then sends SIGTERM.  Returns True if a
    process was stopped.
    """
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
        if s.connect_ex(("127.0.0.1", port)) != 0:
            return False  # port is free

    try:
        result = subprocess.run(
            ["lsof", "-ti", f"tcp:{port}"],
            capture_output=True,
            text=True,
        )
        if result.returncode != 0:
            return False
        pids = [p.strip() for p in result.stdout.strip().splitlines() if p.strip()]
    except (subprocess.SubprocessError, OSError):
        return False

    killed = False
    for pid_str in pids:
        try:
            pid = int(pid_str)
            ps_result = subprocess.run(
                ["ps", "-p", str(pid), "-o", "command="],
                capture_output=True,
                text=True,
            )
            command = ps_result.stdout.strip()
            if "aicrawler" not in command:
                continue
            os.kill(pid, signal.SIGTERM)
            killed = True
        except (ValueError, OSError):
            continue

    if killed:
        import time

        time.sleep(0.5)
    return killed


@main.command()
@click.option("--port", "-p", default=8000, help="Port to run server on")
@click.pass_context
def serve(ctx: click.Context, port: int) -> None:
    """Start the local web server."""
    from .server import serve as start_server

    if _stop_existing_server(port):
        click.echo(f"Stopped previous aicrawler server on port {port}")

    click.echo(f"Starting server at http://localhost:{port}")
    click.echo("Press Ctrl+C to stop")
    start_server(port=port)


@main.command()
@click.pass_context
def status(ctx: click.Context) -> None:
    """Show database and system status."""
    from .database import get_db, get_today

    db = get_db(data_dir=ctx.obj["data_dir"])
    stats = db.get_stats()
    today = get_today()

    click.echo(f"Today: {today}\n")
    click.echo("Articles:")
    click.echo(f"  Total collected: {stats['total_articles']}")
    click.echo(f"  Triaged: {stats['triaged_articles']}")
    click.echo(f"  Relevant: {stats['relevant_articles']}")
    click.echo("\nOutput:")
    click.echo(f"  Storylines: {stats['storylines']}")
    click.echo(f"  Briefings: {stats['briefings']}")
    click.echo(f"  Days with data: {stats['periods_with_articles']}")
    click.echo("\nResearch Priorities:")
    click.echo(f"  Total: {stats['total_priorities']}")
    click.echo(f"  Active: {stats['active_priorities']}")


# --- Priorities subcommand group ---


@main.group()
@click.pass_context
def priorities(ctx: click.Context) -> None:
    """Manage research priorities."""
    from .database import get_db

    get_db(data_dir=ctx.obj["data_dir"])


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
