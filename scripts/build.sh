#!/bin/bash

set -e

VERSION=${1:-"1.0.0"}
OUTPUT_DIR="./bin"

echo "Building ai-terminal v${VERSION}..."

mkdir -p "$OUTPUT_DIR"

echo "Building for current platform..."
go build -ldflags="-s -w -X main.version=${VERSION} -X main.commit=$(git rev-parse --short HEAD 2>/dev/null || echo "dev") -X main.date=$(date -u +%Y-%m-%d)" -o "$OUTPUT_DIR/ai-terminal" ./cmd/ai-terminal

echo "Build complete: ${OUTPUT_DIR}/ai-terminal"
echo ""
echo "To run: ./bin/ai-terminal"
