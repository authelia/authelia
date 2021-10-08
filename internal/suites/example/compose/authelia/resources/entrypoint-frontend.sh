#!/bin/sh

set -x

pnpm install --shamefully-hoist --frozen-lockfile && pnpm start