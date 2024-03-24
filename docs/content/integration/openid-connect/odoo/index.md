---
title: "Odoo"
description: "Integrating Odoo with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-03-24T09:09:0+01:00
draft: false
images: []
weight: 620
toc: true
community: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

* [Authelia]
  * [v4.38.6](https://github.com/authelia/authelia/releases/tag/v4.38.6)
* [Odoo]
  * [17.0](https://github.com/odoo/odoo/tree/17.0)

## Before You Begin

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://odoo.example.com/`
* __Authelia Root URL:__ `https://auth.example.com/`
* __Client ID:__ `client_id_odoo`

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__
[client configuration](../../../configuration/identity-providers/openid-connect/clients.md) for use with [Odoo]
which will operate with the above example:

```yaml
identity_providers:
  oidc:
    clients:
      - client_id: 'client_id_odoo'
        client_name: 'odoo.example.com'
        public: 'true'
        redirect_uris:
          - 'https://odoo.example.com/auth_oauth/signin'
        scopes:
          - 'openid'
          - 'email'
          - 'profile'
        response_types:
          - 'token'
        authorization_policy: 'one_factor'
        jwks:
          - key_id: 'rsa2048'
            key: {{ secret "/config/secrets/oidc/jwks/public.pem" | mindent 14 "|" | msquote }}
```

### Application

To configure [Odoo] to utilize Authelia as an [OpenID Connect 1.0] Provider:

1. Enable OAuth in General Settings/Integrations, save and reload.
2. Create a new OAuth Provider in General Settings/Integrations/OAuth Providers, with the following settings:
 * Provider name : Authelia
 * Client ID : client_id_odoo
 * Allowed : checked
 * Login button label : SSO example.com
 * Authorization URL : https://auth.example.com/api/oidc/authorization
 * Scope : openid profile email
 * UserInfo URL : https://auth.example.com/api/oidc/userinfo
 * Data Endpoint : https://auth.example.com/jwks.json
3. If you want your Authelia user to have a guest access on Odoo, you need to enable it in General Settings/Permissions/Customer Account/Free sign up
4. If you want to allow an already existing user in Odoo to use its Authelia login:
 * Ask the user to reset its password
 * When Odoo prompt for the new password, select the "Connect with SSO example.com" button

## See Also
 * [Odoo Authentication OpenID Connect]

[Authelia]: https://www.authelia.com
[Odoo]: https://www.odoo.com
[Odoo Authentication OpenID Connect]: https://odoo-community.org/shop/authentication-openid-connect-6545#attr=25818
[OpenID Connect 1.0]: ../../openid-connect/introduction.md

