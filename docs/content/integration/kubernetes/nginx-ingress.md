---
title: "NGINX Ingress"
description: "A guide to integrating Authelia with the NGINX Kubernetes Ingress."
summary: "A guide to integrating Authelia with the NGINX Kubernetes Ingress."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
weight: 552
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

There are two nginx ingress controllers for Kubernetes. The Kubernetes official one [ingress-nginx], and the F5 nginx
official one [nginx-ingress-controller]. We only have integration documentation for [ingress-nginx] and there are no
plans to support the F5 [nginx-ingress-controller].

The [nginx documentation](../proxies/nginx.md) may also be useful for crafting advanced snippets to use with annotations
even though it's not specific to Kubernetes.

## Get started

It's __*strongly recommended*__ that users setting up *Authelia* for the first time take a look at our
[Get started](../prologue/get-started.md) guide. This takes you through various steps which are essential to
bootstrapping *Authelia*.

## NGINX Ingress Controller (ingress-nginx)

If you use NGINX Ingress Controller ([ingress-nginx]) you can protect an ingress with the following annotations. The
example assumes that the public domain Authelia is served on is `https://auth.example.com` and there is a
Kubernetes service with the name `authelia` in the `default` namespace with TCP port `80` configured to route to the
Authelia HTTP port and that your cluster is configured with the default
DNS domain name of `cluster.local`.

*__Important Note:__ The following annotations should be applied to an Ingress you wish to protect. They __SHOULD NOT__
be applied to the Authelia Ingress itself.*

### Ingress Annotations

```yaml
annotations:
  nginx.ingress.kubernetes.io/auth-method: 'GET'
  nginx.ingress.kubernetes.io/auth-url: 'http://authelia.default.svc.cluster.local/api/authz/auth-request'
  nginx.ingress.kubernetes.io/auth-signin: 'https://auth.example.com?rm=$request_method'
  nginx.ingress.kubernetes.io/auth-response-headers: 'Remote-User,Remote-Name,Remote-Groups,Remote-Email'
```

[ingress-nginx]: https://kubernetes.github.io/ingress-nginx/
[nginx-ingress-controller]: https://docs.nginx.com/nginx-ingress-controller/
