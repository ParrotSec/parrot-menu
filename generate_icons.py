from PIL import Image
import os
import sys

def generate_icons(image_path, sizes, output_dirs):
    try:
        with Image.open(image_path) as img:
            for size, output_dir in zip(sizes, output_dirs):
                resized_img = img.resize((size, size), Image.LANCZOS)
                if not os.path.exists(output_dir):
                    os.makedirs(output_dir)
                output_path = os.path.join(output_dir, os.path.basename(image_path))
                resized_img.save(output_path)
                # print(f"Saved resized image {output_path}")
    except Exception as e:
        print(f"An error occurred: {e}")

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: python generate_icons.py <path_to_image>")
        sys.exit(1)
    
    image_path = sys.argv[1]
    
    sizes = [16, 24, 32, 48, 256]
    output_dirs = [
        "menu-icons/hicolor/16x16/apps/",
        "menu-icons/hicolor/24x24/apps/",
        "menu-icons/hicolor/32x32/apps/",
        "menu-icons/hicolor/48x48/apps/",
        "menu-icons/hicolor/256x256/apps/"
    ]

    generate_icons(image_path, sizes, output_dirs)
