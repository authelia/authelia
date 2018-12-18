#!/bin/bash

build_and_push_authelia() {
  cd ../../
  docker build -t registry.kube.example.com:80/authelia .
  docker push registry.kube.example.com:80/authelia
}

build_and_push_authelia
