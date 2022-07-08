---
title: "Traefik Ingress"
description: "A guide to integrating Authelia with the Traefik Kubernetes Ingress."
lead: "A guide to integrating Authelia with the Traefik Kubernetes Ingress."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  integration:
    parent: "kubernetes"
weight: 550
toc: true
---

We officially support the Traefik 2.x Kubernetes ingress controllers. These come in two flavors:

* [Traefik Kubernetes Ingress](https://doc.traefik.io/traefik/providers/kubernetes-ingress/)
* [Traefik Kubernetes CRD](https://doc.traefik.io/traefik/providers/kubernetes-crd/)

The [Traefik documentation](../proxies/traefik.md) may also be useful for crafting advanced annotations to use with
this ingress even though it's not specific to Kubernetes.

## Get Started

It's __*strongly recommended*__ that users setting up *Authelia* for the first time take a look at our
[Get Started](../prologue/get-started.md) guide. This takes you through various steps which are essential to
bootstrapping *Authelia*.

## Special Notes

### Cross-Namespace Resources

Depending on your Traefik version you may be required to configure the
[allowCrossNamespace](https://doc.traefik.io/traefik/providers/kubernetes-crd/#allowcrossnamespace) to reuse a
[Middleware] from a namespace different to the Ingress or IngressRoute. Alternatively you can create the [Middleware] in
every namespace you need to use it.

## Middleware

Regardless if you're using the [Traefik Kubernetes Ingress] or purely the [Traefik Kubernetes CRD], you must configure
the [Traefik Kubernetes CRD] as far as we're aware at this time in order to configure a [ForwardAuth] [Middleware].

This is an example [Middleware] manifest. This eample assumes that you have deployed an Authelia pod and you have
configured it to be served on the URL `https://auth.example.com` and there is a Kubernetes Service with the name
`authelia` in the `default` namespace with TCP port `80` configured to route to the Authelia pod's HTTP port and that
your cluster is configured with the default DNS domain name of `cluster.local`.

{{< details "middleware.yml" >}}
```yaml
---
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  labels:
    app.kubernetes.io/instance: authelia
    app.kubernetes.io/name: authelia
    argocd.argoproj.io/instance: authelia
  name: forwardauth-authelia
  namespace: default
spec:
  forwardAuth:
    address: http://authelia.default.svc.cluster.local/api/verify?rd=https%3A%2F%2Fauth.example.com%2F
    authResponseHeaders:
      - Remote-User
      - Remote-Name
      - Remote-Email
      - Remote-Groups
...
```
{{< /details >}}

## Ingress

This is an example Ingress manifest which uses the above [Middleware](#middleware). This example assumes you have an
application you wish to serve on `https://app.example.com` and there is a Kubernetes Service with the name `app` in the
`default` namespace with TCP port `80` configured to route to the application pod's HTTP port.

{{< details "ingress.yml" >}}
```yaml
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: app
  namespace: default
  annotations:
    traefik.ingress.kubernetes.io/router.entrypoints: websecure
    traefik.ingress.kubernetes.io/router.middlewares: default-forwardauth-authelia@kubernetescrd
    traefik.ingress.kubernetes.io/router.tls: "true"
spec:
  rules:
    - host: app.example.com
      http:
        paths:
          - path: /bar
            pathType: Prefix
            backend:
              service:
                name:  app
                port:
                  number: 80
...
```
{{< /details >}}

## IngressRoute

This is an example IngressRoute manifest which uses the above [Middleware](#middleware). This example assumes you have an
application you wish to serve on `https://app.example.com` and there is a Kubernetes Service with the name `app` in the
`default` namespace with TCP port `80` configured to route to the application pod's HTTP port.

{{< details "ingressRoute.yml" >}}
```yaml
---
apiVersion: traefik.containo.us/v1alpha1
kind: IngressRoute
metadata:
  name: app
  namespace: default
spec:
  entryPoints:
    - websecure
  routes:
    - kind: Rule
      match: Host(`app.example.com`)
      middlewares:
        - name: forwardauth-authelia
          namespace: default
      services:
        - kind: Service
          name: app
          namespace: default
          port: 80
          scheme: http
          strategy: RoundRobin
          weight: 10
...
```
{{< /details >}}

[Traefik Kubernetes Ingress]: https://doc.traefik.io/traefik/providers/kubernetes-ingress/
[Traefik Kubernetes CRD]: https://doc.traefik.io/traefik/providers/kubernetes-crd/
[Middleware]: https://doc.traefik.io/traefik/middlewares/overview/
[ForwardAuth]: https://doc.traefik.io/traefik/middlewares/http/forwardauth/
