import os
from PIL import Image

# Use 256x256 directory as the "master" source.
SOURCE_DIR = "menu-icons/hicolor/256x256/apps/"

# Define the logical sizes we want to generate 
# since we will generate both standard (1x) and HiDPI (@2) versions for these.
TARGET_SIZES = [16, 24, 32, 48]
# Also, we exclude 256 from this list because it is the source itself.

def process_all_icons():
    if not os.path.exists(SOURCE_DIR):
        print(f"Source directory not found: {SOURCE_DIR}")
        return

    files = [f for f in os.listdir(SOURCE_DIR) if f.endswith('.png')]
    total_files = len(files)
    
    if total_files == 0:
        print("[!] No .png files found in the source directory.")
        return

    print(f"Found {total_files} icons. Generating the icon set (1x and 2x)...")

    for i, filename in enumerate(files, 1):
        source_path = os.path.join(SOURCE_DIR, filename)
        
        try:
            with Image.open(source_path) as img:
                for size in TARGET_SIZES:

                    dir_1x = f"menu-icons/hicolor/{size}x{size}/apps/"
                    if not os.path.exists(dir_1x):
                        os.makedirs(dir_1x)
                    
                    img_1x = img.resize((size, size), Image.LANCZOS)
                    img_1x.save(os.path.join(dir_1x, filename))

                    size_2x = size * 2
                    dir_2x = f"menu-icons/hicolor/{size}x{size}@2/apps/"
                    if not os.path.exists(dir_2x):
                        os.makedirs(dir_2x)

                    img_2x = img.resize((size_2x, size_2x), Image.LANCZOS)
                    img_2x.save(os.path.join(dir_2x, filename))
                
                print(f"[{i}/{total_files}] Processed: {filename}")

        except Exception as e:
            print(f"[!] Error processing {filename}: {e}")

    print("[!] Complete generation finished.")

if __name__ == "__main__":
    process_all_icons()