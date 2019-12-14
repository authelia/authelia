#!/bin/bash
set -eu

for FILES in authelia-linux-amd64.tar.gz authelia-linux-arm32v7.tar.gz authelia-linux-arm64v8.tar.gz authelia-linux-amd64.tar.gz.sha256 authelia-linux-arm32v7.tar.gz.sha256 authelia-linux-arm64v8.tar.gz.sha256;
do
  hub release create -a ${FILES} $BUILDKITE_TAG
done