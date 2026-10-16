#!/bin/sh

set -eu
cd "$(dirname "$0")/.."

build_arch="${GOARCH:-amd64}"
GOARCH="$build_arch" sh scripts/build.sh
docker build --platform="linux/$build_arch" \
  -f "docker/Dockerfile.linux.$build_arch" \
  -t "${IMAGE_TAG:-plugins/github-actions}" .
