#!/bin/bash

kubectl -n kubernetes-dashboard describe secret $(kubectl -n kubernetes-dashboard get secret | grep admin-user | awk '{print $1}')
kubectl port-forward --address 0.0.0.0 -n kubernetes-dashboard service/kubernetes-dashboard 443:443