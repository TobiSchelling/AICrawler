# Database Schema Migrations

## Overview

AICrawler uses a versioned migration system to manage database schema changes. All schema modifications are tracked in the `schema_versions` table, allowing safe upgrades and downgrades.

## Current Schema Version

**Version 1** (Initial): Implemented with the article relevance feedback system

Includes:
- `articles` - Collected news articles
- `weekly_reports` - Weekly digest metadata
- `research_priorities` - User-defined research topics
- `article_priority_matches` - Article-to-priority associations
- `article_feedback` - User feedback on article relevance

## Checking Schema Status

### View Current Version
```bash
aicrawler schema version
# Output: Current schema version: 1
```

### View Full Migration Status
```bash
aicrawler schema status
# Shows current/latest versions and list of applied migrations
```

### View Database Status (includes schema)
```bash
aicrawler status
# Shows database statistics and current schema version
```

## How Migrations Work

### Automatic Application

When the application starts, it:
1. Creates the `schema_versions` table if needed
2. Checks the current schema version in the database
3. Applies all pending migrations in order
4. Records each applied migration with timestamp

### Manual Check

```python
from src.database import Database

db = Database()
status = db.get_migration_status()
print(f"Current: v{status['current_version']}, Latest: v{status['latest_version']}")
print(f"Pending: {status['pending_migrations']} migrations")
```

## Adding New Migrations

### Step 1: Create Migration Function

Add a new function to `src/database.py` following this pattern:

```python
def _migration_v2_add_article_ratings(conn: sqlite3.Connection) -> None:
    """Migration 2: Add user ratings to articles."""
    conn.executescript(
        """
        -- Add ratings table
        CREATE TABLE article_ratings (
            id INTEGER PRIMARY KEY,
            article_id INTEGER REFERENCES articles(id),
            rating INTEGER CHECK(rating BETWEEN 1 AND 5),
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );

        -- Add index for queries
        CREATE INDEX idx_ratings_article ON article_ratings(article_id);
        """
    )

    # Record the migration
    conn.execute(
        "INSERT INTO schema_versions (version, description) VALUES (?, ?)",
        (2, "Add article ratings table"),
    )
    conn.commit()
```

### Step 2: Register Migration

Add the migration to the `MIGRATIONS` dictionary:

```python
MIGRATIONS = {
    1: _migration_v1_initial_schema,
    2: _migration_v2_add_article_ratings,  # Add this line
}
```

### Step 3: Test Migration

```python
import tempfile
from src.database import Database

with tempfile.TemporaryDirectory() as tmpdir:
    db = Database(f"{tmpdir}/test.db")
    status = db.get_migration_status()
    assert status['current_version'] == 2
    assert status['pending_migrations'] == 0
```

## Best Practices

1. **Idempotent Migrations**: Use `CREATE TABLE IF NOT EXISTS` to handle re-runs
2. **Clear Descriptions**: Write meaningful `description` text for each migration
3. **Backward Compatibility**: Avoid breaking existing code
4. **Test Thoroughly**: Test migrations on a fresh database and with existing data
5. **Small Changes**: Keep migrations focused and single-purpose
6. **Add Data Migration**: If altering columns, include data transformation logic

## Example: Adding a Column

```python
def _migration_v3_add_archived_flag(conn: sqlite3.Connection) -> None:
    """Migration 3: Add archived flag to articles."""
    conn.executescript(
        """
        -- Add column with default value
        ALTER TABLE articles ADD COLUMN is_archived BOOLEAN DEFAULT 0;

        -- Add index
        CREATE INDEX idx_articles_archived ON articles(is_archived);
        """
    )

    # Populate existing data
    conn.execute("UPDATE articles SET is_archived = 0 WHERE is_archived IS NULL")

    # Record migration
    conn.execute(
        "INSERT INTO schema_versions (version, description) VALUES (?, ?)",
        (3, "Add archived flag to articles"),
    )
    conn.commit()
```

## Troubleshooting

### Migration Failed

If a migration fails:
1. Check the database is not corrupted with: `sqlite3 data/articles.db ".tables"`
2. Review error logs for specific SQL issues
3. Create a backup: `cp data/articles.db data/articles.db.backup`
4. Fix the migration and re-test

### Stuck at Wrong Version

If the schema version is recorded incorrectly:
```python
from src.database import Database
import sqlite3

db = Database()
with db._connect() as conn:
    # View recorded versions
    rows = conn.execute("SELECT * FROM schema_versions ORDER BY version").fetchall()
    for row in rows:
        print(f"v{row['version']}: {row['description']} at {row['applied_at']}")

    # Manually delete if needed (careful!)
    # conn.execute("DELETE FROM schema_versions WHERE version = ?", (bad_version,))
```

## Related Files

- `src/database.py` - Migration definitions and tracking
- `src/cli.py` - `schema` command group for CLI
- This file - Migration documentation
