#!/bin/bash
set -eu

for SUITE_NAME in BypassAll Docker DuoPush HighAvailability Kubernetes LDAP Mariadb NetworkACL Postgres ShortTimeouts Standalone Traefik; do
  echo "  - commands:"
  echo "    - \"xvfb-run authelia-scripts --log-level debug suites test ${SUITE_NAME} --headless\""
  echo "    label: \":selenium: ${SUITE_NAME}\""
done


