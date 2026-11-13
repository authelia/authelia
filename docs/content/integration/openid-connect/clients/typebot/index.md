---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Typebot"
description: "A guide on integrating Typebot with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Typebot | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Typebot with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Typebot]
  - [v3.19.1](https://github.com/baptisteArno/typebot.io/releases/tag/v3.19.1)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://typebot.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option is the builder URL and determines the redirect URI in the format of
    `https://typebot.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/callback/custom-oauth`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `typebot`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Typebot] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'typebot'
        client_name: 'Typebot'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://typebot.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/callback/custom-oauth'
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

To configure [Typebot] there is one method, using [Environment Variables](#environment-variables).

#### Environment Variables

To configure [Typebot] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
NEXTAUTH_URL=https://typebot.{{< sitevar name="domain" nojs="example.com" >}}
CUSTOM_OAUTH_NAME=Authelia
CUSTOM_OAUTH_ISSUER=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
CUSTOM_OAUTH_CLIENT_ID=typebot
CUSTOM_OAUTH_CLIENT_SECRET=insecure_secret
CUSTOM_OAUTH_SCOPE=openid profile email
CUSTOM_OAUTH_USE_USERINFO=true
CUSTOM_OAUTH_USER_ID_PATH=sub
CUSTOM_OAUTH_USER_NAME_PATH=name
CUSTOM_OAUTH_USER_EMAIL_PATH=email
CUSTOM_OAUTH_USER_IMAGE_PATH=picture
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  typebot-builder:
    environment:
      NEXTAUTH_URL: 'https://typebot.{{< sitevar name="domain" nojs="example.com" >}}'
      CUSTOM_OAUTH_NAME: 'Authelia'
      CUSTOM_OAUTH_ISSUER: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      CUSTOM_OAUTH_CLIENT_ID: 'typebot'
      CUSTOM_OAUTH_CLIENT_SECRET: 'insecure_secret'
      CUSTOM_OAUTH_SCOPE: 'openid profile email'
      CUSTOM_OAUTH_USE_USERINFO: 'true'
      CUSTOM_OAUTH_USER_ID_PATH: 'sub'
      CUSTOM_OAUTH_USER_NAME_PATH: 'name'
      CUSTOM_OAUTH_USER_EMAIL_PATH: 'email'
      CUSTOM_OAUTH_USER_IMAGE_PATH: 'picture'
```

## See Also

- [Typebot Custom OAuth2 Documentation](https://docs.typebot.com/self-hosting/configuration#custom-oauth2)

[Authelia]: https://www.authelia.com
[Typebot]: https://typebot.com/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
