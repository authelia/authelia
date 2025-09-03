---
title: "Trillium Notes"
description: "Integrating Trillium Notes with the Authelia OpenID Connect 1.0 Provider."
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
  title: "Trillium Notes | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Trillium Notes with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.8](https://github.com/authelia/authelia/releases/tag/v4.39.8)
- [Trillium Notes]
  - [v0.97.1](https://github.com/TriliumNext/Trilium/releases/tag/v0.97.1)

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://trillium.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `trillium`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Trillium Notes] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'trillium'
        client_name: 'Trillium Notes'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
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

To configure [Trillium Notes] there are two methods, using the [Configuration File](#configuration-file) or using the
[Environment Variables](#environment-variables).

#### Configuration File

To configure [Trillium Notes] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

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

To configure [Trillium Notes] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
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

- [Trillium Notes Authentication Documentation](https://github.com/TriliumNext/Trilium/blob/main/docs/User%20Guide/User%20Guide/Installation%20%26%20Setup/Server%20Installation/Authentication.md)
- [Trillium Notes Multi-Factor Authentication Documentation](https://github.com/TriliumNext/Trilium/blob/main/docs/User%20Guide/User%20Guide/Installation%20%26%20Setup/Server%20Installation/Multi-Factor%20Authentication.md)

[Authelia]: https://www.authelia.com
[Trillium Notes]: https://github.com/TriliumNext/Trilium
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
