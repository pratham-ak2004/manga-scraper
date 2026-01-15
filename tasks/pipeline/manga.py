from ..manga import MangaScrapeTask
from utils import get_logger

logger = get_logger(__name__)

class MangaScrapePipelineTask(MangaScrapeTask):
    name = "tasks.pipeline.manga.scrape"