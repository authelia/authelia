#!/bin/bash
set -eu

for BUILD_ARCH in amd64 arm32v7 arm64v8; do
cat << EOF
  - label: ":docker: Build Image [${BUILD_ARCH}]"
    command: "authelia-scripts docker build --arch=${BUILD_ARCH}"
    agents:
      build: "true"
    artifact_paths:
      - "authelia-image-${BUILD_ARCH}.tar.zst"
      - "authelia-linux-${BUILD_ARCH}.tar.gz"
      - "authelia-linux-${BUILD_ARCH}.tar.gz.sha256"
    env:
      ARCH: "${BUILD_ARCH}"
    key: "build-docker-${BUILD_ARCH}"
EOF
done