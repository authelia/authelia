# Authelia on Kubernetes

Authelia is now available on Kube in order to protect your most critical
applications using 2-factor authentication and Single Sign-On.

This example leverages [ingress-nginx](https://github.com/kubernetes/ingress-nginx)
v0.13.0 to delegate authentications and authorizations to Authelia within
the cluster.

## Getting started

In order to deploy Authelia on Kube, you must have a cluster at hand. If you
don't, please follow the next section otherwise skip it and go
to the next.

### Set up a Kube cluster

Hopefully, spawning a development cluster from scratch has become very
easy lately with the use of **minikube**. This project creates a VM on your
computer and start a Kube cluster inside it. It also configure a CLI called
kubectl so that you can deploy applications in the cluster right away.

Basically, you need to follow the instruction from the [repository](https://github.com/kubernetes/minikube).
It should be a matter of downloading the binary and start the cluster with
two commands:

```
curl -Lo minikube https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64 && chmod +x minikube && sudo mv minikube /usr/local/bin/
minikube start # you can use --vm-driver flag for selecting your hypervisor (virtualbox by default otherwise)
```

After few seconds, your cluster should be working and you should be able to
get access to the cluster by creating a proxy with

```
kubectl proxy
```

and visiting `http://localhost:8001/ui`

### Deploy Authelia

Once the cluster is ready and you can access it, run the following command to
deploy Authelia:

```
./bootstrap.sh
```

In order to visit the test applications that have been deployed to test
Authelia, edit your /etc/hosts and add the following lines replacing the IP
with the IP of your VM given by minikube:

```
192.168.39.26       login.kube.example.com
192.168.39.26       app1.kube.example.com
192.168.39.26       app2.kube.example.com
192.168.39.26       mail.kube.example.com
192.168.39.26       home.kube.example.com

# The domain of the private docker registry holding dev version of Authelia
192.168.39.26       registry.kube.example.com
```

Once done, you can visit http://home.kube.example.com and follow the
instructions written in the page.

## How does it work?

### Authentication via Authelia

In a Kube clusters, the routing logic of requests is handled by ingress
controllers following rules provided by ingress configurations.

In this example, [ingress-nginx](https://github.com/kubernetes/ingress-nginx)
controller has been installed to handle the incoming requests. Some of them
(specified in the ingress configuration) are forwarded to Authelia so that
it can verify whether they are allowed and should reach the protected endpoint.

The authentication is provided at the ingress level by an annotation called
`nginx.ingress.kubernetes.io/auth-url` that is filled with the URL of
Authelia's verification endpoint.
The ingress controller also requires the URL to the
authentication portal so that the user can be redirected in case she is not
yet authenticated.

Those annotations can be seen in `apps/secure-ingress.yml` configuration.

### Production grade infrastructure

What is great with using [ingress-nginx](https://github.com/kubernetes/ingress-nginx)
is that it is compatible with [kube-lego](https://github.com/jetstack/kube-lego)
which removes the usual pain of manual SSL certificate renewals. It uses
letsencrypt to issue and renew certificates every three month without any
manual intervention.

## What do I need know to deploy it in my cluster?

Given your cluster is already made of an LDAP server, a Redis cluster, a Mongo
cluster and a SMTP server, you'll only need to install the ingress-controller
and Authelia whose kubernetes deployment configurations are respectively in 
`ingress-controller` and `authelia` directories. A template configuration
is provided there, you just need to create the configmap to use it within
the cluster.

### I'm already using ingress-nginx

If you're already using ingress-nginx as your ingress controller, you only
need to install Authelia with its configuration and that's it!

## Questions

If you have questions about the implementation, please post them on
[![Gitter](https://img.shields.io/gitter/room/badges/shields.svg)](https://gitter.im/authelia/general?utm_source=share-link&utm_medium=link&utm_campaign=share-link)
