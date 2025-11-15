#!/bin/bash
set -e

echo "Building subconvergo..."

cd "$(dirname "$0")"

# Download dependencies
go mod tidy

# Build for current platform
go build -ldflags="-s -w" -o subconvergo main.go

echo "Build completed: subconvergo"
echo ""
echo "Run with: ./subconvergo"
echo "The program will use ../base/ for configuration and templates"
