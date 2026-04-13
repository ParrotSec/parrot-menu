.PHONY: binary icons

export BUILD_DIR := build
export HOME = $(CURDIR)

binary:
	mkdir -p $(BUILD_DIR)
	cd launcher-updater && go build -ldflags="-s -w" -o ../$(BUILD_DIR)/update-launchers ./cmd/launcher-updater
	cd parrot-exec && go build -ldflags="-s -w" -o ../$(BUILD_DIR)/parrot-exec .

icons:
	@read -p "Enter path image: " IMAGE_PATH; \
	if [ -z "$$IMAGE_PATH" ]; then \
		echo "No image path provided"; \
		exit 1; \
	fi; \
	python3 generate_icons.py $$IMAGE_PATH
