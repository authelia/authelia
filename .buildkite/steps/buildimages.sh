#!/bin/bash
set -eu

for BUILD_ARCH in amd64 arm32v7 arm64v8;
do
  echo "  - label: \":docker: Build Image [${BUILD_ARCH}]\""
  echo "    commands:"
  echo "      - \"authelia-scripts docker build --arch=${BUILD_ARCH}\""
  echo "    artifact_paths:"
  echo "      - \"authelia-image-${BUILD_ARCH}.tar.zst\""
  echo "      - \"authelia-linux-${BUILD_ARCH}.tar.gz\""
  echo "      - \"authelia-linux-${BUILD_ARCH}.tar.gz.sha256\""
  if [[ "${BUILD_ARCH}" != "amd64" ]];
  then
    echo "    branches: \"master v*\""
  fi
  echo "    env:"
  echo "      "ARCH: ${BUILD_ARCH}""
  echo "    key: \"build-docker-${BUILD_ARCH}\""
done