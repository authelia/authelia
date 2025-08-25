---
title: "Miniflux"
description: "Integrating Miniflux with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-04-26T11:03:16+00:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/miniflux/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Miniflux | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Miniflux with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.6](https://github.com/authelia/authelia/releases/tag/v4.39.6)
- [Miniflux]
  - [v2.2.8](https://github.com/miniflux/v2/releases/tag/2.2.8)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://miniflux.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `miniflux`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Miniflux] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'miniflux'
        client_name: 'Miniflux'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://miniflux.{{< sitevar name="domain" nojs="example.com" >}}/oauth2/oidc/callback'
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
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [Miniflux] there is one method, using the [Environment Variables](#environment-variables).

#### Environment Variables

To configure [Miniflux] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment variables:

##### Standard

```shell {title=".env"}
OAUTH2_OIDC_DISCOVERY_ENDPOINT=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
OAUTH2_CLIENT_ID=miniflux
OAUTH2_CLIENT_SECRET=insecure_secret
OAUTH2_OIDC_PROVIDER_NAME=Authelia
OAUTH2_PROVIDER=oidc
OAUTH2_REDIRECT_URL=https://miniflux.{{< sitevar name="domain" nojs="example.com" >}}/oauth2/oidc/callback
OAUTH2_USER_CREATION=1

```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  miniflux:
    environment:
      OAUTH2_OIDC_DISCOVERY_ENDPOINT: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      OAUTH2_CLIENT_ID: 'miniflux'
      OAUTH2_CLIENT_SECRET: 'insecure_secret'
      OAUTH2_OIDC_PROVIDER_NAME: 'Authelia'
      OAUTH2_PROVIDER: 'oidc'
      OAUTH2_REDIRECT_URL: 'https://miniflux.{{< sitevar name="domain" nojs="example.com" >}}/oauth2/oidc/callback'
      OAUTH2_USER_CREATION: '1'
```

## See Also

- [Miniflux Configuration Documentation](https://miniflux.app/docs/configuration.html#oauth2-oidc-discovery-endpoint)

[Miniflux]: https://miniflux.app/index.html
[Authelia]: https://www.authelia.com
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
