---
title: "Istio"
description: "A guide to integrating Authelia with the Istio Kubernetes Ingress."
lead: "A guide to integrating Authelia with the Istio Kubernetes Ingress."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
integration:
parent: "kubernetes"
weight: 551
toc: true
---

Istio uses [Envoy](../proxies/envoy.md) as an Ingress. This means it has a relatively comprehensive integration option.

## Example

This example assumes that you have deployed an Authelia pod and you have configured it to be served on the URL
`https://auth.example.com` and there is a Kubernetes Service with the name `authelia` in the `default` namespace with
TCP port `80` configured to route to the Authelia pod's HTTP port and that your cluster is configured with the default
DNS domain name of `cluster.local`.

### Operator

This is an example IstioOperator manifest adjusted to authenticate with Authelia. This example only shows the necessary
portions of the resource that you add as well as context. You will need to adapt it to your needs.

```yaml
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
spec:
  meshConfig:
    extensionProviders:
    - name: "authelia"
      envoyExtAuthzHttp:
      service: "authelia.default.svc.cluster.local"
      port: 80
      pathPrefix: "/api/verify"
      includeRequestHeadersInCheck:
      - cookie
      - proxy-authorization
      headersToUpstreamOnAllow:
      - 'remote-*'
      - 'set-cookie'
      includeAdditionalHeadersInCheck:
        X-Forwarded-Proto: '%REQ(:SCHEME)%'
        X-Forwarded-Method: '%REQ(:METHOD)%'
        X-Forwarded-Uri: '%REQ(:PATH)%'
        X-Forwarded-For: '%DOWNSTREAM_REMOTE_ADDRESS_WITHOUT_PORT%'
```

### Authorization Policy

The following [Authorization Policy] applies the above filter to the `nextcloud.example.com` domain:

```yaml
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: nextcloud
  namespace: apps
spec:
  action: CUSTOM
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
