#!/bin/sh

# SPDX-FileCopyrightText: 2019 Authelia
#
# SPDX-License-Identifier: Apache-2.0

set -x

pnpm install --frozen-lockfile && pnpm start