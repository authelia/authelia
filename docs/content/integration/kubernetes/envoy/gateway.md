---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Envoy Gateway"
description: "A guide to integrating Authelia with the Kubernetes Envoy Gateway implementation using the external authorization filter and the Gateway API resource manifest."
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
[here](../../openid-connect/clients/envoy-gateway/index.md).

The [Envoy Proxy documentation](../../proxies/envoy.md) may also be useful with this ingress even though it's not
specific to Kubernetes.

[external authorization]: https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/ext_authz/v3/ext_authz.proto.html#extensions-filters-http-ext-authz-v3-extauthz

## Get started

It's **_strongly recommended_** that users setting up _Authelia_ for the first time take a look at our
[Get started](../../prologue/get-started.md) guide. This takes you through various steps which are essential to
bootstrapping _Authelia_.

## Variables

Some of the values within this page can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Example

This example assumes that you have deployed an Authelia pod and you have configured it to be served on the URL
`https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}` and
there is a Kubernetes Service with the name `authelia` in the `default` namespace with TCP port `80` configured to route
to the Authelia pod's HTTP port and that your cluster is configured with the default DNS domain name of `cluster.local`.

### Client Traffic Policy

Authelia uses the first address in the `X-Forwarded-For` header as the client address. [Envoy Gateway] appends the
address of the client to any `X-Forwarded-For` header sent by the client, which means the client can choose the first
address. The following [ClientTrafficPolicy] manifest removes the `X-Forwarded-For` header sent by the client before
[Envoy Gateway] adds the address of the client, which requires [Envoy Gateway] v1.2.0 or higher.

```yaml {title="client-traffic-policy.yml"}
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: ClientTrafficPolicy
metadata:
  name: 'authelia-client-address'
spec:
  targetRefs:
    - group: 'gateway.networking.k8s.io'
      kind: 'Gateway'
      name: 'eg'
  headers:
    earlyRequestHeaders:
      remove:
        - 'x-forwarded-for'
```

This assumes [Envoy Gateway] receives connections directly from clients, for example via a `LoadBalancer` Service with
`externalTrafficPolicy: Local`, or via the PROXY protocol using the `proxyProtocol` option of the
[ClientTrafficPolicy]. You should not configure the `clientIPDetection` option of the [ClientTrafficPolicy] with this
integration, as it causes [Envoy Gateway] to keep the `X-Forwarded-Proto` header sent by the client.

### Security Policy

The following Security Policy examples assume the following:

1. Authelia is deployed to the `default` namespace.
2. Authelia is deployed with a service named `authelia` which translates port `:80` to the Authelia server port.

#### Scoped to Gateway

This is an example [SecurityPolicy] manifest adjusted to authenticate with Authelia which is scoped to a single
[Gateway] named `eg`.

```yaml {title="security-policy-gateway.yml"}
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
      - 'authorization'
      - 'proxy-authorization'
      - 'x-forwarded-proto'
      - 'x-forwarded-for'
    failOpen: false
    http:
      backendRefs:
        - name: 'authelia'
          namespace: 'default'
          port: 80
      path: '/api/authz/ext-authz/'
      headersToBackend:
        - 'remote-user'
        - 'remote-groups'
        - 'remote-name'
        - 'remote-email'
```

#### Scoped to HTTP Route

This is an example [SecurityPolicy] manifest adjusted to authenticate with Authelia which is scoped to a single
[HTTPRoute] named `example`.

```yaml {title="security-policy-route.yml"}
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
      - 'x-forwarded-for'
    failOpen: false
    http:
      backendRefs:
        - name: 'authelia'
          namespace: 'default'
          port: 80
      path: '/api/authz/ext-authz/'
      headersToBackend:
        - 'remote-user'
        - 'remote-groups'
        - 'remote-name'
        - 'remote-email'
```

##### HTTP Route

The following [HTTPRoute] has the above [SecurityPolicy] applied to it for the
`app.{{< sitevar name="domain" nojs="example.com" >}}` domain:

```yaml {title="httproute.yml"}
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: 'example'
spec:
  parentRefs:
    - name: 'eg'
  hostnames:
    - 'app.{{< sitevar name="domain" nojs="example.com" >}}'
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

```yaml {title="reference-grant.yml"}
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
[ClientTrafficPolicy]: https://gateway.envoyproxy.io/docs/api/extension_types/#clienttrafficpolicy
[HTTPRoute]: https://gateway-api.sigs.k8s.io/api-types/httproute/
[Gateway]: https://gateway-api.sigs.k8s.io/api-types/gateway/
[Envoy]: https://www.envoyproxy.io/
