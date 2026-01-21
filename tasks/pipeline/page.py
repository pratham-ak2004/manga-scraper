from ..page import PageDownloadTask
from utils import get_logger

lg = get_logger(__name__)

class PageDownloadPipelineTask(PageDownloadTask):
    name = "tasks.pipeline.page.download"
    
    def __init__(self, logger=lg):
        super().__init__(logger)