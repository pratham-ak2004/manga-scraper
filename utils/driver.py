from selenium import webdriver
from selenium.webdriver.chrome.options import Options
from config.settings import settings
from contextlib import contextmanager

def get_chrome_options():
    """Configure Chrome options for headless browsing"""
    options = Options()
    # options.add_argument('--headless')
    options.add_argument('--no-sandbox')
    options.add_argument('--disable-dev-shm-usage')
    return options

@contextmanager
def get_webdriver():
    """
    Context manager to create and cleanup WebDriver instance.
    
    Usage:
        with get_webdriver() as browser:
            browser.get("https://example.com")
    """
    driver = None
    try:
        driver = webdriver.Remote(
            command_executor=settings.SELENIUM_HUB_URL,
            options=get_chrome_options()
        )
        yield driver
    finally:
        if driver:
            driver.quit()