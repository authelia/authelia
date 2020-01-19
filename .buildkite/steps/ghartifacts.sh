#!/bin/bash
set -eu

artifacts=()

for FILES in \
  authelia-linux-amd64.tar.zst authelia-linux-amd64.tar.zst.sha256 \
  authelia-linux-arm32v7.tar.zst authelia-linux-arm32v7.tar.zst.sha256 \
  authelia-linux-arm64v8.tar.zst authelia-linux-arm64v8.tar.zst.sha256;
do
  artifacts+=(-a "${FILES}")
done

echo "--- :github: Deploy artifacts for release: ${BUILDKITE_TAG}"
hub release create "${artifacts[@]}" -m "${BUILDKITE_TAG}" "${BUILDKITE_TAG}"