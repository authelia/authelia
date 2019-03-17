#!/bin/bash

start_authelia() {
  kubectl create configmap authelia-config --namespace=authelia --from-file=authelia/configs/config.yml
  kubectl apply -f authelia
}

start_authelia