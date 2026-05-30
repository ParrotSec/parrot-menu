import argparse
import logging
import os

from PIL import Image

SOURCE_DIR = "menu-icons/hicolor/256x256/apps/"

# Sizes follow the Breeze icon theme logic (see index.theme):
#   16  fixed — smallest toolbar/panel icons
#   22  fixed — KDE toolbar workaround
#   24        — derived from 48 via scaling rule (48 ÷ 2)
#   32  fixed — System Settings entry, scalable to 32/64/128/256
#   48  scalable (MinSize=48, MaxSize=256) — app icons, derives 48/96 and 24
# @2x variants (HiDPI) are generated automatically from each size.
TARGET_SIZES: list[int] = [16, 22, 24, 32, 48]

def setup_logging(dry_run: bool) -> None:
    logging.basicConfig(
        level=logging.DEBUG if dry_run else logging.INFO,
        format="%(levelname)s: %(message)s",
    )

def process_all_icons(dry_run: bool) -> None:
    if not os.path.exists(SOURCE_DIR):
        logging.error("Source directory not found: %s", SOURCE_DIR)
        return

    files = [f for f in os.listdir(SOURCE_DIR) if f.endswith(".png")]
    total_files = len(files)

    if total_files == 0:
        logging.warning("No .png files found in the source directory.")
        return

    logging.info("Found %d icons. Generating the icon set (1x and 2x)...", total_files)

    for i, filename in enumerate(files, 1):
        source_path = os.path.join(SOURCE_DIR, filename)

        try:
            with Image.open(source_path) as img:
                for size in TARGET_SIZES:
                    for suffix, multiplier in [("", 1), ("@2", 2)]:
                        dir_name = f"menu-icons/hicolor/{size}x{size}{suffix}/apps/"
                        dest_path = os.path.join(dir_name, filename)
                        target_size = size * multiplier

                        if dry_run:
                            logging.info("[DRY RUN] Would create %s", dest_path)
                        else:
                            os.makedirs(dir_name, exist_ok=True)
                            resized = img.resize((target_size, target_size), Image.LANCZOS)
                            resized.save(dest_path)

                logging.info("[%d/%d] Processed: %s", i, total_files, filename)

        except Exception as e:
            logging.error("Error processing %s: %s", filename, e)

    if dry_run:
        logging.info("Dry run complete. No files were modified.")
    else:
        logging.info("Generation complete.")


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Generate icon set from 256x256 icons.")
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Preview what would be done without modifying anything",
    )
    return parser.parse_args()

def main() -> None:
    args = parse_args()
    setup_logging(args.dry_run)
    process_all_icons(args.dry_run)

if __name__ == "__main__":
    main()