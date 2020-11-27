#!/usr/bin/env bash
set -eu

for SUITE_NAME in $(authelia-scripts suites list); do
cat << EOF
  - label: ":selenium: ${SUITE_NAME} Suite"
    command: "authelia-scripts --log-level debug suites test ${SUITE_NAME} --headless"
    retry:
      automatic: true
EOF
if [[ "${SUITE_NAME}" = "ActiveDirectory" ]]; then
cat << EOF
    agents:
      suite: "activedirectory"
EOF
elif [[ "${SUITE_NAME}" = "Kubernetes" ]]; then
cat << EOF
    agents:
      suite: "kubernetes"
EOF
else
cat << EOF
    agents:
      suite: "all"
EOF
fi
done