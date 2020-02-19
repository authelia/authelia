#!/bin/bash
set -eu

for AUR_PACKAGE in authelia authelia-bin authelia-git;
do
  echo "  - label: \":linux: Deploy AUR Package [${AUR_PACKAGE}]\""
  echo "    commands:"
  echo "      - \"aurhelper.sh\""
  echo "    agents:"
  echo "      upload: \"fast\""
  echo "    env:"
  echo "      "PACKAGE: ${AUR_PACKAGE}""
  if [[ "${AUR_PACKAGE}" != "authelia-git" ]]; then
    echo "    depends_on:"
    echo "      - \"artifacts\""
    echo "    if: build.tag != null"
  else
    echo "    if: build.branch == \"master\""
  fi
done