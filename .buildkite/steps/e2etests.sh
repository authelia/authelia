#!/usr/bin/env bash
set -eu

declare -A SUITE_AGENTS=(
  [ActiveDirectory]="activedirectory"
  [HighAvailability]="highavailability"
  [Kubernetes]="kubernetes"
  [Standalone]="standalone"
)

declare -A SUITE_TIMEOUTS=(
  [Kubernetes]="30"
)

for SUITE_NAME in $(authelia-scripts suites list); do
  AGENT="${SUITE_AGENTS[${SUITE_NAME}]:-all}"
  TIMEOUT="${SUITE_TIMEOUTS[${SUITE_NAME}]:-20}"
cat << EOF
  - label: ":selenium: ${SUITE_NAME} Suite"
    command: "authelia-scripts --log-level debug suites test ${SUITE_NAME} --failfast --headless"
    artifact_paths:
      - "test-results-*.json"
      - "test-results-*.xml"
      - "screenshots/**/*.png"
      - "screenshots/**/*.html"
      - "screenshots/**/*.resources.json"
      - "screenshots/**/*.containers.log"
    timeout_in_minutes: ${TIMEOUT}
    retry:
      automatic:
        - exit_status: "*"
          limit: 2
      manual:
        permit_on_passed: true
    agents:
      suite: "${AGENT}"
    env:
      SUITE: "${SUITE_NAME}"
EOF
done
