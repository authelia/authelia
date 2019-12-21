#!/bin/sh

set -x

if [ "$CI" == "true" ] && [ "$TRAVIS" != "true" ];
then
  echo "Use CI version of Authelia frontend"
  /resources/run-frontend.sh
else
  yarn install && yarn start
fi