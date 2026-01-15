from .base import BaseEventTask
from .pipeline import bind_pipeline_tasks
from .archive import bind_archve_tasks
from eventprocessor.celery_app import app


bind_pipeline_tasks()
bind_archve_tasks()

__all__ = ["BaseEventTask"]