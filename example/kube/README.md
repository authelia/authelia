# Authelia on Kubernetes

Authelia is now available on Kube in order to protect your most critical
applications using 2-factor authentication and Single Sign-On.

This example leverages [ingress-nginx](https://github.com/kubernetes/ingress-nginx)
to delegate authentication and authorization to Authelia within the cluster.

## Getting started

You can either try to install **Authelia** on your running instance of Kubernetes
or deploy the dedicated [suite](/docs/suites.md) called *kubernetes*.

### Set up a Kube cluster

The simplest way to start a Kubernetes cluster is to deploy the *kubernetes* suite with

    authelia-scripts suites start kubernetes

This will take a few seconds (or minutes) to deploy the cluster.

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
authentication portal so that the user can be redirected if he is not
yet authenticated. This annotation is as follows:
`nginx.ingress.kubernetes.io/auth-signin: "https://login.example.com:8080/#/"`

Those annotations can be seen in `apps/apps.yml` configuration.

### Production grade infrastructure

What is great with using [ingress-nginx](https://github.com/kubernetes/ingress-nginx)
is that it is compatible with [kube-lego](https://github.com/jetstack/kube-lego)
which removes the usual pain of manually renewing SSL certificates. It uses
letsencrypt to issue and renew certificates every three month without any
manual intervention.

## What do I need to know to deploy it in my cluster?

Given your cluster already runs a LDAP server, a Redis, a Mongo database,
a SMTP server and a nginx ingress-controller, you can deploy **Authelia**
and update your ingress configurations. An example is provided 
[here](./authelia).

## Questions

If you have questions about the implementation, please post them on
[![Gitter](https://img.shields.io/gitter/room/badges/shields.svg)](https://gitter.im/authelia/general?utm_source=share-link&utm_medium=link&utm_campaign=share-link)
