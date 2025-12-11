import requests, json, os
from PIL import Image

anime_name = "Vinland Saga"
json_path = f"./json/{anime_name.lower().replace(' ', '_')}_chapters.json"

def download_images(payload):
    res = requests.post("http://localhost:8080/download", json=payload)
    print(res.status_code)
    
with open(json_path, "r", encoding="utf-8") as f:
    chapters_map = json.load(f)

    for chapter_name, chapter_info in chapters_map.items():
        print("Downloading images for chapter:", chapter_name)
        image_links = chapter_info["images"]
        if len(image_links) > 0:
            payload = {
                "folder": f"{anime_name}/{chapter_name.lower().replace(' ', '_').replace('.', '_')}",
                "image_links": image_links
            }
            download_images(payload)
            
images = []

for chapter_folder in sorted(os.listdir(anime_name)):
    chapter_path = os.path.join(anime_name, chapter_folder)
    
    if os.path.isdir(chapter_path):
        for image_file in sorted(os.listdir(chapter_path)):
            image_path = os.path.join(chapter_path, image_file)
            images.append(Image.open(image_path).convert("RGB"))