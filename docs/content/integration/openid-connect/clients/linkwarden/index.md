---
title: "Linkwarden"
description: "Integrating Linkwarden with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-01-13T08:35:59+11:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/linkwarden/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Linkwarden | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Linkwarden with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.38.18](https://github.com/authelia/authelia/releases/tag/v4.38.18)
- [Linkwarden]
  - [v2.9.2](https://github.com/linkwarden/linkwarden/releases/tag/v2.9.2)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://linkwarden.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `linkwarden`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Linkwarden] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'linkwarden'
        client_name: 'Linkwarden'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://linkwarden.{{< sitevar name="domain" nojs="example.com" >}}/api/v1/auth/callback/authelia'
        scopes:
          - 'openid'
          - 'groups'
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

To configure [Linkwarden] there is one method, using [Environment Variables](#environment-variables).

#### Environment Variables

To configure [Linkwarden] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment variables:

##### Standard

```shell {title=".env"}
NEXT_PUBLIC_AUTHELIA_ENABLED=true
AUTHELIA_WELLKNOWN_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration
AUTHELIA_CLIENT_ID=linkwarden
AUTHELIA_CLIENT_SECRET=insecure_secret
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  linkwarden:
    environment:
      NEXT_PUBLIC_AUTHELIA_ENABLED: 'true'
      AUTHELIA_WELLKNOWN_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration'
      AUTHELIA_CLIENT_ID: 'linkwarden'
      AUTHELIA_CLIENT_SECRET: 'insecure_secret'
```

## See Also

- [Linkwarden OIDC documentation](https://docs.linkwarden.app/self-hosting/sso-oauth#authelia)

[Authelia]: https://www.authelia.com
[Linkwarden]: https://linkwarden.app/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
