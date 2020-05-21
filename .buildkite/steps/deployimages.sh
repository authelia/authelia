#!/bin/bash
set -eu

for BUILD_ARCH in amd64 arm32v7 arm64v8; do
cat << EOF
  - label: ":docker: Deploy Image [${BUILD_ARCH}]"
    command: "authelia-scripts docker push-image --arch=${BUILD_ARCH}"
    depends_on:
EOF
if [[ "${BUILD_ARCH}" == "amd64" ]]; then
cat << EOF
      - "build-docker-linux-amd64"
EOF
elif [[ "${BUILD_ARCH}" == "arm32v7" ]]; then
cat << EOF
      - "build-docker-linux-arm32v7"
EOF
else
cat << EOF
      - "build-docker-linux-arm64v8"
EOF
fi
cat << EOF
    agents:
      upload: "fast"
    env:
      ARCH: "${BUILD_ARCH}"
EOF
done