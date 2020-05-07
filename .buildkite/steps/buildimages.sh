#!/bin/bash
set -eu

declare -A BUILDS=(["linux"]="amd64 arm32v7 arm64v8")

for BUILD_OS in "${!BUILDS[@]}"; do
  for BUILD_ARCH in ${BUILDS[$BUILD_OS]}; do
cat << EOF
  - label: ":docker: Build Image [${BUILD_ARCH}]"
    command: "authelia-scripts docker build --arch=${BUILD_ARCH}"
    agents:
      build: "${BUILD_OS}-${BUILD_ARCH}"
    artifact_paths:
      - "authelia-image-${BUILD_ARCH}.tar.zst"
      - "authelia-${BUILD_OS}-${BUILD_ARCH}.tar.gz"
      - "authelia-${BUILD_OS}-${BUILD_ARCH}.tar.gz.sha256"
    env:
      ARCH: "${BUILD_ARCH}"
      OS: "${BUILD_OS}"
    key: "build-docker-${BUILD_OS}-${BUILD_ARCH}"
EOF
  done
done