#!/bin/sh

./node_modules/.bin/ts-node -P ./server/tsconfig.json ./server/src/index.ts $*
