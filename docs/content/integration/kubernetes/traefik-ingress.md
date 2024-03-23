---
title: "Traefik Ingress"
description: "A guide to integrating Authelia with the Traefik Kubernetes Ingress."
summary: "A guide to integrating Authelia with the Traefik Kubernetes Ingress."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
weight: 550
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

We officially support the Traefik 2.x Kubernetes ingress controllers. These come in two flavors:

* [Traefik Kubernetes Ingress](https://doc.traefik.io/traefik/providers/kubernetes-ingress/)
* [Traefik Kubernetes CRD](https://doc.traefik.io/traefik/providers/kubernetes-crd/)

The [Traefik documentation](../proxies/traefik.md) may also be useful for crafting advanced annotations to use with
this ingress even though it's not specific to Kubernetes.

## Get started

It's __*strongly recommended*__ that users setting up *Authelia* for the first time take a look at our
[Get started](../prologue/get-started.md) guide. This takes you through various steps which are essential to
bootstrapping *Authelia*.

## Special Notes

### Cross-Namespace Resources

Depending on your Traefik version you may be required to configure the
[allowCrossNamespace](https://doc.traefik.io/traefik/providers/kubernetes-crd/#allowcrossnamespace) to reuse a
[Middleware] from a [Namespace] different to the [Ingress] / [IngressRoute]. Alternatively you can create the [Middleware]
in every [Namespace] you need to use it.

## Middleware

Regardless if you're using the [Traefik Kubernetes Ingress] or purely the [Traefik Kubernetes CRD], you must configure
the [Traefik Kubernetes CRD] as far as we're aware at this time in order to configure a [ForwardAuth] [Middleware].

This is an example [Middleware] manifest. This example assumes that you have deployed an Authelia [Pod] and you have
configured it to be served on the URL `https://auth.example.com` and there is a Kubernetes [Service] with the name
`authelia` in the `default` [Namespace] with TCP port `80` configured to route to the Authelia [Pod]'s HTTP port and
that your cluster is configured with the default DNS domain name of `cluster.local`.

*__Important Note:__ The [Middleware] should be applied to an [Ingress] / [IngressRoute] you wish to protect. It
__SHOULD NOT__ be applied to the Authelia [Ingress] / [IngressRoute] itself.*

{{< details "middleware.yml" >}}
```yaml
---
apiVersion: 'traefik.containo.us/v1alpha1'
kind: 'Middleware'
metadata:
  name: 'forwardauth-authelia' # name of middleware as it appears in Traefik, and how you reference in ingress rules
  namespace: 'default' # name of namespace that Traefik is in
  labels:
    app.kubernetes.io/instance: 'authelia'
    app.kubernetes.io/name: 'authelia'
spec:
  forwardAuth:
    address: 'http://authelia.default.svc.cluster.local/api/authz/forward-auth'
    authResponseHeaders:
      - 'Remote-User'
      - 'Remote-Groups'
      - 'Remote-Email'
      - 'Remote-Name'
...
```
{{< /details >}}

## Ingress

This is an example [Ingress] manifest which uses the above [Middleware](#middleware). This example assumes you have an
application you wish to serve on `https://app.example.com` and there is a Kubernetes [Service] with the name `app` in
the `default` [Namespace] with TCP port `80` configured to route to the application [Pod]'s HTTP port.

{{< details "ingress.yml" >}}
```yaml
---
apiVersion: 'networking.k8s.io/v1'
kind: 'Ingress'
metadata:
  name: 'app'
  namespace: 'default'
  annotations:
    traefik.ingress.kubernetes.io/router.entryPoints: 'websecure' # name of your https entry point (default is 'websecure')
    traefik.ingress.kubernetes.io/router.middlewares: 'default-forwardauth-authelia@kubernetescrd' # name of your middleware, as defined in your middleware.yaml
    traefik.ingress.kubernetes.io/router.tls: 'true'
spec:
  rules:
    - host: 'app.example.com'
      http:
        paths:
          - path: '/bar'
            pathType: 'Prefix'
            backend:
              service:
                name:  'app'
                port:
                  number: 80
...
```
{{< /details >}}

## IngressRoute

This is an example [IngressRoute] manifest which uses the above [Middleware](#middleware). This example assumes you have
an application you wish to serve on `https://app.example.com` and there is a Kubernetes [Service] with the name `app` in
the `default` [Namespace] with TCP port `80` configured to route to the application [Pod]'s HTTP port.

{{< details "ingressRoute.yml" >}}
```yaml
---
apiVersion: 'traefik.containo.us/v1alpha1'
kind: 'IngressRoute'
metadata:
  name: 'app'
  namespace: 'default'
spec:
  entryPoints:
    - 'websecure'  # name of your https entry point (default is 'websecure')
  routes:
    - kind: 'Rule'
      match: 'Host(`app.example.com`)'
      middlewares:
        - name: 'forwardauth-authelia' # name of your middleware, as defined in your middleware.yaml
          namespace: 'default'
      services:
        - kind: 'Service'
          name: 'app'
          namespace: 'default'
          port: 80
          scheme: 'http'
          strategy: 'RoundRobin'
          weight: 10
...
```
{{< /details >}}

[Namespace]: https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/
[Pod]: https://kubernetes.io/docs/concepts/workloads/pods/
[Service]: https://kubernetes.io/docs/concepts/services-networking/service/
[IngressRoute]: https://doc.traefik.io/traefik/providers/kubernetes-crd/
[Ingress]: https://kubernetes.io/docs/concepts/services-networking/ingress/
[Traefik Kubernetes Ingress]: https://doc.traefik.io/traefik/providers/kubernetes-ingress/
[Traefik Kubernetes CRD]: https://doc.traefik.io/traefik/providers/kubernetes-crd/
[Middleware]: https://doc.traefik.io/traefik/middlewares/overview/
[ForwardAuth]: https://doc.traefik.io/traefik/middlewares/http/forwardauth/
