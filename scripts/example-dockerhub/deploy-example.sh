#!/bin/bash

DC_SCRIPT=./scripts/example-dockerhub/dc-example.sh

#$DC_SCRIPT build
$DC_SCRIPT up -d mongo redis openldap authelia nginx
