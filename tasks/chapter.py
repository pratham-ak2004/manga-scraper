from utils.playwright import PlaywrightManager
from utils.logger import get_logger
from .base import BaseEventTask
from bs4 import BeautifulSoup
from eventprocessor.celery_app import app
import os

lg = get_logger(__name__)

class ChapterScrapeTask(BaseEventTask):
    name = "tasks.chapter.scrape"
    image_section_selector = "section[hx-trigger=\"change from:[name='reading_style']\"]"
    image_format = "webp"
    
    def __init__(self, logger=lg):
        super().__init__(logger=logger)
    
    def validate(self, event: dict) -> bool:
        result = super().validate(event)
        url = event.get('url', '')
        id = event.get('id', '')
        name = event.get('name', '')
        chapter = event.get('chapter', '')
        base_path = event.get('base_path', '')
        return result and url and url.strip() != '' and id and id.strip() != '' and name and name.strip() != '' and chapter and chapter.strip() != '' and base_path and base_path.strip() != ''
    
    def scrape_chapter(self, url: str) -> str:
        """Scrape manga content with error handling"""
        try:
            with PlaywrightManager() as manager:
                with manager.new_page() as page:
                    page.goto(url, wait_until='networkidle')
                    return page.content()
        except Exception as e:
            raise e
        
    def get_images_from_page(self, soup: BeautifulSoup) -> list[dict]:
        try:
            image_tags = soup.select_one(self.image_section_selector).find_all('img')
        
            image_list = list()    
            for index, tag in enumerate(image_tags):
                image_list.append({
                    "link": tag.get('src'),
                    "alt_text": tag.get('alt', ''),
                    "index": index+1,
                    "file_path": None
                })
                
            return image_list
        except Exception as e:
            raise e
        
    def process(self, event: dict) -> dict:
        """
        Process the event data to scrape chapter information.
        """
        chapter_url = event.get('url')
        self.logger.info(f"Starting scrape for chapter URL: {chapter_url}")
        
        try:
            html_content = self.scrape_chapter(chapter_url)
            soup = BeautifulSoup(html_content, 'html.parser')
            image_list = self.get_images_from_page(soup)
            
            name = event.get('name')
            chapter = str(float(event.get('chapter')))
            base_path = event.get('base_path')
            path = os.path.join(base_path, name, "chapters", chapter)
            
            if not os.path.exists(path):
                self.logger.debug(f"Creating directory: {path}")
                os.makedirs(path)
                
            for index, image in enumerate(image_list):
                file_path = os.path.join(path, f"page_{image['index']}.{self.image_format}")
                image_list[index]['file_path'] = file_path
                
            self.logger.debug(f"Scraped {len(image_list)} images for chapter URL: {chapter_url}")
            
            return {
                'url': chapter_url,
                'status': 'success',
                'data': image_list,
                'payload': event
            }
        except Exception as e:
            self.logger.error(f"Scraping failed for {chapter_url}: {str(e)}")
            raise e