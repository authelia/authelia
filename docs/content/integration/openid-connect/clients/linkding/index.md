---
title: "Linkding"
description: "Integrating Linkding with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-08-19T13:35:49+05:30
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/linkding/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Linkding | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Linkding with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.4](https://github.com/authelia/authelia/releases/tag/v4.39.4)
- [Linkding]
  - [v1.42.0](https://github.com/sissbruecker/linkding/releases/tag/v1.42.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://linkding.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `linkding`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Linkding

The following YAML configuration is an example Linkding [client configuration] for use with [Linkding] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'linkding'
        client_name: 'Linkding'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://linkding.{{< sitevar name="domain" nojs="example.com" >}}/oidc/callback/'
        scopes:
          - 'openid'
          - 'email'
          - 'profile'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

To configure [Linkding] there is one method, using the [Environment Variables](#environment-variables).

#### Environment Variables

To configure [Linkding] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
LD_ENABLE_OIDC=True
LD_CSRF_TRUSTED_ORIGINS=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
OIDC_RP_CLIENT_ID=linkding
OIDC_RP_CLIENT_SECRET=insecure_secret
OIDC_OP_AUTHORIZATION_ENDPOINT=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization
OIDC_OP_TOKEN_ENDPOINT=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token
OIDC_OP_USER_ENDPOINT=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo
OIDC_OP_JWKS_ENDPOINT=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/jwks.json
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  linkding:
    environment:
      LD_ENABLE_OIDC: 'True'
      LD_CSRF_TRUSTED_ORIGINS: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      OIDC_RP_CLIENT_ID: 'linkding'
      OIDC_RP_CLIENT_SECRET: 'insecure_secret'
      OIDC_OP_AUTHORIZATION_ENDPOINT: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization'
      OIDC_OP_TOKEN_ENDPOINT: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token'
      OIDC_OP_USER_ENDPOINT: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo'
      OIDC_OP_JWKS_ENDPOINT: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/jwks.json'
```

## See Also

- [Linkding Authenticating With an OpenID Provider Documentation](https://linkding.link/options/#ld_enable_oidc)

[Authelia]: https://www.authelia.com
[Linkding]: https://linkding.link/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
