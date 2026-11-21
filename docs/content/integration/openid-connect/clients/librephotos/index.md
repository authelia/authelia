---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "LibrePhotos"
description: "A guide on integrating LibrePhotos with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "LibrePhotos | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring LibrePhotos with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [LibrePhotos]
  - [v1.6.0](https://github.com/LibrePhotos/librephotos/releases/tag/1.6.0)

{{% oidc-common bugs="client-credentials-encoding" %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://librephotos.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://librephotos.{{< sitevar name="domain" nojs="example.com" >}}/api/accounts/oidc/authelia/login/callback/`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `librephotos`
- **Client Secret:** `insecure_secret`
- **Provider ID:** `authelia`
  - This option determines the redirect URI in the format of
    `https://librephotos.{{< sitevar name="domain" nojs="example.com" >}}/api/accounts/oidc/authelia/login/callback/`.
    This means if you change this value, you need to update the redirect URI.

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [LibrePhotos] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'librephotos'
        client_name: 'LibrePhotos'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://librephotos.{{< sitevar name="domain" nojs="example.com" >}}/api/accounts/oidc/authelia/login/callback/'
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

To configure [LibrePhotos] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
The `FRONTEND_BASE_URL` environment variable must be set to the public URL of [LibrePhotos] (i.e.
`https://librephotos.{{< sitevar name="domain" nojs="example.com" >}}`) as it is used to build the redirect URI.
{{< /callout >}}

To configure [LibrePhotos] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [LibrePhotos] as an administrator.
2. Navigate to the Django administration page by going to `https://librephotos.{{< sitevar name="domain" nojs="example.com" >}}/api/django-admin/`.
3. Navigate to `Social applications`.
4. Click `Add social application`.
5. Configure the following options:
   - Provider: `OpenID Connect`
   - Provider ID: `authelia`
   - Name: `Authelia`
   - Client id: `librephotos`
   - Secret key: `insecure_secret`
   - Settings: `{"server_url": "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}", "token_auth_method": "client_secret_basic", "oauth_pkce_enabled": true}`
   - Sites: Select your site and move it to `Chosen sites`.
6. Click `Save`.
7. Navigate to the LibrePhotos `Site settings` and open the `Single sign-on` section.
8. Configure the following options:
   - OIDC_ENABLED: Enabled
   - OIDC_BUTTON_LABEL: `Login with Authelia`
   - OIDC_ALLOW_SIGNUP: Enable if you want first-time logins to create accounts (requires email to be configured).

## See Also

- [LibrePhotos Single Sign-On Documentation](https://docs.librephotos.com/docs/user-guide/settings/single-sign-on)
- [django-allauth OpenID Connect Provider Documentation](https://docs.allauth.org/en/latest/socialaccount/providers/openid_connect.html)

[Authelia]: https://www.authelia.com
[LibrePhotos]: https://librephotos.com
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
