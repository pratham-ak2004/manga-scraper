from ..chapter import ChapterScrapeTask
from utils import get_logger

logger = get_logger(__name__)

class ChapterScrapePipelineTask(ChapterScrapeTask):
    name = "tasks.pipeline.chapter.scrape"