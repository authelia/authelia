#!/bin/sh
set -e

MISSING=""
COUNT=0

for TOOL in goimports-reviser golangci-lint pnpm shellcheck trufflehog typos yamllint zizmor; do
  if ! command -v "${TOOL}" >/dev/null 2>&1; then
    if [ "${COUNT}" -eq 0 ]; then
      MISSING=${TOOL}
    elif [ "${COUNT}" -eq 1 ]; then
      MISSING="${MISSING} and ${TOOL}"
    else
      MISSING="$(echo "${MISSING}" | sed 's/ and /, /') and ${TOOL}"
    fi
    COUNT=$((COUNT + 1))
  fi
done

if [ "${COUNT}" -gt 0 ]; then
  echo "❌ You must install ${MISSING}"
  exit 1
fi
