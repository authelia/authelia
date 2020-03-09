#!/bin/bash
set -eu

for BUILD_ARCH in amd64 arm32v7 arm64v8; do
cat << EOF
  - label: ":docker: Deploy Image [${BUILD_ARCH}]"
    command: "authelia-scripts docker push-image --arch=${BUILD_ARCH}"
    agents:
      upload: "fast"
    env:
      ARCH: "${BUILD_ARCH}"
EOF
done