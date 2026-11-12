#!/usr/bin/env sh
set -eu
cd "$(dirname "$0")"
mkdir -p dist
VERSION=1.0.0
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "-s -w -X main.version=$VERSION" -o dist/s5levelctl-windows-amd64.exe ./cmd/s5levelctl
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w -X main.version=$VERSION" -o dist/s5levelctl-linux-amd64 ./cmd/s5levelctl
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags "-s -w -X main.version=$VERSION" -o dist/s5levelctl-linux-arm64 ./cmd/s5levelctl
echo "Builds wurden unter dist/ erstellt."
