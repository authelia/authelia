#!/bin/bash

./scripts/dc-dev.sh up -d
./scripts/dc-dev.sh kill -s SIGHUP nginx-portal
