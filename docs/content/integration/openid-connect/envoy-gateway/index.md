---
title: "Envoy Gateway"
description: "Integrating Envoy Gateway with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-01-25T10:04:53+11:00
draft: false
images: []
weight: 620
toc: true
support:
  level: community
  versions: true
  integration: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.4](https://github.com/authelia/authelia/releases/tag/v4.39.4)
- [Envoy Gateway]
  - [v1.4.1](https://github.com/envoyproxy/gateway/releases/tag/v1.4.1)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://envoy-app.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `envoy-app`
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
      - client_id: 'envoy-app'
        client_name: 'Envoy Gateway'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://envoy-app.{{< sitevar name="domain" nojs="example.com" >}}/authelia/openid_connect/callback'
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

To configure [Envoy Gateway] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following
instructions:

1. Use `kubectl` to create the secret:
   - `kubectl create secret generic envoy-app-oidc-client-secret --from-literal=client-secret=insecure_secret`
2. Apply the below manifests for the example application.

```yaml {title="httproute.yaml
---
apiVersion: 'gateway.networking.k8s.io/v1'
kind: 'HTTPRoute'
metadata:
  name: 'envoy-app'
spec:
  parentRefs:
  - name: 'eg'
  hostnames:
    - 'envoy-app.{{< sitevar name="domain" nojs="example.com" >}}'
  rules:
  - matches:
    - path:
        type: 'PathPrefix'
        value: '/'
    backendRefs:
    - name: 'envoy-app-service-backend'
      port: 80
...
```

```yaml
---
apiVersion: 'gateway.envoyproxy.io/v1alpha1'
kind: 'SecurityPolicy'
metadata:
  name: 'envoy-app-oidc'
spec:
  targetRefs:
    - group: 'gateway.networking.k8s.io'
      kind: 'HTTPRoute'
      name: 'envoy-app'
  oidc:
    provider:
      issuer: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      authorizationEndpoint: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization'
      tokenEndpoint: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token'
    clientID: 'app'
    clientSecret:
      name: 'envoy-app-oidc-client-secret'
    cookieDomain: 'envoy-app.{{< sitevar name="domain" nojs="example.com" >}}'
    cookieNames:
      idToken: ''
      accessToken: ''
    scopes:
      - 'openid'
      - 'offline_access'
    redirectURL: 'https://envoy-app.{{< sitevar name="domain" nojs="example.com" >}}/authelia/openid_connect/callback'
    forwardAccessToken: false
    refreshToken: true
    passThroughAuthHeader: false
```

## See Also

- [Envoy Gateway]
- [OpenID Connect (OIDC) Authentication Documentation](https://docs.espocrm.com/administration/oidc/)

[Authelia]: https://www.authelia.com
[Envoy Gateway]: https://www.espocrm.com/
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
