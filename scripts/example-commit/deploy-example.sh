#!/bin/bash

DC_SCRIPT=./scripts/example-commit/dc-example.sh

$DC_SCRIPT build
$DC_SCRIPT up -d mongo redis openldap authelia nginx
