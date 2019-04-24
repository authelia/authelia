#!/bin/bash

start_apps() {                                                                                  
  # Create TLS certificate and key for HTTPS termination                          
  kubectl create secret generic test-app-tls --namespace=authelia --from-file=apps/ssl/server.key --from-file=apps/ssl/server.cert
  
  # Spawn the applications
  kubectl apply -f apps
}

start_ingress_controller() {
  kubectl apply -f ingress-controller
}

start_dashboard() {
  kubectl apply -f https://raw.githubusercontent.com/kubernetes/dashboard/v2.0.0-beta4/aio/deploy/recommended.yaml
  kubectl apply -f dashboard.yml

  echo "Bearer token for UI user."
  kubectl -n kubernetes-dashboard describe secret $(kubectl -n kubernetes-dashboard get secret | grep admin-user | awk '{print $1}')
}

# Spawn Redis and Mongo as backend for Authelia
# Please note they are not configured to be distributed on several machines
start_storage() {
  kubectl apply -f storage
}

# Create a fake mailbox to catch emails sent by Authelia
start_mail() {
  kubectl apply -f mail
}

start_ldap() {
  kubectl create configmap ldap-config --namespace=authelia --from-file=ldap/base.ldif --from-file=ldap/access.rules
  kubectl apply -f ldap
}

# Create the Authelia namespace in the cluster
create_namespace() {
  kubectl apply -f namespace.yml
}

create_namespace
start_dashboard
start_storage
start_ldap
start_mail
start_ingress_controller
start_apps
