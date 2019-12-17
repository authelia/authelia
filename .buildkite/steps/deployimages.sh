#!/bin/bash
set -eu

for BUILD_ARCH in amd64 arm32v7 arm64v8;
do
  echo "  - commands:"
  echo "    - \"authelia-scripts docker push-image --arch=${BUILD_ARCH}\""
  echo "    label: \":docker: Deploy Image [${BUILD_ARCH}]\""
  echo "    agents:"
  echo "      "upload: fast""
  echo "    env:"
  echo "      "ARCH: ${BUILD_ARCH}""
done