#!/bin/sh

MODE=${1}

sed "s/__SUITE_SUBNET__/${SUITE_SUBNET:-192.168.240}/g" "/templates/${MODE}.conf" > /data/redis.conf
chown -R redis:redis /data

if [ "${MODE}" = "master" ] || [ "${MODE}" = "slave" ]; then
  redis-server /data/redis.conf
elif [ "${MODE}" = "sentinel" ]; then
  redis-server /data/redis.conf --sentinel
else
  echo "invalid argument: entrypoint.sh [master|slave|sentinel]"
  exit 1
fi
