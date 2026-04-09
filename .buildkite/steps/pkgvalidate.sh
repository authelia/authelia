#!/usr/bin/env bash

case "${1}" in
  docs)
    pnpm -C docs install && pnpm -C docs build
    ;;
  templates)
    pnpm -C internal/templates/src install && pnpm -C internal/templates/src render
    ;;
  *)
    echo "Usage: pkgvalidate.sh [docs|templates]"
    exit 1
    ;;
esac
