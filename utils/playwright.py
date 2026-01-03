from playwright.sync_api import sync_playwright, Browser, Page
from contextlib import contextmanager
from typing import Optional

class PlaywrightManager:
    _instance: Optional['PlaywrightManager'] = None
    _playwright = None
    _browser: Optional[Browser] = None
    
    def __new__(cls):
        if cls._instance is None:
            cls._instance = super().__new__(cls)
        return cls._instance
    
    def __enter__(self):
        if self._playwright is None:
            self._playwright = sync_playwright().start()
            self._browser = self._playwright.chromium.launch(
                headless=False, 
                args=["--disable-blink-features=AutomationControlled","--no-sandbox",]
            )
        return self
    
    def __exit__(self, exc_type, exc_val, exc_tb):
        if self._browser:
            self._browser.close()
            self._browser = None
        if self._playwright:
            self._playwright.stop()
            self._playwright = None
    
    @contextmanager
    def new_page(self):
        """Creates a new page (tab) in the existing browser"""
        page = self._browser.new_page()
        try:
            yield page
        finally:
            page.close()