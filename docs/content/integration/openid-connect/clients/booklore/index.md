---
title: "BookLore"
description: "Integrating BookLore with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-03-14T06:00:14+11:00
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
  title: "BookLore | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring BookLore with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.14](https://github.com/authelia/authelia/releases/tag/v4.39.14)
- [BookLore]
  - [v1.5.1](https://github.com/booklore-app/booklore/releases/tag/v1.5.1)

{{% oidc-common bugs="claims-hydration" %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://booklore.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `booklore`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [BookLore] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'booklore'
        client_name: 'BookLore'
        public: true
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://booklore.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc'
        scopes:
          - 'openid'
          - 'offline_access'
          - 'profile'
          - 'email'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
          - 'refresh_token'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'none'
```

#### Configuration Escape Hatch

{{% oidc-escape-hatch-claims-hydration client_id="booklore" claims="email,preferred_username,name" %}}

### Application

To configure [BookLore] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [BookLore] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. In the top right click the settings icon (looks like a cog)
2. Click `Authentication`
3. Under `OIDC Authentication (Experimental)` configure the following options:
   - OIDC Enabled: Toggle to the on Position
   - Provider Name: `Authelia`
   - Client ID: `booklore`
   - Issuer URI: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
   - Scope: `openid profile email offline_access`
   - Username Claim: `preferred_username`
   - Email Claim: `email`
   - Display Name Claim: `name`
4. Click `Save Settings`.

{{< figure src="booklore.png" alt="BookLore" >}}

## See Also

There are currently no additional resources related to this client.

[Authelia]: https://www.authelia.com
[BookLore]: https://github.com/booklore-app/booklore
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
