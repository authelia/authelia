#!/bin/bash
set -eu

for SUITE_NAME in BypassAll Docker DuoPush HighAvailability Kubernetes LDAP Mariadb NetworkACL Postgres ShortTimeouts Standalone Traefik;
do
  echo "  - label: \":selenium: ${SUITE_NAME} Suite\""
  echo "    commands:"
  echo "      - \"authelia-scripts --log-level debug suites test ${SUITE_NAME} --headless\""
  if [[ "${SUITE_NAME}" != "Kubernetes" ]];
  then
    echo "    agents:"
    echo "      "suite: all""
  else
    echo "    agents:"
    echo "      "suite: kubernetes""
  fi
done