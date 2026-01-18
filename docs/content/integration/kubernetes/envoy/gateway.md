---
title: "Envoy Gateway"
description: "A guide to integrating Authelia with the Kubernetes Envoy Gateway."
summary: "A guide to integrating Authelia with the Kubernetes Envoy Gateway."
date: 2025-06-13T14:12:09+00:00
draft: false
images: []
menu:
integration:
parent: "kubernetes"
weight: 552
toc: true
aliases:
  - '/kubernetes/istio/'
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

[Envoy Gateway] is a [Gateway API] implementation. This means it has a relatively comprehensive integration option.
[Envoy Gateway] is supported with Authelia v4.37.0 and higher via the [Envoy] proxy [external authorization] filter.

In addition to this configuration, it's possible to configure the integration via OpenID Connect 1.0 which may be more
desirable when you wish to share an ID Token or Access Token with a backend. See that guide
[here](../../openid-connect/envoy-gateway/index.md).

The [Envoy Proxy documentation](../../proxies/envoy.md) may also be useful with this ingress even though it's not
specific to Kubernetes.

[external authorization]: https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/ext_authz/v3/ext_authz.proto.html#extensions-filters-http-ext-authz-v3-extauthz

## Get started

It's __*strongly recommended*__ that users setting up *Authelia* for the first time take a look at our
[Get started](../../prologue/get-started.md) guide. This takes you through various steps which are essential to
bootstrapping *Authelia*.

## Variables

Some of the values within this page can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Example

This example assumes that you have deployed an Authelia pod and you have configured it to be served on the URL
`https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}` and
there is a Kubernetes Service with the name `authelia` in the `default` namespace with TCP port `80` configured to route
to the Authelia pod's HTTP port and that your cluster is configured with the default DNS domain name of `cluster.local`.

### Security Policy

The following Security Policy examples assume the following:

1. Authelia is deployed to the `default` namespace.
2. AUthelia is deployed with a service named `authelia` which translates port `:80` to the Authelia server port.

#### Scoped to Gateway

This is an example [SecurityPolicy] manifest adjusted to authenticate with Authelia which is scoped to a single
[Gateway] named `eg`.

```yaml {title="istio-operator.yml"}
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: 'authelia-extauthz-by-gateway'
spec:
  targetRefs:
    - group: 'gateway.networking.k8s.io'
      kind: 'Gateway'
      name: 'eg'
  extAuth:
    headersToExtAuth:
      - 'accept'
      - 'cookie'
      - 'location'
      - 'authorization'
      - 'proxy-authorization'
      - 'x-forwarded-proto'
    failOpen: false
    http:
      backendRefs:
        - name: 'authelia'
          namespace: 'default'
          port: 80
      path: '/api/authz/ext-authz/'
      headersToBackend:
        - Remote-User
        - Remote-Groups
        - Remote-Name
        - Remote-Email
```

#### Scoped to HTTP Route

This is an example [SecurityPolicy] manifest adjusted to authenticate with Authelia which is scoped to a single
[HTTPRoute] named `example`.

```yaml {title="istio-operator.yml"}
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: 'authelia-extauthz-by-route'
spec:
  targetRefs:
    - group: 'gateway.networking.k8s.io'
      kind: 'HTTPRoute'
      name: 'example'
  extAuth:
    headersToExtAuth:
      - 'accept'
      - 'cookie'
      - 'authorization'
      - 'proxy-authorization'
      - 'x-forwarded-proto'
    failOpen: false
    http:
      backendRefs:
        - name: 'authelia'
          namespace: 'default'
          port: 80
      path: '/api/authz/ext-authz/'
      headersToBackend:
        - Remote-User
        - Remote-Groups
        - Remote-Name
        - Remote-Email
```

##### HTTP Route

The following [HTTPRoute] has the above [SecurityPolicy] applied to it for the
`app.{{< sitevar name="domain" nojs="example.com" >}}` domain:

```yaml {title="authoriztion-policy.yml"}
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: 'example'
spec:
  parentRefs:
    - name: 'eg'
  hostnames:
    - 'app.example.com'
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /
      backendRefs:
        - name: app
          port: 80
```

### Reference Grant

If the Gateway is deployed to a namespace different to Authelia you may need to apply a Reference grant to permit this
communication. The following example assumes the following:

1. Authelia is deployed to the `default` namespace.
2. You have deployed the [Scoped to Gateway](#scoped-to-gateway) Security Policy to the `eg` namespace.

```yaml
---
apiVersion: gateway.networking.k8s.io/v1beta1
kind: ReferenceGrant
metadata:
  name: 'example-ref-authelia-svc'
  namespace: 'default'
spec:
  from:
    - group: 'gateway.envoyproxy.io'
      kind: 'SecurityPolicy'
      namespace: 'eg'
      name: 'authelia-extauthz-by-gateway'
  to:
    - group: ''
      kind: 'Service'
      name: 'authelia'
```

## See Also

- Envoy Gateway [General](https://gateway.envoyproxy.io/docs/) Documentation
- Envoy Gateway [External Authorization Security Tasks](https://gateway.envoyproxy.io/docs/tasks/security/ext-auth/) Documentation
- Envoy Gateway [OIDC Authentication Security Tasks](https://gateway.envoyproxy.io/docs/tasks/security/oidc/) Documentation

[Envoy Gateway]: https://gateway.envoyproxy.io/
[Gateway API]: https://gateway-api.sigs.k8s.io/
[SecurityPolicy]: https://gateway.envoyproxy.io/contributions/design/security-policy/
[HTTPRoute]: https://gateway-api.sigs.k8s.io/api-types/httproute/
[Gateway]: https://gateway-api.sigs.k8s.io/api-types/gateway/
[Envoy]: https://www.envoyproxy.io/
