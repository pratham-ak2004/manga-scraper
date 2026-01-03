from .base import BaseEventTask
from utils.logger import get_logger
from utils.playwright import PlaywrightManager
from bs4 import BeautifulSoup
import re

logger = get_logger(__name__)

class MangaScrapeTask(BaseEventTask):
    name = "tasks.manga.scrape"
    
    def __init__(self):
        super().__init__()
    
    def validate(self, event):
        result = super().validate(event)
        url = event.get('url', '')
        id = event.get('id', '')
        return result and url and url.strip() != '' and id and id.strip() != ''
    
    def extract_chapter_number(self, name):
        match = re.search(r'(\d+(?:\.\d+)?)\s*$', name)
        if match:
            return match.group(1)
        return None
    
    def extract_details(self, soup):
        try:
            details_section = soup.find("main").find_all("section")[1]
            
            name = details_section.find_next("h1").text.strip()
            picture = details_section.find("picture").find_next("img")['src']
            details_map = dict()
            
            for li in details_section.find("ul").find_all("li"):
                key = li.find("strong").text.strip().rstrip(':')
                values = list()
                for val in li.find_all("a"):
                    values.append({
                        "link": val['href'],
                        "name": val.text.strip() or "N/A"
                    })
                    
                details_map[key] = values
                
            return {
                "name": name,
                "picture": picture,
                "details": details_map
            }
        except Exception as e:
            logger.error(f"Error extracting details: {str(e)}")
            return {
                "name": "N/A",
                "picture": "N/A",
                "details": {}
            }
    
    def scrape_manga(self, url):
        """Scrape manga content with error handling"""
        try:
            with PlaywrightManager() as manager:
                with manager.new_page() as page:
                    page.goto(url, wait_until='networkidle')
                    button = page.locator('button[hx-target="#chapter-list"]')
                    with page.expect_response(lambda response: 'chapters' in response.url or response.request.resource_type == "fetch") as response_info:
                        button.click()
                    return page.content()
        except Exception as e:
            logger.error(f"Failed to scrape {url}: {str(e)}")
            raise e
    
    def process(self, event_data):
        """
        Process a manga scrape event
        
        Args:
            - **event_data** (dict) The event data containing manga info
            
        Returns:
            - **dict** Result of the scraping operation
        """
        manga_url = event_data.get('url')
        if not manga_url:
            logger.error("Manga URL missing in event data")
            return {'status': 'error', 'message': 'URL missing'}
        
        logger.info(f"Scraping manga from URL: {manga_url}")
        
        try:
            html_content = self.scrape_manga(manga_url)
            soup = BeautifulSoup(html_content, "html.parser")
            
            chapters = []
            for chapter in reversed(soup.find(id="chapter-list").find_all("a")):
                link = chapter['href']
                if not (link.startswith("https://weebcentral.com/chapters")):
                    continue

                name = chapter.find_all("span")[2].text.strip()
                number = self.extract_chapter_number(name)
                
                chapters.append({
                    "name": name,
                    "number": number,
                    "link": link
                })
                
            try:
                description = soup.find_all("ul")[2]
                if description.find("strong"):
                    description.find("strong").decompose()
                description = description.p.text.strip()
            except:
                description = "N/A"
                
            data = self.extract_details(soup)
            data["chapters"] = chapters
            data["description"] = description
            
            logger.info(f"Total chapters found: {len(chapters)}")
            
            return {
                'url': manga_url,
                'status': 'success',
                'data': data,
                'payload': event_data
            }
        except Exception as e:
            logger.error(f"Scraping failed for {manga_url}: {str(e)}")
            raise e