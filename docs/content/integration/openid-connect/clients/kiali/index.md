---
title: "Kiali"
description: "Integrating Kiali with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 620
toc: true
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Kiali | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Kiali with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.14](https://github.com/authelia/authelia/releases/tag/v4.39.14)
- [Kiali]
  - [v2.12.0](https://kiali.io/news/release-notes/#v2120)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://kiali.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `kiali`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
At the time of this writing this third party client has a bug and does not support [OpenID Connect 1.0](https://openid.net/specs/openid-connect-core-1_0.html). This
configuration will likely require configuration of an escape hatch to work around the bug on their end. See
[Configuration Escape Hatch](#configuration-escape-hatch) for details.
{{< /callout >}}

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Kiali] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'kiali'
        client_name: 'Kiali'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://kiali.{{< sitevar name="domain" nojs="example.com" >}}/kiali'
        scopes:
          - 'openid'
          - 'email'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

To configure [Kiali] there are two methods, using the [Configuration File](#configuration-file), or using
[Environment Variables](#environment-variables).

#### Configuration File

To configure [Kiali] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

Adjust your Kiali CR YAML file:

```yaml {title="Kiali CR"}
spec:
  auth:
    strategy: 'openid'
    openid:
      client_id: 'kiali'
      disable_rbac: true
      issuer_uri: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      scopes: ['openid', 'email']
      username_claim: 'email'
```

Add the [OpenID Connect 1.0] Client Secret:

```yaml {title="kiali.openid.secret.yaml"}
apiVersion: v1
kind: Secret
metadata:
  name: 'kiali'
  namespace: 'istio-system'
  labels:
    app: 'kiali'
type: 'Opaque'
stringData:
  oidc-secret: 'insecure_secret'
```

## See Also

- [Kiali OpenID Connect strategy Documentation](https://kiali.io/docs/configuration/authentication/openid)

[Authelia]: https://www.authelia.com
[Kiali]: https://kiali.io/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
