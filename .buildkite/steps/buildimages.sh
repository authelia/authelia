#!/bin/bash
set -eu

for BUILD_OS in linux darwin; do
  if [[ $BUILD_OS == "linux" ]]; then
    for BUILD_ARCH in amd64 arm32v7 arm64v8; do
cat << EOF
  - label: ":docker: Build Image [${BUILD_ARCH}]"
    command: "authelia-scripts docker build --arch=${BUILD_ARCH}"
    agents:
      build: "true"
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
  else
    for BUILD_ARCH in amd64; do
cat << EOF
  - label: ":docker: Build Image [${BUILD_OS}]"
    command: "authelia-scripts docker build --arch=${BUILD_OS}"
    agents:
      build: "true"
    artifact_paths:
      - "authelia-${BUILD_OS}-${BUILD_ARCH}.tar.gz"
      - "authelia-${BUILD_OS}-${BUILD_ARCH}.tar.gz.sha256"
    env:
      ARCH: "${BUILD_ARCH}"
      OS: "${BUILD_OS}"
    key: "build-docker-${BUILD_OS}-${BUILD_ARCH}"
EOF
    done
  fi
done