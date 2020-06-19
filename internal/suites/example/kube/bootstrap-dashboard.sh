#!/usr/bin/env bash

start_dashboard() {
  kubectl apply -f https://raw.githubusercontent.com/kubernetes/dashboard/v2.0.0-beta8/aio/deploy/recommended.yaml
  kubectl apply -f dashboard.yml

  echo "Bearer token for UI user."
  kubectl -n kubernetes-dashboard describe secret $(kubectl -n kubernetes-dashboard get secret | grep admin-user | awk '{print $1}')
}

start_dashboard
