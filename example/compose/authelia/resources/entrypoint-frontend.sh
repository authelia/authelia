#!/bin/sh

set -x

if [ "$CI" == "true" ];
then
  echo "Use CI version of Authelia frontend"
  yarn start
else
  yarn install && yarn start
fi