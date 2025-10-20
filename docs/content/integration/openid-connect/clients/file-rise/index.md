---
title: "FileRise"
description: "Integrating FileRise with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-04-26T18:35:57+10:00
draft: false
images: []
weight: 620
toc: true
aliases: []
support:
  level: community
  versions: true
  integration: true
seo:
  title: "FileRise | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring FileRise with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.13](https://github.com/authelia/authelia/releases/tag/v4.39.13)
- [FileRise]
  - [v1.3.9](https://github.com/error311/FileRise/releases/tag/v1.3.9)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://filerise.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `filerise`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [FileRise] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'filerise'
        client_name: 'FileRise'
        public: true
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://filerise.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/auth.php?oidc=callback'
        scopes:
          - 'openid'
          - 'email'
          - 'profile'
        response_modes:
          - 'form_post'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'none'
```

### Application

To configure [FileRise] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [FileRise] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to your FileRise administrator account.
2. Select `Admin Panel` from the context menu revealed by clicking your profile icon.
3. Select `OIDC Configuration & TOTP`.
4. Enter the following values:
  - OIDC Provider URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
  - OIDC Client ID: `filerise`
  - OIDC Client Secret: `insecure_secret`
  - OIDC Redirect URI: `https://filerise.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/auth.php?oidc=callback`
5. Click `Save Settings`.

[Authelia]: https://www.authelia.com
[FileRise]: https://github.com/error311/FileRise
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
