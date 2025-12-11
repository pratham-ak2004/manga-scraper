import os, json
import img2pdf
from natsort import natsorted
import zipfile

anime_name = "Vinland Saga"
json_path = f"./json/{anime_name.lower().replace(' ', '_')}_chapters.json"

batch_size = 500
part = 1

def save_pdf(image_paths, part):
    pdf_path = f"./pdfs/{anime_name}_{part}.pdf"
    # print(f"Saving PDF: {pdf_path}")

    with open(pdf_path, "wb") as f:
        f.write(img2pdf.convert(image_paths))

    print(f"Saved PDF: {pdf_path}")


# Load chapter structure
with open(json_path, "r", encoding="utf-8") as f:
    chapter_map = json.load(f)

image_paths = []

# Collect pages
for chapter in chapter_map.keys():
    folder = chapter.lower().replace(" ", "_").replace(".", "_")
    chapter_path = os.path.join(anime_name, folder)

    if os.path.isdir(chapter_path):
        for img in natsorted(os.listdir(chapter_path)):
            img_path = os.path.join(chapter_path, img)
            image_paths.append(img_path)

        # Save one volume
        if len(image_paths) >= batch_size:
            save_pdf(image_paths, part)
            part += 1
            image_paths = []

# Save leftover volume
if image_paths:
    save_pdf(image_paths, part)
    part += 1


# ZIP PDFs
def zip_pdfs(folder_path, output_zip):
    with zipfile.ZipFile(output_zip, "w", zipfile.ZIP_DEFLATED) as zipf:
        for file_name in os.listdir(folder_path):
            if (
                file_name.startswith(anime_name) 
                and file_name.endswith(".pdf")
            ):
                zipf.write(os.path.join(folder_path, file_name), arcname=file_name)

zip_pdfs("./pdfs", f"./zips/{anime_name}.zip")

print("Zipped successfully!")
