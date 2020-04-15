#!/bin/sh

start_authelia() {
  kubectl create configmap authelia-config --namespace=authelia --from-file=authelia/configs/configuration.yml
  kubectl create configmap authelia-ssl --namespace=authelia --from-file=authelia/ssl/cert.pem --from-file=authelia/ssl/key.pem
  kubectl apply -f authelia
}

start_authelia