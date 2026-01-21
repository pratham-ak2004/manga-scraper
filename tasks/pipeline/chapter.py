from ..chapter import ChapterScrapeTask
from utils import get_logger

lg = get_logger(__name__)

class ChapterScrapePipelineTask(ChapterScrapeTask):
    name = "tasks.pipeline.chapter.scrape"
    
    def __init__(self, logger=lg):
        super().__init__(logger=logger)