#!/bin/sh

set -x

if [ "$CI" == "true" ];
then
    echo "Use CI version of Authelia frontend"
    /resources/run-frontend.sh
else
    npm ci && npm run start
fi