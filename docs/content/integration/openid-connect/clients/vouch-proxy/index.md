---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Vouch Proxy"
description: "A guide on integrating Vouch Proxy with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
summary: ""
date: 2026-11-12T00:00:00+00:00
draft: false
images: []
weight: 620
toc: true
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Vouch Proxy | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Vouch Proxy with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Vouch Proxy]
  - [v0.50.0](https://github.com/vouch/vouch-proxy/releases/tag/v0.50.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://vouch.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://vouch.{{< sitevar name="domain" nojs="example.com" >}}/auth`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `vouch-proxy`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Vouch Proxy] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'vouch-proxy'
        client_name: 'Vouch Proxy'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://vouch.{{< sitevar name="domain" nojs="example.com" >}}/auth'
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

To configure [Vouch Proxy] there are two methods, using the [Configuration File](#configuration-file), or using
[Environment Variables](#environment-variables).

#### Configuration File

To configure [Vouch Proxy] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```yaml {title="config/config.yml"}
vouch:
  domains:
    - '{{< sitevar name="domain" nojs="example.com" >}}'
  cookie:
    secure: true
    domain: '{{< sitevar name="domain" nojs="example.com" >}}'

oauth:
  provider: 'oidc'
  client_id: 'vouch-proxy'
  client_secret: 'insecure_secret'
  auth_url: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization'
  token_url: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token'
  user_info_url: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo'
  scopes:
    - 'openid'
    - 'email'
    - 'profile'
  callback_url: 'https://vouch.{{< sitevar name="domain" nojs="example.com" >}}/auth'
  code_challenge_method: 'S256'
```

#### Environment Variables

To configure [Vouch Proxy] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
VOUCH_DOMAINS={{< sitevar name="domain" nojs="example.com" >}}
VOUCH_COOKIE_SECURE=true
VOUCH_COOKIE_DOMAIN={{< sitevar name="domain" nojs="example.com" >}}
OAUTH_PROVIDER=oidc
OAUTH_CLIENT_ID=vouch-proxy
OAUTH_CLIENT_SECRET=insecure_secret
OAUTH_AUTH_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization
OAUTH_TOKEN_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token
OAUTH_USER_INFO_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo
OAUTH_SCOPES=openid,email,profile
OAUTH_CALLBACK_URL=https://vouch.{{< sitevar name="domain" nojs="example.com" >}}/auth
OAUTH_CODE_CHALLENGE_METHOD=S256
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  vouch-proxy:
    environment:
      VOUCH_DOMAINS: '{{< sitevar name="domain" nojs="example.com" >}}'
      VOUCH_COOKIE_SECURE: 'true'
      VOUCH_COOKIE_DOMAIN: '{{< sitevar name="domain" nojs="example.com" >}}'
      OAUTH_PROVIDER: 'oidc'
      OAUTH_CLIENT_ID: 'vouch-proxy'
      OAUTH_CLIENT_SECRET: 'insecure_secret'
      OAUTH_AUTH_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization'
      OAUTH_TOKEN_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token'
      OAUTH_USER_INFO_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo'
      OAUTH_SCOPES: 'openid,email,profile'
      OAUTH_CALLBACK_URL: 'https://vouch.{{< sitevar name="domain" nojs="example.com" >}}/auth'
      OAUTH_CODE_CHALLENGE_METHOD: 'S256'
```

## See Also

- [Vouch Proxy OpenID Connect Example Configuration](https://github.com/vouch/vouch-proxy/blob/master/config/config.yml_example_oidc)
- [Vouch Proxy Example Configuration](https://github.com/vouch/vouch-proxy/blob/master/config/config.yml_example)

[Authelia]: https://www.authelia.com
[Vouch Proxy]: https://github.com/vouch/vouch-proxy
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
