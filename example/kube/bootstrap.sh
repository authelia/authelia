#!/bin/bash

start_apps() {
  # Create the test application pages                                             
  kubectl create configmap app1-page --namespace=authelia --from-file=apps/app1/index.html
  kubectl create configmap app2-page --namespace=authelia --from-file=apps/app2/index.html
  kubectl create configmap app-home-page --namespace=authelia --from-file=apps/app-home/index.html
                                                                                  
  # Create TLS certificate and key for HTTPS termination                          
  kubectl create secret generic app1-tls --namespace=authelia --from-file=apps/app1/ssl/tls.key --from-file=apps/app1/ssl/tls.crt
  kubectl create secret generic app2-tls --namespace=authelia --from-file=apps/app2/ssl/tls.key --from-file=apps/app2/ssl/tls.crt
  kubectl create secret generic authelia-tls --namespace=authelia --from-file=authelia/ssl/tls.key --from-file=authelia/ssl/tls.crt
  
  # Spawn the applications
  kubectl apply -f apps
  kubectl apply -f apps/app1
  kubectl apply -f apps/app2
  kubectl apply -f apps/app-home
}

start_ingress_controller() {
  kubectl create configmap authelia-ingress-controller-config --namespace=authelia --from-file=ingress-controller/configs/nginx.tmpl
  kubectl apply -f ingress-controller
}

start_authelia() {
  kubectl create configmap authelia-config --namespace=authelia --from-file=authelia/configs/config.yml
  kubectl apply -f authelia
}

# Spawn Redis and Mongo as backend for Authelia
# Please note they are not configured to be distributed on several machines
start_storage() {
  kubectl apply -f storage
}

# Create a fake mailbox to catch emails sent by Authelia
start_mailcatcher() {
  kubectl apply -f mailcatcher
}

start_ldap() {
  kubectl apply -f ldap
}

# Create the Authelia namespace in the cluster
create_namespace() {
  kubectl apply -f namespace.yml
}

create_namespace
start_storage
start_ldap
start_mailcatcher
start_ingress_controller
start_authelia
start_apps
