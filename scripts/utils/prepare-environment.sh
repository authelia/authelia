#!/bin/bash

bridge_exists=`docker network ls | grep " authelianet " | wc -l`

if [ "$bridge_exists" != "1" ];
then
  docker network create -d bridge --subnet 192.168.240.0/24 --gateway 192.168.240.1 authelianet
else
  echo "Bridge authelianet already exist."
fi

./scripts/dc-dev.sh up -d

./scripts/dc-dev.sh kill -s SIGHUP nginx-portal
