#!/bin/sh

set -x

if [ "$CI" == "true" ];
then
  echo "Use CI version of Authelia"
  /resources/run-backend.sh
else
  echo "Use hot reloaded version of Authelia backend"
  go get github.com/cespare/reflex

  # Sleep 10 seconds to wait the end of npm install updating web directory
  # and making reflex reload multiple times.
  sleep 10

  reflex -c /resources/reflex.conf
fi