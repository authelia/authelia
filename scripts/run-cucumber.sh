#!/bin/bash

REQ=`for f in test/features/step_definitions/*.ts; do echo "--require $f"; done;`

./node_modules/.bin/cucumber-js --format-options '{"colorsEnabled": true}' --require-module ts-node/register --require test/features/support/world.ts $REQ $*
