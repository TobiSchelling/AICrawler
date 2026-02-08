"""Flask web server for AICrawler."""

import os
from pathlib import Path

import markdown
from flask import Flask, redirect, render_template, request, url_for

from .database import format_period_display, get_db

# Determine paths relative to this file
SRC_DIR = Path(__file__).parent

app = Flask(
    __name__,
    template_folder=str(SRC_DIR / "templates"),
    static_folder=str(SRC_DIR / "static"),
)

app.secret_key = os.environ.get("SECRET_KEY", "dev-key-change-in-production")

# Markdown renderer
md = markdown.Markdown(extensions=["fenced_code", "tables"])


def render_markdown(text: str) -> str:
    """Render markdown text to HTML."""
    md.reset()
    return md.convert(text)


# Register template filters
app.jinja_env.filters["markdown"] = render_markdown
app.jinja_env.filters["format_period"] = format_period_display


# --- Routes ---


@app.route("/")
def index():
    """Archive page listing all briefings."""
    db = get_db()
    briefings = db.get_all_briefings()

    return render_template("index.html", briefings=briefings)


@app.route("/briefing/<period_id>")
def briefing(period_id: str):
    """Display a briefing."""
    db = get_db()
    brief = db.get_briefing(period_id)

    if not brief:
        return render_template(
            "briefing.html",
            briefing=None,
            period_id=period_id,
        )

    return render_template(
        "briefing.html",
        briefing=brief,
        period_id=period_id,
    )


# --- Priorities ---


@app.route("/priorities")
def priorities():
    """List and manage research priorities."""
    db = get_db()
    items = db.get_all_priorities()

    return render_template("priorities.html", priorities=items)


@app.route("/priorities/add", methods=["POST"])
def add_priority():
    """Add a new research priority."""
    db = get_db()

    title = request.form.get("title", "").strip()
    description = request.form.get("description", "").strip()

    if title:
        db.insert_priority(title=title, description=description)

    return redirect(url_for("priorities"))


@app.route("/priorities/<int:priority_id>/edit", methods=["POST"])
def edit_priority(priority_id: int):
    """Update an existing priority."""
    db = get_db()

    title = request.form.get("title", "").strip()
    description = request.form.get("description", "").strip()

    if title:
        db.update_priority(priority_id, title=title, description=description)

    return redirect(url_for("priorities"))


@app.route("/priorities/<int:priority_id>/toggle", methods=["POST"])
def toggle_priority(priority_id: int):
    """Toggle a priority's active state."""
    db = get_db()
    db.toggle_priority(priority_id)
    return redirect(url_for("priorities"))


@app.route("/priorities/<int:priority_id>/delete", methods=["POST"])
def delete_priority(priority_id: int):
    """Delete a priority."""
    db = get_db()
    db.delete_priority(priority_id)
    return redirect(url_for("priorities"))


# --- Server entry point ---


def serve(port: int = 8000, debug: bool = False) -> None:
    """Start the Flask development server."""
    app.run(
        host="127.0.0.1",
        port=port,
        debug=debug,
    )


if __name__ == "__main__":
    serve(debug=True)
