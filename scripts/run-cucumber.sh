#!/bin/bash

./node_modules/.bin/cucumber-js --colors --compiler ts:ts-node/register $*
