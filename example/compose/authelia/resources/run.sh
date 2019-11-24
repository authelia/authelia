#!/bin/bash

set -e

# Retries a command on failure.
# $1 - the max number of attempts
# $2... - the command to run

retry() {
    local -r -i max_attempts="$1"; shift
    local -r cmd="$@"
    local -i attempt_num=1
    until $cmd
    do
        if ((attempt_num==max_attempts))
        then
            echo "Attempt $attempt_num failed and there are no more attempts left!"
            return 1
        else
            echo "Attempt $attempt_num failed! Trying again in 10 seconds..."
            sleep 10
        fi
    done
}


# Build the binary
go build -o /tmp/authelia/authelia-tmp cmd/authelia/main.go

retry 3 /tmp/authelia/authelia-tmp -config /etc/authelia/configuration.yml