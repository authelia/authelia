---
title: "NGINX Ingress"
description: "A guide to integrating Authelia with the NGINX Kubernetes Ingress."
lead: "A guide to integrating Authelia with the NGINX Kubernetes Ingress."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  integration:
    parent: "kubernetes"
weight: 551
toc: true
---

There are two nginx ingress controllers for Kubernetes. The Kubernetes official one [ingress-nginx], and the F5 nginx
official one [nginx-ingress-controller]. Currently we only have support docs for [ingress-nginx].

The [nginx documentation](../proxies/nginx.md) may also be useful for crafting advanced snippets to use with annotations
even though it's not specific to Kubernetes.

## Get Started

It's __*strongly recommended*__ that users setting up *Authelia* for the first time take a look at our
[Get Started](../prologue/get-started.md) guide. This takes you through various steps which are essential to
bootstrapping *Authelia*.

## NGINX Ingress Controller (ingress-nginx)

If you use NGINX Ingress Controller (ingress-nginx) you can protect an ingress with the following annotations. The
example assumes that the public domain Authelia is served on is `https://auth.example.com` and there is a
Kubernetes service with the name `authelia` in the `default` namespace with TCP port `80` configured to route to the
Authelia HTTP port and that your cluster is configured with the default
DNS domain name of `cluster.local`.

### Ingress Annotations

```yaml
annotations:
  nginx.ingress.kubernetes.io/auth-response-headers: Remote-User,Remote-Name,Remote-Groups,Remote-Email
  nginx.ingress.kubernetes.io/auth-signin: https://auth.example.com
  nginx.ingress.kubernetes.io/auth-snippet: |
    proxy_set_header X-Forwarded-Method $request_method;
  nginx.ingress.kubernetes.io/auth-url: http://authelia.default.svc.cluster.local/api/verify
```

[ingress-nginx]: https://kubernetes.github.io/ingress-nginx/
[nginx-ingress-controller]: https://docs.nginx.com/nginx-ingress-controller/
