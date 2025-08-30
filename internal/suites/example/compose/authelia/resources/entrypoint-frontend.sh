#!/bin/sh

PNPM_MODULE="./node_modules/.modules.yaml"

if [[ -f "${PNPM_MODULE}" ]]; then
  rm "${PNPM_MODULE}"
fi

pnpm install --ignore-scripts --frozen-lockfile && pnpm start
