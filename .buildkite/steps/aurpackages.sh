#!/usr/bin/env bash
set -eu

for AUR_PACKAGE in authelia authelia-bin authelia-git; do
cat << EOF
  - label: ":linux: Deploy AUR Package [${AUR_PACKAGE}]"
    command: "aurhelper.sh"
    agents:
      upload: "fast"
    env:
      PACKAGE: "${AUR_PACKAGE}"
EOF
if [[ "${AUR_PACKAGE}" != "authelia-git" ]]; then
cat << EOF
    depends_on:
      - "artifacts"
EOF
fi
cat << EOF
    if: build.tag != null
EOF
done
