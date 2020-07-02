#!/usr/bin/env bash
set -eu

declare -A BUILDS=(["linux"]="amd64 arm32v7 arm64v8 coverage")

for BUILD_OS in "${!BUILDS[@]}"; do
  for BUILD_ARCH in ${BUILDS[$BUILD_OS]}; do
cat << EOF
  - label: ":docker: Build Image [${BUILD_ARCH}]"
    command: "authelia-scripts docker build --arch=${BUILD_ARCH}"
    agents:
      build: "${BUILD_OS}-${BUILD_ARCH}"
    artifact_paths:
      - "authelia-image-${BUILD_ARCH}.tar.zst"
EOF
if [[ "${BUILD_ARCH}" != "coverage" ]]; then
cat << EOF
      - "authelia-${BUILD_OS}-${BUILD_ARCH}.tar.gz"
      - "authelia-${BUILD_OS}-${BUILD_ARCH}.tar.gz.sha256"
EOF
fi
cat << EOF
    env:
      ARCH: "${BUILD_ARCH}"
      OS: "${BUILD_OS}"
    key: "build-docker-${BUILD_OS}-${BUILD_ARCH}"
EOF
if [[ "${BUILD_ARCH}" == "coverage" ]]; then
cat << EOF
    if: build.branch !~ /^(v[0-9]+\.[0-9]+\.[0-9]+)$\$/
EOF
fi
  done
done