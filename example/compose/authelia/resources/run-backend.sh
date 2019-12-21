#!/bin/sh

set -e

#TODO(nightah): Remove when turning off Travis
if [ "$CI" == "true" ] && [ "$TRAVIS" == "true" ];
then
  go build -o /app/dist/authelia cmd/authelia/*.go
fi
#TODO(nightah): Remove when turning off Travis

while /app/dist/authelia --config /etc/authelia/configuration.yml; [ $? -ne 0 ];
do
  echo "Waiting on services for Authelia"
done