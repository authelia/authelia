#!/bin/bash
set -eu

echo "  - commands:"
echo "    - \"buildkite-agent artifact download "dist/*" .\""
echo "    - \"chmod +x dist/authelia\""
echo "    - \"authelia-scripts --log-level debug suites test Kubernetes --headless\""
echo "    label: \":selenium: Kubernetes Suite\""
echo "    agents:"
echo "      "kubernetes: true""
echo "    env:"
echo "      "CI: true""

for SUITE_NAME in BypassAll Docker DuoPush HighAvailability  LDAP Mariadb NetworkACL Postgres ShortTimeouts Standalone Traefik; do
  echo "  - commands:"
  echo "    - \"buildkite-agent artifact download "dist/*" .\""
  echo "    - \"chmod +x dist/authelia\""
  echo "    - \"authelia-scripts --log-level debug suites test ${SUITE_NAME} --headless\""
  echo "    label: \":selenium: ${SUITE_NAME} Suite\""
  echo "    env:"
  echo "      "CI: true""
done
