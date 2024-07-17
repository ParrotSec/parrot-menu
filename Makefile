.PHONY: icons

install:
	mkdir -p build
	nim c --nimcache:/tmp -d:release -o:build/update-launchers launcher-updater/update_launchers.nim

icons:
	@read -p "Enter path image: " IMAGE_PATH; \
	if [ -z "$$IMAGE_PATH" ]; then \
		echo "No image path provided"; \
		exit 1; \
	fi; \
	python3 generate_icons.py $$IMAGE_PATH