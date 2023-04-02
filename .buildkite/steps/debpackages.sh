#!/usr/bin/env bash

# SPDX-FileCopyrightText: 2019 Authelia
#
# SPDX-License-Identifier: Apache-2.0

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
EOF
elif [[ "${DEB_PACKAGE}" == "armhf" ]]; then
cat << EOF
      ARCH: "arm"
EOF
else
cat << EOF
      ARCH: "arm64"
EOF
fi
cat << EOF
    depends_on:
      - "unit-test"
    key: "build-deb-package-${DEB_PACKAGE}"
EOF
done