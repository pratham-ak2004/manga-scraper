from eventprocessor.celery_app import app
from .manga import MangaScrapePipelineTask
from .chapter import ChapterScrapePipelineTask
from .page import PageDownloadPipelineTask

def bind_pipeline_tasks():
    app.register_task(MangaScrapePipelineTask())
    app.register_task(ChapterScrapePipelineTask())
    app.register_task(PageDownloadPipelineTask())
    
__all__ = ["bind_pipeline_tasks", "MangaScrapePipelineTask", "ChapterScrapePipelineTask", "PageDownloadPipelineTask"]