from ..page import PageDownloadTask
from utils import get_logger

logger = get_logger(__name__)

class PageDownloadPipelineTask(PageDownloadTask):
    name = "tasks.pipeline.page.download"