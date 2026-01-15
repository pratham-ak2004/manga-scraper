from eventprocessor.celery_app import app
from .manga import MangaArchiveTask 

def bind_archve_tasks():
    app.register_task(MangaArchiveTask())
    
__all__ = ["bind_archve_tasks", "MangaArchiveTask"]