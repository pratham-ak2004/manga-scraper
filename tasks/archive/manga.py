from ..base import BaseEventTask
from utils.logger import get_logger
from zipfile import ZipFile
import os, zipfile

lg = get_logger(__name__)

class MangaArchiveTask(BaseEventTask):
    name = "tasks.archive.manga.cbz"
    
    def __init__(self, logger=lg):
        super().__init__(logger=logger)
        
    def validate(self, event: dict) -> bool:
        try:
            result = super().validate(event)
            
            manga = event.get("manga", None)
            archive = event.get("archive", None)
            pages = event.get("pages", None)
            
            file_path = archive.get("Filepath", None) if archive else None
    
            return result and manga is not None and archive is not None and pages is not None and file_path is not None
        except Exception as e:
            return False
    
    def process(self, event: dict) -> dict:
        try:
            manga = event.get("manga")
            archive = event.get("archive")
            pages = event.get("pages")
            
            file_path = archive.get("Filepath")
                        
            os.makedirs(os.path.dirname(file_path), exist_ok=True)
            
            self.logger.info(f"Creating manga archive at {file_path} with {len(pages)} pages.")
                        
            with zipfile.ZipFile(file_path, "w", zipfile.ZIP_DEFLATED) as cbz:
                for i, page in enumerate(pages):
                    page_path = page.get("Filepath")
                    arcname = f"{i+1:04d}-{os.path.splitext(page_path)[1]}"
                    cbz.write(page_path, arcname)
            
            self.logger.info(f"Successfully created manga archive at {file_path}")
            
            return {
                'payload': event,
                "status": "success",
                "data": {
                    "file_path": file_path,
                    "file_size": os.path.getsize(file_path)
                }
            }
        except Exception as e:
            self.logger.error(f"Error creating manga archive: {str(e)}")
            raise e