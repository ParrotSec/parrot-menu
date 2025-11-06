.PHONY: binary icons

export BINARY_NAME := update-launchers
export BUILD_DIR := build
export HOME = $(CURDIR)

binary:
	mkdir -p $(BUILD_DIR)
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) ./launcher-updater/main.go

icons:
	@read -p "Enter path image: " IMAGE_PATH; \
	if [ -z "$$IMAGE_PATH" ]; then \
		echo "No image path provided"; \
		exit 1; \
	fi; \
	python3 generate_icons.py $$IMAGE_PATH
