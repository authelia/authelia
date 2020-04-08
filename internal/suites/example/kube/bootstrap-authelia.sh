#!/bin/sh

start_authelia() {
  kubectl create configmap authelia-config --namespace=authelia --from-file=authelia/configs/configuration.yml
  kubectl apply -f authelia
}

start_authelia