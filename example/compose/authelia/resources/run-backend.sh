#!/bin/sh

set -e

while /app/dist/authelia --config /etc/authelia/configuration.yml; [ $? -ne 0 ];
do
  echo "Waiting on services for Authelia"
done