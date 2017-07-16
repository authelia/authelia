#!/bin/bash

service_count=`docker ps -a | grep "Up " | wc -l`

if [ "${service_count}" -eq "5" ]
then
  echo "Service are up and running."
  exit 0
else
  echo "Some services exited..."
  docker ps -a
  exit 1
fi

