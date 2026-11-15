---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Trilium Notes"
description: "A guide on integrating Trilium Notes with the Authelia OpenID Connect 1.0 Provider with configuration examples and an outline of the available options."
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
  title: "Trilium Notes | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Trilium Notes with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Trilium Notes]
  - [v0.97.1](https://github.com/TriliumNext/Trilium/releases/tag/v0.97.1)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://trillium.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://trillium.{{< sitevar name="domain" nojs="example.com" >}}/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `trillium`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Trilium Notes] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'trillium'
        client_name: 'Trilium Notes'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://trillium.{{< sitevar name="domain" nojs="example.com" >}}/callback'
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

To configure [Trilium Notes] there are two methods, using the [Configuration File](#configuration-file) or using the
[Environment Variables](#environment-variables).

#### Configuration File

To configure [Trilium Notes] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```ini {title="config.ini"}
[MultiFactorAuthentication]
oauthBaseUrl=https://trillium.{{< sitevar name="domain" nojs="example.com" >}}
oauthClientId=trillium
oauthClientSecret=insecure_secret
oauthIssuerBaseUrl=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
oauthIssuerName=Authelia
oauthIssuerIcon=https://www.authelia.com/images/branding/logo-cropped.png
```

#### Environment Variables

To configure [Trilium Notes] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
TRILIUM_OAUTH_BASE_URL=https://trillium.{{< sitevar name="domain" nojs="example.com" >}}
TRILIUM_OAUTH_CLIENT_ID=trillium
TRILIUM_OAUTH_CLIENT_SECRET=insecure_secret
TRILIUM_OAUTH_ISSUER_BASE_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
TRILIUM_OAUTH_ISSUER_NAME=Authelia
TRILIUM_OAUTH_ISSUER_ICON=https://www.authelia.com/images/branding/logo-cropped.png
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  trillium:
    environment:
      TRILIUM_OAUTH_BASE_URL: 'https://trillium.{{< sitevar name="domain" nojs="example.com" >}}'
      TRILIUM_OAUTH_CLIENT_ID: 'trillium'
      TRILIUM_OAUTH_CLIENT_SECRET: 'insecure_secret'
      TRILIUM_OAUTH_ISSUER_BASE_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      TRILIUM_OAUTH_ISSUER_NAME: 'Authelia'
      TRILIUM_OAUTH_ISSUER_ICON: 'https://www.authelia.com/images/branding/logo-cropped.png'
```

## See Also

- [Trilium Notes Authentication Documentation](https://github.com/TriliumNext/Trilium/blob/main/docs/User%20Guide/User%20Guide/Installation%20%26%20Setup/Server%20Installation/Authentication.md)
- [Trilium Notes Multi-Factor Authentication Documentation](https://github.com/TriliumNext/Trilium/blob/main/docs/User%20Guide/User%20Guide/Installation%20%26%20Setup/Server%20Installation/Multi-Factor%20Authentication.md)

[Authelia]: https://www.authelia.com
[Trilium Notes]: https://github.com/TriliumNext/Trilium
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
