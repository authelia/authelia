#!/usr/bin/env bash
set -eu

for DEB_PACKAGE in amd64 armhf arm64; do
cat << EOF
  - label: ":debian: Build Package [${DEB_PACKAGE}]"
    command: "debhelper.sh"
    artifact_paths:
      - "*.deb"
      - "*.deb.sha256"
    env:
      PACKAGE: "${DEB_PACKAGE}"
EOF
if [[ "${DEB_PACKAGE}" == "amd64" ]]; then
cat << EOF
      ARCH: "${DEB_PACKAGE}"
    depends_on:
      - "build-docker-linux-amd64"
EOF
elif [[ "${DEB_PACKAGE}" == "armhf" ]]; then
cat << EOF
      ARCH: "arm32v7"
    depends_on:
      - "build-docker-linux-arm32v7"
EOF
else
cat << EOF
      ARCH: "arm64v8"
    depends_on:
      - "build-docker-linux-arm64v8"
EOF
fi
cat << EOF
    key: "build-deb-package-${DEB_PACKAGE}"
EOF
done