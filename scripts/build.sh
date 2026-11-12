#!/bin/sh

# Preserve the existing AMD64 default; CI selects each native architecture.
set -eu
export GOPATH=""
export CGO_ENABLED=0

build_arch="${GOARCH:-amd64}"
case "$build_arch" in
  amd64|arm64) ;;
  *) echo "Unsupported Linux architecture: $build_arch" >&2; exit 1 ;;
esac

GOOS=linux GOARCH="$build_arch" go build -o "release/linux/$build_arch/plugin" ./cmd
