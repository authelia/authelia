---
title: "Kanboard"
description: "Integrating Kanboard with the Authelia OpenID Connect 1.0 Provider."
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
  title: "Kanboard | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Kanboard with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.11](https://github.com/authelia/authelia/releases/tag/v4.39.11)
- [Kanboard]
  - [v1.2.46](https://github.com/kanboard/kanboard/releases/tag/v1.2.46)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://kanboard.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `kanboard`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Kanboard] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'kanboard'
        client_name: 'Kanboard'
        public: true
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://kanboard.{{< sitevar name="domain" nojs="example.com" >}}/oauth/callback'
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
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

To configure [Kanboard] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [Kanboard] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to your Kanboard administrator account.
2. Select `Settings` in the menu displayed when you click your profile icon.
3. If you do not have the OAuth2 plugin installed already:
   1. Select `Plugins`.
   2. Find the `OAuth2` plugin and click `Install`.
4. Select `Settings`.
5. Select `Integrations`.
6. Find `OAuth2 Authentication` and configure the following values:
   - Client ID: `kanboard`
   - Client Secret: `insecure_secret`
   - Authorize URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization`
   - Token URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token`
   - User API URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo`
   - Scopes: `openid email profile`
   - Username Key: `preferred_username`
   - Name Key: `name`
   - Email Key: `email`
   - User ID Key: `sub`
   - Allow Account Creation: Enabled
7. Click `Save`.

[Authelia]: https://www.authelia.com
[Kanboard]: https://kanboard.org/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
