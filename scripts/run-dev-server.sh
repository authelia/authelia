#!/bin/sh

# Starts the server with the provided configuration in $1
# This scripts is called from authelia-scripts.

./node_modules/.bin/ts-node -P ./server/tsconfig.json ./server/src/index.ts $*
