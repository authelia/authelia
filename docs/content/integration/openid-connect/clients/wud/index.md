---
title: "WUD (What's Up Docker)"
description: "Integrating WUD (What's Up Docker) with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-01-13T08:35:59+11:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/wud/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "WUD (What's Up Docker) | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring WUD (What's Up Docker) with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.38.18](https://github.com/authelia/authelia/releases/tag/v4.38.18)
- [WUD]
  - [v8.0.0](https://github.com/getwud/wud/releases/tag/8.0.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://wud.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `wud`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [WUD] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'wud'
        client_name: 'WUD'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://wud.{{< sitevar name="domain" nojs="example.com" >}}/auth/oidc/authelia/cb'
        scopes:
          - 'openid'
          - 'email'
          - 'profile'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [WUD] there is one method, using the [Environment Variables](#environment-variables).

#### Environment Variables

To configure [WUD] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment variables:

##### Standard

```shell {title=".env"}
WUD_AUTH_OIDC_AUTHELIA_DISCOVERY=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration
WUD_AUTH_OIDC_AUTHELIA_CLIENTID=wud
WUD_AUTH_OIDC_AUTHELIA_CLIENTSECRET=insecure_secret
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  wud:
    environment:
      WUD_AUTH_OIDC_AUTHELIA_DISCOVERY: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration'
      WUD_AUTH_OIDC_AUTHELIA_CLIENTID: 'wud'
      WUD_AUTH_OIDC_AUTHELIA_CLIENTSECRET: 'insecure_secret'
```

## See Also

- [WUD OIDC documentation](https://getwud.github.io/wud/#/configuration/authentications/oidc/?id=how-to-integrate-withnbspauthelia)

[Authelia]: https://www.authelia.com
[WUD]: https://getwud.github.io/wud/#/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
