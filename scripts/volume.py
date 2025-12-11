# %%
from bs4 import BeautifulSoup
from selenium import webdriver
from pyvirtualdisplay import Display
from selenium.webdriver.common.by import By
from selenium.webdriver.chrome.options import Options

from PIL import Image
import requests
import time
from io import BytesIO
from PyPDF2 import PdfMerger
import os

# %%
# options = webdriver.ChromeOptions()
# options.add_argument('--headless')
# Replace with your remote Selenium server URL
driver = webdriver.Chrome(
    # command_executor='https://hoa-reformed-kimberlee.ngrok-free.dev',
    # options=options
)

# %%
url = "https://weebcentral.com/series/01J76XYCERXE60T7FKXVCCAQ0H/Jujutsu-Kaisen"

# %%
driver.get(url)
button = driver.find_element(By.CSS_SELECTOR, 'button[hx-target="#chapter-list"]')
button.click()

time.sleep(5)

# %%
html_content = driver.page_source
soup = BeautifulSoup(html_content, 'html.parser')

# %%
chapters = soup.find(id="chapter-list").find_all("a")
links = [chapter["href"] for chapter in chapters]
links.reverse()

# %%
driver.get(links[209])
required = links[209:209+9]

# %%
for link in required:
    driver.get(link)
    time.sleep(2)  # Wait for the page to load completely
    html_content = driver.page_source
    soup = BeautifulSoup(html_content, 'html.parser')
    
    title = soup.find("title").text.split(" | ")
    name, chapter = title[1], title[0].split()[1]

    image_section = soup.select_one('section[hx-include="[name=\'reading_style\']"]')
    image_tags = image_section.find_all("img")
    image_links = [img["src"] for img in image_tags]

    images = []
    
    for img_url in image_links:
        response = requests.get(img_url)
        img = Image.open(BytesIO(response.content)).convert("RGB")
        images.append(img)
    
    if images:
        pdf_path = f"./pdfs/{name} - {chapter}.pdf"
        images[0].save(pdf_path, save_all=True, append_images=images[1:])
        print(f"Saved {pdf_path}")

# %%
pdf_files = os.scandir("./pdfs")

merger = PdfMerger()
for pdf in pdf_files:
    merger.append(pdf.path)

merger.write("merged.pdf")
merger.close()


