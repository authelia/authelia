#!/bin/sh

set -x

pnpm install --force --frozen-lockfile && pnpm start
