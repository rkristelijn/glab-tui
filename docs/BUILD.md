# Build Guide

This document explains how to build glab-tui with various optimizations.

## Quick Start

```bash
# Standard optimized build
make build

# Development build (unoptimized, for debugging)
make build-debug

# Maximum optimization for release
make release
```

## Build Targets

### Development Builds

- `make build` - Optimized build with `-s -w` flags (~20-30% smaller)
- `make build-debug` - Unoptimized build for debugging with symbols
- `make quick` - Fast format + build cycle

### Release Builds

- `make release` - Maximum optimization with version info
- `make build-all` - Multi-platform builds (Linux, macOS, Windows)

### Utilities

- `make size` - Show binary size information
- `make compress` - Compress with UPX (requires UPX installed)
- `make clean` - Remove all build artifacts

## Build Flags Explained

### Standard Build (`make build`)
```bash
go build -ldflags="-s -w" -o glab-tui .
```
- `-s` - Strip symbol table and debug info
- `-w` - Strip DWARF debug info
- **Result**: ~20-30% smaller binary

### Release Build (`make release`)
```bash
CGO_ENABLED=0 go build -ldflags="-w -s -X main.version=$(git describe --tags --always)" -o glab-tui .
```
- `CGO_ENABLED=0` - Static binary, no C dependencies
- `-X main.version=...` - Embed version info
- **Result**: Fully static, optimized binary

## Multi-Platform Builds

```bash
make build-all
```

Creates binaries for:
- Linux AMD64
- macOS AMD64 (Intel)
- macOS ARM64 (Apple Silicon)
- Windows AMD64

All binaries are placed in `dist/` directory.

## Size Optimization

### Current Sizes (approximate)
- Unoptimized: ~14MB
- With `-s -w`: ~9.8MB (30% reduction)
- With UPX: ~3-4MB (70% reduction)

### UPX Compression (Optional)

Install UPX:
```bash
# macOS
brew install upx

# Ubuntu/Debian
apt install upx-ucl
```

Compress binary:
```bash
make compress
```

**Trade-offs:**
- ✅ 50-70% size reduction
- ❌ Slightly slower startup (~50ms)
- ❌ Some antivirus false positives

## Automated Releases

GitHub Actions automatically builds and releases binaries when you push a tag:

```bash
git tag v1.0.3
git push origin v1.0.3
```

This creates:
- Multi-platform binaries
- GitHub release with assets
- Automatic release notes

## Development Workflow

```bash
# Setup development environment
make dev

# Quick development cycle
make quick

# Before committing
make commit-check

# Check binary size
make size
```

## Troubleshooting

### Large Binary Size
- Use `make release` instead of `make build-debug`
- Consider UPX compression for distribution
- Check `make size` to compare variants

### Build Failures
- Run `make clean` and try again
- Check Go version (requires Go 1.19+)
- Ensure all dependencies with `make install`

### Cross-Platform Issues
- Use `make build-all` for consistent multi-platform builds
- `CGO_ENABLED=0` ensures static binaries
- Test on target platforms when possible

## Performance Notes

- Optimized builds are ~30% smaller with no runtime performance impact
- UPX compression trades startup time for size
- Static binaries (`CGO_ENABLED=0`) are more portable
- Multi-platform builds ensure consistent behavior across systems
