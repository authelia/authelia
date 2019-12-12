#!/bin/sh

set -e

while true;
do
    /app/dist/authelia --config /etc/authelia/configuration.yml
    sleep 10
done
