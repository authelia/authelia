#!/bin/bash
set -eu

for BUILD_ARCH in amd64 arm32v7 arm64v8;
do
  echo "  - commands:"
  echo "    - \"authelia-scripts docker build --arch=${BUILD_ARCH}\""
  echo "    label: \":docker: Build Image [${BUILD_ARCH}]\""
  echo "    artifact_paths:"
  echo "      - \"authelia-linux-${BUILD_ARCH}.tar.gz\""
  echo "      - \"authelia-linux-${BUILD_ARCH}.tar.gz.sha256\""
  echo "    env:"
  echo "      "ARCH: ${BUILD_ARCH}""
done