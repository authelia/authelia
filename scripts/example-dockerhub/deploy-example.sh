#!/bin/bash

DC_SCRIPT=./scripts/example-dockerhub/dc-example.sh

#$DC_SCRIPT build
$DC_SCRIPT up -d httpbin mongo redis openldap authelia smtp nginx-portal nginx-backend
