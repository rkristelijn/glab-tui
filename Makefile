.PHONY: build test fmt vet clean install dev commit-check release build-all

# Build the application (optimized)
build:
	@echo "🏗️  Building glab-tui..."
	go build -ldflags="-s -w" -o glab-tui .

# Build unoptimized (for debugging)
build-debug:
	@echo "🐛 Building debug version..."
	go build -o glab-tui-debug .

# Run tests
test:
	@echo "🧪 Running tests..."
	go test -v ./...

# Format Go code
fmt:
	@echo "📝 Formatting Go code..."
	go fmt ./...

# Run go vet
vet:
	@echo "🔍 Running go vet..."
	go vet ./...

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -f glab-tui glab-tui-debug glab-tui-*
	go clean

# Install dependencies
install:
	@echo "📦 Installing dependencies..."
	go mod download
	go mod tidy

# Development setup
dev: install fmt vet build
	@echo "✅ Development environment ready!"

# Check commit readiness (run before committing)
commit-check: fmt vet build test
	@echo "🔒 Checking for sensitive PII (excluding public GitHub username)..."
	@if grep -r -E "(remi\.kristelijn|nlrxk0145)" . --exclude-dir=.git --exclude="glab-tui*" --exclude="*.log" --exclude="Makefile" --exclude-dir=".git/hooks" 2>/dev/null; then \
		echo "❌ Sensitive PII found"; \
		exit 1; \
	fi
	@echo "✅ Ready to commit!"

# Quick development cycle
quick: fmt build
	@echo "⚡ Quick build complete!"

# Release build (with maximum optimizations)
release:
	@echo "🚀 Building release version..."
	CGO_ENABLED=0 go build -ldflags="-w -s -X main.version=$(shell git describe --tags --always)" -o glab-tui .

# Build for multiple platforms
build-all:
	@echo "🌍 Building for multiple platforms..."
	@mkdir -p dist
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o dist/glab-tui-linux-amd64 .
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-w -s" -o dist/glab-tui-darwin-amd64 .
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-w -s" -o dist/glab-tui-darwin-arm64 .
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-w -s" -o dist/glab-tui-windows-amd64.exe .
	@echo "✅ Multi-platform builds complete in dist/"

# Compress binaries with UPX (optional - requires UPX installed)
compress:
	@echo "🗜️  Compressing binaries with UPX..."
	@if command -v upx >/dev/null 2>&1; then \
		upx --best glab-tui 2>/dev/null || echo "⚠️  UPX compression failed or already compressed"; \
	else \
		echo "⚠️  UPX not installed. Install with: brew install upx"; \
	fi

# Show binary size
size:
	@echo "📊 Binary size information:"
	@ls -lh glab-tui* 2>/dev/null || echo "No binaries found. Run 'make build' first."

# Help
help:
	@echo "Available commands:"
	@echo "  build        - Build optimized application (-s -w flags)"
	@echo "  build-debug  - Build unoptimized for debugging"
	@echo "  test         - Run tests"
	@echo "  fmt          - Format Go code"
	@echo "  vet          - Run go vet"
	@echo "  clean        - Clean build artifacts"
	@echo "  install      - Install dependencies"
	@echo "  dev          - Full development setup"
	@echo "  commit-check - Check if ready to commit"
	@echo "  quick        - Quick format and build"
	@echo "  release      - Build maximum optimized release"
	@echo "  build-all    - Build for multiple platforms"
	@echo "  compress     - Compress binary with UPX"
	@echo "  size         - Show binary size information"
	@echo "  help         - Show this help"
