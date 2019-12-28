#!/bin/bash
set -eu

for SUITE_NAME in $(authelia-scripts suites list);
do
  echo "  - label: \":selenium: ${SUITE_NAME} Suite\""
  echo "    commands:"
  echo "      - \"authelia-scripts --log-level debug suites test ${SUITE_NAME} --headless\""
  echo "    retry:"
  echo "      "automatic: true""
  if [[ "${SUITE_NAME}" != "Kubernetes" ]];
  then
    echo "    agents:"
    echo "      "suite: all""
  else
    echo "    agents:"
    echo "      "suite: kubernetes""
  fi
done