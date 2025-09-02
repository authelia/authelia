---
title: "Envoy Gateway"
description: "Integrating Envoy Gateway with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-06-13T14:12:09+00:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/envoy-gateway/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Envoy Gateway | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Envoy Gateway with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.8](https://github.com/authelia/authelia/releases/tag/v4.39.8)
- [Envoy Gateway]
  - [v1.4.1](https://github.com/envoyproxy/gateway/releases/tag/v1.4.1)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://envoy.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `envoy`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Envoy Gateway] which will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'envoy'
        client_name: 'Envoy Gateway'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://envoy.{{< sitevar name="domain" nojs="example.com" >}}/authelia/openid_connect/callback'
        scopes:
          - 'openid'
          - 'offline_access'
        grant_types:
          - 'authorization_code'
          - 'refresh_token'
        response_types:
          - 'code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [Envoy Gateway] there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

{{< callout context="caution" title="Important Notes" icon="outline/alert-triangle" >}}
Because this setup stores the ID Token and Access Token in session cookies, it is strongly recommended that all of the
following are true:
  - Each application has an individual Security Policy applied.
  - Each Security Policy has a specific domain configured that is a complete match for the protected application.
{{< /callout >}}

##### Apply to a HTTPRoute

To configure [Envoy Gateway] to utilize Authelia as an [OpenID Connect 1.0] Provider for a single [HTTPRoute], use the
following instructions:

1. Use `kubectl` to create the secret:
   - `kubectl create secret generic envoy-oidc-client-secret --from-literal=client-secret=insecure_secret`
2. Apply the below manifests for the example application.

The following example [HTTPRoute] is a example real application just for the purposes of showcasing this. The important
factors are the `name` value being `envoy`.

```yaml {title="httproute.yaml
---
apiVersion: 'gateway.networking.k8s.io/v1'
kind: 'HTTPRoute'
metadata:
  name: 'envoy'
spec:
  parentRefs:
    - name: 'eg'
  hostnames:
    - 'envoy.{{< sitevar name="domain" nojs="example.com" >}}'
  rules:
    - matches:
        - path:
            type: 'PathPrefix'
            value: '/'
      backendRefs:
        - name: 'envoy-service-backend'
          port: 80
...
```

The following [SecurityPolicy] requires [OpenID Connect 1.0] authorization for just the `envoy` [HTTPRoute] as
described above, the important factors are the `targetRefs` which indicates what resource to apply this to.

```yaml
---
apiVersion: 'gateway.envoyproxy.io/v1alpha1'
kind: 'SecurityPolicy'
metadata:
  name: 'envoy-oidc'
spec:
  targetRefs:
    - group: 'gateway.networking.k8s.io'
      kind: 'HTTPRoute'
      name: 'envoy'
  oidc:
    provider:
      issuer: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      authorizationEndpoint: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization'
      tokenEndpoint: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token'
    clientID: 'envoy'
    clientSecret:
      name: 'envoy-oidc-client-secret'
    cookieDomain: 'envoy.{{< sitevar name="domain" nojs="example.com" >}}'
    cookieNames:
      idToken: ''
      accessToken: ''
    scopes:
      - 'openid'
      - 'offline_access'
    redirectURL: 'https://envoy.{{< sitevar name="domain" nojs="example.com" >}}/authelia/openid_connect/callback'
    forwardAccessToken: false
    refreshToken: true
    passThroughAuthHeader: false
```

##### Apply to an entire Gateway

To configure [Envoy Gateway] to utilize Authelia as an [OpenID Connect 1.0] Provider for an entire [Gateway], use the
following instructions:

1. Use `kubectl` to create the secret:
  - `kubectl create secret generic envoy-oidc-client-secret --from-literal=client-secret=insecure_secret`
2. Apply the below manifests for the `eg` [Gateway].

The following example [HTTPRoute] is a fake application just for the redirection behaviour.

```yaml {title="httproute.yaml
---
apiVersion: 'gateway.networking.k8s.io/v1'
kind: 'HTTPRoute'
metadata:
  name: 'envoy-oidc'
spec:
  parentRefs:
    - name: 'eg'
  hostnames:
    - 'envoy-oidc.{{< sitevar name="domain" nojs="example.com" >}}'
  rules:
    - matches:
        - path:
            type: 'PathPrefix'
            value: '/'
...
```

The following [SecurityPolicy] requires [OpenID Connect 1.0] authorization for every [HTTPRoute] on the `eg` [Gateway],
the important factors are the `targetRefs` which indicates what resource to apply this to.

```yaml
---
apiVersion: 'gateway.envoyproxy.io/v1alpha1'
kind: 'SecurityPolicy'
metadata:
  name: 'envoy-oidc'
spec:
  targetRefs:
    - group: 'gateway.networking.k8s.io'
      kind: 'Gateway'
      name: 'eg'
  oidc:
    provider:
      issuer: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      authorizationEndpoint: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization'
      tokenEndpoint: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token'
    clientID: 'envoy'
    clientSecret:
      name: 'envoy-oidc-client-secret'
    cookieDomain: 'envoy-oidc.{{< sitevar name="domain" nojs="example.com" >}}'
    cookieNames:
      idToken: ''
      accessToken: ''
    scopes:
      - 'openid'
      - 'offline_access'
    redirectURL: 'https://envoy-oidc.{{< sitevar name="domain" nojs="example.com" >}}/authelia/openid_connect/callback'
    forwardAccessToken: false
    refreshToken: true
    passThroughAuthHeader: false
```

## See Also

- [Envoy Gateway]
- [OIDC Authentication Security Tasks](https://gateway.envoyproxy.io/latest/tasks/security/oidc/)

[Authelia]: https://www.authelia.com
[Envoy Gateway]: https://gateway.envoyproxy.io/
[Gateway API]: https://gateway-api.sigs.k8s.io/
[SecurityPolicy]: https://gateway.envoyproxy.io/contributions/design/security-policy/
[HTTPRoute]: https://gateway-api.sigs.k8s.io/api-types/httproute/
[Gateway]: https://gateway-api.sigs.k8s.io/api-types/gateway/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
