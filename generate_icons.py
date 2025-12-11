from PIL import Image
import os
import sys

def save_icon(img: Image.Image, size: int, output_dir: str, filename: str) -> None:
    resized_img = img.resize((size, size), Image.LANCZOS)
    
    if not os.path.exists(output_dir):
        os.makedirs(output_dir)
    
    output_path = os.path.join(output_dir, filename)
    resized_img.save(output_path)
    print(f"Saved: {output_path} (Resolution: {size}x{size})")

def generate_icons(image_path: str, base_sizes: list[int]) -> None:
    try:
        filename = os.path.basename(image_path)
        
        with Image.open(image_path) as img:
            for size in base_sizes:

                dir_1x = f"menu-icons/hicolor/{size}x{size}/apps/"
                save_icon(img, size, dir_1x, filename)

                size_2x = size * 2
                dir_2x = f"menu-icons/hicolor/{size}x{size}@2/apps/"
                save_icon(img, size_2x, dir_2x, filename)

    except Exception as e:
        print(f"An error occurred: {e}")

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: python3 generate_icons.py <path_to_image>")
        sys.exit(1)
    
    image_path: str = sys.argv[1]
    
    sizes: list[int] = [16, 24, 32, 48, 256]

    generate_icons(image_path, sizes)