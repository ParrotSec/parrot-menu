from __future__ import annotations

import argparse
import logging
import shutil
from dataclasses import dataclass
from pathlib import Path

try:
    from PIL import Image
except ImportError as exc:
    raise SystemExit("Pillow is required to generate icons. Install python3-pil.") from exc

ICON_ROOT = Path("menu-icons/hicolor")
SOURCE_DIR = ICON_ROOT / "256x256" / "apps"
SOURCE_SIZE = 256

# Sizes follow the hicolor layout shipped by this package.
TARGET_SIZES = (16, 22, 24, 32, 48)
DENSITY_VARIANTS = (("", 1), ("@2", 2))

try:
    RESAMPLE_FILTER = Image.Resampling.LANCZOS
except AttributeError:
    RESAMPLE_FILTER = Image.LANCZOS


@dataclass(frozen=True)
class IconJob:
    source_path: Path
    file_name: str
    import_source: bool


def setup_logging() -> None:
    logging.basicConfig(
        level=logging.INFO,
        format="%(levelname)s: %(message)s",
    )


def collect_icon_jobs(paths: list[Path]) -> list[IconJob]:
    if paths:
        return [build_import_job(path) for path in paths]

    if not SOURCE_DIR.is_dir():
        raise FileNotFoundError(f"Source directory not found: {SOURCE_DIR}")

    source_icons = sorted(
        path for path in SOURCE_DIR.iterdir() if path.is_file() and is_png(path)
    )
    if not source_icons:
        raise FileNotFoundError(f"No .png files found in {SOURCE_DIR}")

    return [
        IconJob(source_path=path, file_name=path.name, import_source=False)
        for path in source_icons
    ]


def build_import_job(path: Path) -> IconJob:
    source_path = path.expanduser()
    if not source_path.is_file():
        raise FileNotFoundError(f"Image not found: {source_path}")
    if not is_png(source_path):
        raise ValueError(f"Only .png images are supported: {source_path}")

    source_dest = SOURCE_DIR / source_path.name
    return IconJob(
        source_path=source_path,
        file_name=source_path.name,
        import_source=source_path.resolve() != source_dest.resolve(),
    )


def is_png(path: Path) -> bool:
    return path.suffix.lower() == ".png"


def process_icons(jobs: list[IconJob], dry_run: bool) -> int:
    logging.info("Processing %d icon(s)...", len(jobs))

    failures = 0
    for index, job in enumerate(jobs, 1):
        try:
            process_icon(job, dry_run)
            logging.info("[%d/%d] Processed: %s", index, len(jobs), job.file_name)
        except OSError as exc:
            failures += 1
            logging.error("Error processing %s: %s", job.source_path, exc)

    return failures


def process_icon(job: IconJob, dry_run: bool) -> None:
    with Image.open(job.source_path) as image:
        image.load()

        if image.width != image.height:
            logging.warning(
                "%s is not square; generated icons will be stretched",
                job.source_path,
            )

        if job.import_source:
            write_source_icon(
                image,
                job.source_path,
                SOURCE_DIR / job.file_name,
                dry_run,
            )

        for size in TARGET_SIZES:
            for suffix, multiplier in DENSITY_VARIANTS:
                target_size = size * multiplier
                destination = (
                    ICON_ROOT / f"{size}x{size}{suffix}" / "apps" / job.file_name
                )
                write_resized_icon(image, destination, target_size, dry_run)


def write_source_icon(
    image: Image.Image,
    source_path: Path,
    destination: Path,
    dry_run: bool,
) -> None:
    if dry_run:
        logging.info("[DRY RUN] Would import %s", destination)
        return

    destination.parent.mkdir(parents=True, exist_ok=True)
    if image.size == (SOURCE_SIZE, SOURCE_SIZE):
        shutil.copy2(source_path, destination)
        return

    write_resized_icon(image, destination, SOURCE_SIZE, dry_run=False)


def write_resized_icon(
    image: Image.Image,
    destination: Path,
    size: int,
    dry_run: bool,
) -> None:
    if dry_run:
        logging.info("[DRY RUN] Would write %s (%dx%d)", destination, size, size)
        return

    destination.parent.mkdir(parents=True, exist_ok=True)
    image.resize((size, size), RESAMPLE_FILTER).save(destination)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description=(
            "Generate scaled app icons from the 256x256 icon source directory, "
            "or import specific PNG files before generating their derived sizes."
        )
    )
    parser.add_argument(
        "images",
        nargs="*",
        type=Path,
        help=(
            "PNG image path(s) to import into menu-icons/hicolor/256x256/apps. "
            "When omitted, every source icon is regenerated."
        ),
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Preview the generated files without modifying anything",
    )
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    setup_logging()

    try:
        jobs = collect_icon_jobs(args.images)
    except (FileNotFoundError, ValueError) as exc:
        logging.error("%s", exc)
        return 1

    failures = process_icons(jobs, args.dry_run)
    if failures:
        logging.error("Generation failed for %d icon(s).", failures)
        return 1

    if args.dry_run:
        logging.info("Dry run complete. No files were modified.")
    else:
        logging.info("Generation complete.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
