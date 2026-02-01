"""Shared test fixtures for AICrawler."""

import tempfile

import pytest

from src.database import Database, reset_db


@pytest.fixture
def temp_db():
    """Create a temporary database for testing."""
    reset_db()
    with tempfile.TemporaryDirectory() as tmpdir:
        db = Database(f"{tmpdir}/test.db")
        yield db
    reset_db()
