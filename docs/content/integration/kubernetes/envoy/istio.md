---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Istio"
description: "A guide to integrating Authelia with the Istio Kubernetes Ingress which uses Envoy and the external authorization filter for authentication and authorization."
summary: "A guide to integrating Authelia with the Istio Kubernetes Ingress."
date: 2025-06-13T14:12:09+00:00
draft: false
images: []
menu:
integration:
parent: "kubernetes"
weight: 553
toc: true
aliases:
  - '/kubernetes/istio/'
  - '/integration/kubernetes/istio/'
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

Istio uses [Envoy] as an Ingress. This means it has a relatively comprehensive integration option.
Istio is supported with Authelia v4.37.0 and higher via the [Envoy] proxy [external authorization] filter.
The example uses the [Authz Endpoints](../../../configuration/miscellaneous/server-endpoints-authz.md) which require
Authelia v4.38.0 or higher.

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
`https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}` and there is a Kubernetes Service with the name `authelia` in the `default` namespace with
TCP port `80` configured to route to the Authelia pod's HTTP port and that your cluster is configured with the default
DNS domain name of `cluster.local`.

### Mesh Configuration

The Authelia extension provider is configured in the [MeshConfig] of the mesh. The following examples only show the
necessary portions of the configuration that you add as well as context. You will need to adapt them to your needs.

The `includeAdditionalHeadersInCheck` values send the trusted client address as the `X-Forwarded-For` header and the
scheme determined by [Envoy] as the `X-Forwarded-Proto` header in the authorization request. [Envoy] does not include
the `X-Forwarded-For` header in the authorization request unless it's configured, and Authelia uses the first address
in this header as the client address, so you should not add `x-forwarded-for` to `includeRequestHeadersInCheck` as the
first address can be set by the client.

#### Helm

If you installed Istio using Helm, add the following to the values of the `istiod` chart:

```yaml {title="values.yml"}
meshConfig:
  extensionProviders:
    - name: 'authelia'
      envoyExtAuthzHttp:
        service: 'authelia.default.svc.cluster.local'
        port: 80
        pathPrefix: '/api/authz/ext-authz/'
        includeRequestHeadersInCheck:
          - 'accept'
          - 'cookie'
          - 'authorization'
          - 'proxy-authorization'
        headersToUpstreamOnAllow:
          - 'remote-*'
          - 'authelia-*'
        includeAdditionalHeadersInCheck:
          X-Forwarded-For: '%DOWNSTREAM_REMOTE_ADDRESS_WITHOUT_PORT%'
          X-Forwarded-Proto: '%REQ(X-FORWARDED-PROTO)%'
        headersToDownstreamOnDeny:
          - 'set-cookie'
        headersToDownstreamOnAllow:
          - 'set-cookie'
```

#### istioctl

If you installed Istio using `istioctl`, add the following to the IstioOperator manifest used with `istioctl install`:

```yaml {title="istio-operator.yml"}
apiVersion: 'install.istio.io/v1alpha1'
kind: 'IstioOperator'
spec:
  meshConfig:
    extensionProviders:
      - name: 'authelia'
        envoyExtAuthzHttp:
          service: 'authelia.default.svc.cluster.local'
          port: 80
          pathPrefix: '/api/authz/ext-authz/'
          includeRequestHeadersInCheck:
            - 'accept'
            - 'cookie'
            - 'authorization'
            - 'proxy-authorization'
          headersToUpstreamOnAllow:
            - 'remote-*'
            - 'authelia-*'
          includeAdditionalHeadersInCheck:
            X-Forwarded-For: '%DOWNSTREAM_REMOTE_ADDRESS_WITHOUT_PORT%'
            X-Forwarded-Proto: '%REQ(X-FORWARDED-PROTO)%'
          headersToDownstreamOnDeny:
            - 'set-cookie'
          headersToDownstreamOnAllow:
            - 'set-cookie'
```

### Authorization Policy

The following [Authorization Policy] applies the above filter extension provider to requests for the
`app.{{< sitevar name="domain" nojs="example.com" >}}` domain which pass through the Istio ingress gateway. An
[Authorization Policy] applies to the workloads matched by its `selector` in the namespace it is created in, or in all
namespaces when it is created in the root namespace which is `istio-system` by default. This example assumes the
ingress gateway is deployed with the `app: istio-ingressgateway` label, and you will need to adjust the `selector` if
this is not the case.

The policy should be applied to the ingress gateway. If it's applied to the sidecar of an application instead, the
client address Authelia receives is the address of the ingress gateway rather than the address of the client.

```yaml {title="authorization-policy.yml"}
apiVersion: 'security.istio.io/v1'
kind: 'AuthorizationPolicy'
metadata:
  name: 'example'
  namespace: 'istio-system'
spec:
  selector:
    matchLabels:
      app: 'istio-ingressgateway'
  action: 'CUSTOM'
  provider:
    name:  'authelia'
  rules:
    - to:
        - operation:
            hosts:
              - 'app.{{< sitevar name="domain" nojs="example.com" >}}'
```

## FAQ

### The Istio ingress gateway is behind another proxy which terminates TLS, how do I configure that?

By default the Istio ingress gateway does not trust the `X-Forwarded-For` and `X-Forwarded-Proto` headers sent to it. If
another proxy such as a cloud load balancer terminates TLS and sends requests to the ingress gateway over plain HTTP,
the `X-Forwarded-Proto` header is replaced with `http` and the client address is the address of that proxy. Authelia
then rejects the requests with a `400 Bad Request` status code and logs an error stating the Target URL has an insecure
scheme `http`.

To resolve this set `numTrustedProxies` to the number of trusted proxies in front of the ingress gateway, each of which
must append to the `X-Forwarded-For` header, as described in the Istio [Network Topologies] documentation. For example,
using the [MeshConfig]:

```yaml {title="values.yml"}
meshConfig:
  defaultConfig:
    gatewayTopology:
      numTrustedProxies: 1
```

You must restart the ingress gateway pods after changing this option. This option trusts these headers based on the
number of hops rather than the address of the proxy, so the ingress gateway should only be reachable via the trusted
proxies.

## See Also

- Istio [External Authentication](https://istio.io/latest/docs/tasks/security/authorization/authz-custom/) Documentation
- Istio [Authorization Policy] Documentation
- Istio [IstioOperator Options](https://istio.io/latest/docs/reference/config/istio.operator.v1alpha1/) Documentation
- Istio [Helm Installation](https://istio.io/latest/docs/setup/install/helm/) Documentation
- Istio [MeshConfig Extension Provider EnvoyExtAuthz HTTP Provider](https://istio.io/latest/docs/reference/config/istio.mesh.v1alpha1/#MeshConfig-ExtensionProvider-EnvoyExternalAuthorizationHttpProvider) Documentation

[Authorization Policy]: https://istio.io/latest/docs/reference/config/security/authorization-policy/
[MeshConfig]: https://istio.io/latest/docs/reference/config/istio.mesh.v1alpha1/
[Network Topologies]: https://istio.io/latest/docs/ops/configuration/traffic-management/network-topologies/
[Envoy]: https://www.envoyproxy.io/
