#!/bin/sh

set -x

go get github.com/cespare/reflex

mkdir -p /var/lib/authelia
mkdir -p /etc/authelia

reflex -c /resources/reflex.conf