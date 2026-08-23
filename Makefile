.PHONY: binary icons

export BUILD_DIR := build
export HOME = $(CURDIR)
export GO111MODULE = on
export GOFLAGS = -mod=mod
IMAGES ?= $(IMAGE)

binary:
	mkdir -p $(BUILD_DIR)
	cd launcher-updater && go build -ldflags="-s -w" -o $(CURDIR)/$(BUILD_DIR)/update-launchers ./cmd/launcher-updater
	cd parrot-exec && go build -ldflags="-s -w" -o $(CURDIR)/$(BUILD_DIR)/parrot-exec .
	ln -sf parrot-exec $(BUILD_DIR)/parrot-ls

icons:
	python3 generate_icons.py $(IMAGES)
