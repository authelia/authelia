---
title: "KitchenOwl"
description: "Integrating KitchenOwl with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-08-29T10:51:00+01:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/kitchenowl/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "KitchenOwl | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring KitchenOwl with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.38.18](https://github.com/authelia/authelia/releases/tag/v4.38.18)
- [KitchenOwl]
  - [v0.7.3](https://github.com/TomBursch/kitchenowl/releases/tag/v0.7.3)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://kitchenowl.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://kitchenowl.{{< sitevar name="domain" nojs="example.com" >}}/signin/redirect`.
    This means if you change this value, you need to update the redirect URI.
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `kitchenowl`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [KitchenOwl] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'kitchenowl'
        client_name: 'KitchenOwl'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://kitchenowl.{{< sitevar name="domain" nojs="example.com" >}}/signin/redirect'
          - 'kitchenowl:/signin/redirect'
        scopes:
          - 'openid'
          - 'profile'
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

To configure [KitchenOwl] there is one method, using [Environment Variables](#environment-variables).

#### Environment Variables

To configure [KitchenOwl] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
FRONT_URL=https://kitchenowl.{{< sitevar name="domain" nojs="example.com" >}}
OIDC_ISSUER=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
OIDC_CLIENT_ID=kitchenowl
OIDC_CLIENT_SECRET=insecure_secret
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  kitchenowl:
    environment:
      FRONT_URL: 'https://kitchenowl.{{< sitevar name="domain" nojs="example.com" >}}'
      OIDC_ISSUER: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      OIDC_CLIENT_ID: 'kitchenowl'
      OIDC_CLIENT_SECRET: 'insecure_secret'
```

## See Also

- [KitchenOwl OpenID Connect Documentation](https://docs.kitchenowl.org/latest/self-hosting/oidc/)

[Authelia]: https://www.authelia.com
[KitchenOwl]: https://kitchenowl.org
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
