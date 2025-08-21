.PHONY: build test fmt vet clean install dev commit-check

# Build the application
build:
	@echo "🏗️  Building glab-tui..."
	go build -o glab-tui .

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
	rm -f glab-tui
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
	@if grep -r -E "(remi\.kristelijn|nlrxk0145)" . --exclude-dir=.git --exclude="glab-tui" --exclude="*.log" --exclude="Makefile" --exclude-dir=".git/hooks" 2>/dev/null; then \
		echo "❌ Sensitive PII found"; \
		exit 1; \
	fi
	@echo "✅ Ready to commit!"

# Quick development cycle
quick: fmt build
	@echo "⚡ Quick build complete!"

# Release build (with optimizations)
release:
	@echo "🚀 Building release version..."
	CGO_ENABLED=0 go build -ldflags="-w -s" -o glab-tui .

# Help
help:
	@echo "Available commands:"
	@echo "  build        - Build the application"
	@echo "  test         - Run tests"
	@echo "  fmt          - Format Go code"
	@echo "  vet          - Run go vet"
	@echo "  clean        - Clean build artifacts"
	@echo "  install      - Install dependencies"
	@echo "  dev          - Full development setup"
	@echo "  commit-check - Check if ready to commit"
	@echo "  quick        - Quick format and build"
	@echo "  release      - Build optimized release"
	@echo "  help         - Show this help"
