---
title: "Zammad"
description: "Integrating Zammad with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-06-07T03:50:17+00:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/zammad/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Zammad | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Zammad with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.10](https://github.com/authelia/authelia/releases/tag/v4.39.10)
- [Zammad]
  - [v6.5.0](https://github.com/zammad/zammad/releases/tag/6.5.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://zammad.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `zammad`
- __Authentication Name (Zammad):__ `authelia`:
    - This option determines the redirect URI in the format of
      `https://zammad.{{< sitevar name="domain" nojs="example.com" >}}/user/oauth2/<Authentication Name>/callback`.
      This means if you change this value you need to update the redirect URI.

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Zammad] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'zammad'
        client_name: 'Zammad'
        public: true
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://zammad.{{< sitevar name="domain" nojs="example.com" >}}/auth/openid_connect/callback'
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
        token_endpoint_auth_method: 'none'
```

### Application

To configure [Zammad] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [Zammad] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Visit the Admin Panel
2. Visit Settings
3. Visit Security
4. Visit Third Party Applications
5. Enable Authentication via OpenID Connect
6. Configure the following options:
   - Display Name: `Authelia`
   - Identifier: `zammad`
   - Issuer: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
   - UID Field: `sub`
   - PKCE: `yes`
   - Scopes: `openid, email, profile`

## See Also

- [Zammad]:
    - [Security > Thirt Party Applications > OpenID Connect](https://admin-docs.zammad.org/en/pre-release/settings/security/third-party/openid-connect.html

[Authelia]: https://www.authelia.com
[Zammad]: https://zammad.com/en
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
