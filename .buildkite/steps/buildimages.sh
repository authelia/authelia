#!/bin/bash
set -eu

declare -A BUILDS=(["linux"]="amd64 arm32v7 arm64v8" ["darwin"]="amd64")

for BUILD_OS in "${!BUILDS[@]}"; do
  for BUILD_ARCH in ${BUILDS[$BUILD_OS]}; do
if [[ "${BUILD_OS}" == "darwin" ]]; then
cat << EOF
  - label: ":docker: Build Image [${BUILD_OS}]"
    command: "authelia-scripts docker build --arch=${BUILD_OS}"
EOF
else
cat << EOF
  - label: ":docker: Build Image [${BUILD_ARCH}]"
    command: "authelia-scripts docker build --arch=${BUILD_ARCH}"
EOF
fi
cat << EOF
    agents:
      build: "${BUILD_OS}-${BUILD_ARCH}"
    artifact_paths:
EOF
if [[ "${BUILD_OS}" == "linux" ]]; then
cat << EOF
      - "authelia-image-${BUILD_ARCH}.tar.zst"
EOF
fi
cat << EOF
      - "authelia-${BUILD_OS}-${BUILD_ARCH}.tar.gz"
      - "authelia-${BUILD_OS}-${BUILD_ARCH}.tar.gz.sha256"
    env:
      ARCH: "${BUILD_ARCH}"
      OS: "${BUILD_OS}"
    key: "build-docker-${BUILD_OS}-${BUILD_ARCH}"
EOF
  done
done