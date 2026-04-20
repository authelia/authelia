#!/usr/bin/env bash
set -eu

declare -A SUITE_AGENTS=(
  [ActiveDirectory]="activedirectory"
  [HighAvailability]="highavailability"
  [Kubernetes]="kubernetes"
  [Standalone]="standalone"
)

for SUITE_NAME in $(authelia-scripts suites list); do
  AGENT="${SUITE_AGENTS[${SUITE_NAME}]:-all}"
cat << EOF
  - label: ":selenium: ${SUITE_NAME} Suite"
    command: "authelia-scripts --log-level debug suites test ${SUITE_NAME} --failfast --headless"
    retry:
      automatic: true
      manual:
        permit_on_passed: true
    agents:
      suite: "${AGENT}"
    env:
      SUITE: "${SUITE_NAME}"
EOF
done
