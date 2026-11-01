#!/usr/bin/env sh

# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

PNPM_MODULE="./node_modules/.modules.yaml"

if [ -f "${PNPM_MODULE}" ]; then
  rm "${PNPM_MODULE}"
fi

pnpm install --ignore-scripts --frozen-lockfile && pnpm start
