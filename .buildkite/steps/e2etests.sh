#!/bin/bash
set -eu

for SUITE_NAME in BypassAll Docker DuoPush HighAvailability Kubernetes LDAP Mariadb NetworkACL Postgres ShortTimeouts Standalone Traefik; do
  echo "  - commands:"
  echo "    - \"buildkite-agent artifact download "dist/*" .\""
  echo "    - \"chmod +x dist/authelia\""
  echo "    - \"authelia-scripts --log-level debug suites test ${SUITE_NAME} --headless\""
  echo "    label: \":selenium: ${SUITE_NAME} Suite\""
  echo "    env:"
  echo "      "CI: true""
  if [[ ${SUITE_NAME} == "Kubernetes" ]]; then
    echo "    agents:"
    echo "      "kubernetes: true""
  fi
done
