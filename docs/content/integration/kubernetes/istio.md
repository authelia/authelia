---
title: "Istio"
description: "A guide to integrating Authelia with the Istio Kubernetes Ingress."
summary: "A guide to integrating Authelia with the Istio Kubernetes Ingress."
date: 2022-10-02T13:59:09+11:00
draft: false
images: []
menu:
integration:
parent: "kubernetes"
weight: 551
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

Istio uses [Envoy](../proxies/envoy.md) as an Ingress. This means it has a relatively comprehensive integration option.
Istio is supported with Authelia v4.37.0 and higher via the [Envoy] proxy [external authorization] filter.

[external authorization]: https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/ext_authz/v3/ext_authz.proto.html#extensions-filters-http-ext-authz-v3-extauthz

## Example

This example assumes that you have deployed an Authelia pod and you have configured it to be served on the URL
`https://auth.example.com` and there is a Kubernetes Service with the name `authelia` in the `default` namespace with
TCP port `80` configured to route to the Authelia pod's HTTP port and that your cluster is configured with the default
DNS domain name of `cluster.local`.

### Operator

This is an example IstioOperator manifest adjusted to authenticate with Authelia. This example only shows the necessary
portions of the resource that you add as well as context. You will need to adapt it to your needs.

```yaml
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
            X-Forwarded-Proto: '%REQ(:SCHEME)%'
          headersToDownstreamOnDeny:
            - 'set-cookie'
          headersToDownstreamOnAllow:
            - 'set-cookie'
```

### Authorization Policy

The following [Authorization Policy] applies the above filter extension provider to the `nextcloud.example.com` domain:

```yaml
apiVersion: 'security.istio.io/v1beta1'
kind: 'AuthorizationPolicy'
metadata:
  name: 'nextcloud'
  namespace: 'apps'
spec:
  action: 'CUSTOM'
  provider:
    name:  'authelia'
  rules:
    - to:
        - operation:
            hosts:
              - 'nextcloud.example.com'
```

## See Also

- Istio [External Authentication](https://istio.io/latest/docs/tasks/security/authorization/authz-custom/) Documentation
- Istio [Authorization Policy] Documentation
- Istio [IstioOperator Options](https://istio.io/latest/docs/reference/config/istio.operator.v1alpha1/) Documentation
- Istio [MeshConfig Extension Provider EnvoyExtAuthz HTTP Provider](https://istio.io/latest/docs/reference/config/istio.mesh.v1alpha1/#MeshConfig-ExtensionProvider-EnvoyExternalAuthorizationHttpProvider) Documentation

[Authorization Policy]: https://istio.io/latest/docs/reference/config/security/authorization-policy/
