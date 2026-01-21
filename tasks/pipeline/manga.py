from ..manga import MangaScrapeTask
from utils import get_logger

lg = get_logger(__name__)

class MangaScrapePipelineTask(MangaScrapeTask):
    name = "tasks.pipeline.manga.scrape"
    
    def __init__(self, logger=lg):
        super().__init__(logger=logger)