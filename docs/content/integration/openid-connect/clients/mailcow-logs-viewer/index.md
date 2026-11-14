---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "mailcow Logs Viewer"
description: "A guide on integrating mailcow Logs Viewer with the Authelia OpenID Connect 1.0 Provider for SSO, with configuration examples and available options."
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
  title: "mailcow Logs Viewer | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring mailcow Logs Viewer with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [mailcow Logs Viewer]
  - [v2.9.0](https://github.com/ShlomiPorush/mailcow-logs-viewer/releases/tag/v2.9.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://mailcow-logs-viewer.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://mailcow-logs-viewer.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `mailcow-logs-viewer`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [mailcow Logs Viewer]
which will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'mailcow-logs-viewer'
        client_name: 'mailcow Logs Viewer'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://mailcow-logs-viewer.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/callback'
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

To configure [mailcow Logs Viewer] there is one method, using [Environment Variables](#environment-variables).

#### Environment Variables

To configure [mailcow Logs Viewer] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following
environment variables:

##### Standard

```shell {title=".env"}
OAUTH2_ENABLED=true
OAUTH2_PROVIDER_NAME=Authelia
OAUTH2_ISSUER_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
OAUTH2_USE_OIDC_DISCOVERY=true
OAUTH2_CLIENT_ID=mailcow-logs-viewer
OAUTH2_CLIENT_SECRET=insecure_secret
OAUTH2_REDIRECT_URI=https://mailcow-logs-viewer.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/callback
OAUTH2_SCOPES=openid profile email
SESSION_SECRET_KEY=<output of `openssl rand -hex 32`>
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  app:
    environment:
      OAUTH2_ENABLED: 'true'
      OAUTH2_PROVIDER_NAME: 'Authelia'
      OAUTH2_ISSUER_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      OAUTH2_USE_OIDC_DISCOVERY: 'true'
      OAUTH2_CLIENT_ID: 'mailcow-logs-viewer'
      OAUTH2_CLIENT_SECRET: 'insecure_secret'
      OAUTH2_REDIRECT_URI: 'https://mailcow-logs-viewer.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/callback'
      OAUTH2_SCOPES: 'openid profile email'
      SESSION_SECRET_KEY: '<output of `openssl rand -hex 32`>'
```

## See Also

- [mailcow Logs Viewer OAuth2 Configuration Documentation](https://github.com/ShlomiPorush/mailcow-logs-viewer/blob/main/documentation/OAuth2_Configuration.md)

[Authelia]: https://www.authelia.com
[mailcow Logs Viewer]: https://github.com/ShlomiPorush/mailcow-logs-viewer
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
