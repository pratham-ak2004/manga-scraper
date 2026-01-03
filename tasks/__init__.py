from .base import BaseEventTask
from .manga import MangaScrapeTask
from .chapter import ChapterScrapeTask
from eventprocessor.celery_app import app

app.register_task(MangaScrapeTask())
app.register_task(ChapterScrapeTask())

__all__ = ["BaseEventTask", "manga", "chapter"]