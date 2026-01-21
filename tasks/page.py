from .base import BaseEventTask
from utils.logger import get_logger
import requests, os
from PIL import Image
from io import BytesIO

lg = get_logger(__name__)

class PageDownloadTask(BaseEventTask):
    name = "tasks.page.download"
    image_format = "webp"
    compression_method = 4
    
    def __init__(self, logger=lg):
        super().__init__(logger=logger)
        
    def validate(self, event: dict) -> bool:
        result = super().validate(event)
        url = event.get('url', '')
        index = event.get('index', '')
        path = event.get('file_path', '')
        return result and url and url.strip() != '' and path and path.strip() != '' and index is not None
    
    def process_image(self, image_content: bytes, file_path: str) -> None:
        img = Image.open(BytesIO(image_content))
        if img.mode not in ('RGB', 'RGBA'):
            self.logger.debug(f"Converting image mode from {img.mode} to RGB for {file_path}")
            img = img.convert('RGB')
            
        img.save(file_path, format=self.image_format.upper(), quality=100, lossless=True, method=self.compression_method)
    
    def process(self, event: dict) -> dict:
        """
        Process the event data to fetch page information.
        """
        page_url = event.get('url')
        file_path = event.get('file_path')
        
        self.logger.info(f"Starting download for page URL: {page_url}")
        
        try:         
            response = requests.get(page_url, timeout=60)
            response.raise_for_status()
            
            self.process_image(response.content, file_path)
                
            self.logger.debug(f"Successfully fetched and saved page from {page_url} to {file_path}")
            
            return {
                'payload': event,
                "status": "success",
                "data": {
                    "response_code": response.status_code,
                    "file_path": file_path,
                    "file_size": os.path.getsize(file_path)
                }
            }
            
        except Exception as e:
            self.logger.error(f"Error fetching {page_url} : {str(e)}")
            raise e