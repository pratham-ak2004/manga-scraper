from utils.playwright import PlaywrightManager
from utils.logger import get_logger
from .base import BaseEventTask
from bs4 import BeautifulSoup

logger = get_logger(__name__)

class ChapterScrapeTask(BaseEventTask):
    name = "tasks.chapter.scrape"
    
    def __init__(self):
        super().__init__()
    
    def validate(self, event):
        result = super().validate(event)
        url = event.get('url', '')
        id = event.get('id', '')
        return result and url and url.strip() != '' and id and id.strip() != ''
    
    def scrape_chapter(self, url):
        """Scrape manga content with error handling"""
        try:
            with PlaywrightManager() as manager:
                with manager.new_page() as page:
                    page.goto(url, wait_until='networkidle')
                    return page.content()
        except Exception as e:
            logger.error(f"Failed to scrape {url}: {str(e)}")
            raise e
        
    def get_images_from_page(self, soup):
        try:
            image_tags = soup.select_one("section[hx-trigger=\"change from:[name='reading_style']\"]").find_all('img')
        
            image_list = list()    
            for index, tag in enumerate(image_tags):
                image_list.append({
                    "link": tag.get('src'),
                    "alt_text": tag.get('alt', ''),
                    "index": index
                })
                
            return image_list
        except Exception as e:
            logger.error(f"Failed to extract images: {str(e)}")
            raise e
        
    def process(self, event_data):
        """
        Process the event data to scrape chapter information.
        """
        chapter_url = event_data.get('url')
        try:
            html_content = self.scrape_chapter(chapter_url)
            soup = BeautifulSoup(html_content, 'html.parser')
            image_list = self.get_images_from_page(soup)
            # Further processing can be done here to extract chapter details
            return {
                'url': chapter_url,
                'status': 'success',
                'data': image_list,
                'payload': event_data
            }
        except Exception as e:
            logger.error(f"Scraping failed for {chapter_url}: {str(e)}")
            raise e