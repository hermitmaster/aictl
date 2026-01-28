.PHONY: build install uninstall clean test lint coverage completions

VERSION := 0.1.0
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BINARY := aictl
BUILD_DIR := bin
INSTALL_DIR := /usr/local/bin
COMPLETION_DIR := $(HOME)/.local/share/aictl/completions

LDFLAGS := -ldflags "\
	-X github.com/hermitmaster/aictl/internal/cli.version=$(VERSION) \
	-X github.com/hermitmaster/aictl/internal/cli.commit=$(COMMIT) \
	-X github.com/hermitmaster/aictl/internal/cli.buildDate=$(BUILD_DATE)"

build:
	@echo "Building $(BINARY) $(VERSION) ($(COMMIT))..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) ./cmd/aictl

install: build completions
	@echo "Installing $(BINARY) to $(INSTALL_DIR)..."
	@mkdir -p $(INSTALL_DIR)
	@cp $(BUILD_DIR)/$(BINARY) $(INSTALL_DIR)/$(BINARY)
	@chmod +x $(INSTALL_DIR)/$(BINARY)
	@echo ""
	@echo "✓ Installed $(BINARY) to $(INSTALL_DIR)/$(BINARY)"
	@echo ""
	@echo "To enable shell completions, add to your shell config:"
	@echo "  # Bash (~/.bashrc)"
	@echo "  source $(COMPLETION_DIR)/aictl.bash"
	@echo ""
	@echo "  # Zsh (~/.zshrc)"
	@echo "  source $(COMPLETION_DIR)/aictl.zsh"
	@echo ""
	@echo "  # Fish (~/.config/fish/config.fish)"
	@echo "  source $(COMPLETION_DIR)/aictl.fish"

uninstall:
	@echo "Uninstalling $(BINARY)..."
	@rm -f $(INSTALL_DIR)/$(BINARY)
	@rm -rf $(COMPLETION_DIR)
	@echo "✓ Uninstalled $(BINARY)"

completions: build
	@echo "Generating shell completions..."
	@mkdir -p $(COMPLETION_DIR)
	@$(BUILD_DIR)/$(BINARY) completion bash > $(COMPLETION_DIR)/aictl.bash
	@$(BUILD_DIR)/$(BINARY) completion zsh > $(COMPLETION_DIR)/aictl.zsh
	@$(BUILD_DIR)/$(BINARY) completion fish > $(COMPLETION_DIR)/aictl.fish
	@echo "✓ Completions saved to $(COMPLETION_DIR)"

clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)

test:
	@echo "Running tests..."
	go test -v ./...

coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | grep total
	@rm -f coverage.out

lint:
	@echo "Running linter..."
	golangci-lint run

# Development helpers
run:
	go run ./cmd/aictl $(ARGS)

# Cross-platform builds
build-all: build-linux build-darwin build-windows

build-linux:
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-linux-amd64 ./cmd/aictl
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-linux-arm64 ./cmd/aictl

build-darwin:
	@echo "Building for macOS..."
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-darwin-amd64 ./cmd/aictl
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-darwin-arm64 ./cmd/aictl

build-windows:
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-windows-amd64.exe ./cmd/aictl

# Release helper
release: clean build-all
	@echo "Creating release archives..."
	@cd $(BUILD_DIR) && for f in $(BINARY)-*; do \
		if [ -f "$$f" ]; then \
			tar -czf "$$f.tar.gz" "$$f"; \
		fi; \
	done
	@echo "✓ Release archives created in $(BUILD_DIR)"
