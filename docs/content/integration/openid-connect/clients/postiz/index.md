---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Postiz"
description: "A guide on integrating Postiz with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Postiz | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Postiz with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Postiz]
  - [v2.25.3](https://github.com/gitroomhq/postiz-app/releases/tag/v2.25.3)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://postiz.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://postiz.{{< sitevar name="domain" nojs="example.com" >}}/settings`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `postiz`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Postiz] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'postiz'
        client_name: 'Postiz'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://postiz.{{< sitevar name="domain" nojs="example.com" >}}/settings'
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

To configure [Postiz] there is one method, using [Environment Variables](#environment-variables).

#### Environment Variables

To configure [Postiz] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
FRONTEND_URL=https://postiz.{{< sitevar name="domain" nojs="example.com" >}}
POSTIZ_GENERIC_OAUTH=true
NEXT_PUBLIC_POSTIZ_OAUTH_DISPLAY_NAME=Authelia
NEXT_PUBLIC_POSTIZ_OAUTH_LOGO_URL=https://www.authelia.com/images/branding/logo-cropped.png
POSTIZ_OAUTH_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
POSTIZ_OAUTH_AUTH_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization
POSTIZ_OAUTH_TOKEN_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token
POSTIZ_OAUTH_USERINFO_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo
POSTIZ_OAUTH_CLIENT_ID=postiz
POSTIZ_OAUTH_CLIENT_SECRET=insecure_secret
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  postiz:
    environment:
      FRONTEND_URL: 'https://postiz.{{< sitevar name="domain" nojs="example.com" >}}'
      POSTIZ_GENERIC_OAUTH: 'true'
      NEXT_PUBLIC_POSTIZ_OAUTH_DISPLAY_NAME: 'Authelia'
      NEXT_PUBLIC_POSTIZ_OAUTH_LOGO_URL: 'https://www.authelia.com/images/branding/logo-cropped.png'
      POSTIZ_OAUTH_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      POSTIZ_OAUTH_AUTH_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization'
      POSTIZ_OAUTH_TOKEN_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token'
      POSTIZ_OAUTH_USERINFO_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo'
      POSTIZ_OAUTH_CLIENT_ID: 'postiz'
      POSTIZ_OAUTH_CLIENT_SECRET: 'insecure_secret'
```

## See Also

- [Postiz OAuth / OIDC Configuration Documentation](https://docs.postiz.com/configuration/oauth)

[Authelia]: https://www.authelia.com
[Postiz]: https://postiz.com/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
