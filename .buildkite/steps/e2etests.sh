#!/bin/bash
set -eu

for SUITE_NAME in BypassAll Docker DuoPush HighAvailability Kubernetes LDAP Mariadb NetworkACL Postgres ShortTimeouts Standalone Traefik;
do
  echo "  - commands:"
  echo "    - \"authelia-scripts --log-level debug suites test ${SUITE_NAME} --headless\""
  echo "    label: \":selenium: ${SUITE_NAME} Suite\""
  if [[ "${SUITE_NAME}" != "Kubernetes" ]];
  then
    echo "    agents:"
    echo "      "suite: all""
  else
    echo "    agents:"
    echo "      "suite: kubernetes""
  fi
  echo "    plugins:"
  echo "      - \"nightah/github-checks#v0.0.4\""
done