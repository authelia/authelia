#!/bin/sh

set -x

go get github.com/cespare/reflex

# Sleep 10 seconds to wait the end of npm install updating web directory
# and making reflex reload multiple times.
sleep 10

reflex -c /resources/reflex.conf
