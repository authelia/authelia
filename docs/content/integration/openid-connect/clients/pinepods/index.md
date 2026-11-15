---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Pinepods"
description: "A guide on integrating Pinepods with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Pinepods | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Pinepods with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Pinepods]
  - [0.9.2](https://github.com/madeofpendletonwool/PinePods/releases/tag/0.9.2)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://pinepods.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://pinepods.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `pinepods`
- **Client Secret:** `insecure_secret`
- **User Group Name:** `pinepods-users`
- **Admin Group Name:** `pinepods-admins`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Pinepods] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'pinepods'
        client_name: 'Pinepods'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://pinepods.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/callback'
        scopes:
          - 'openid'
          - 'email'
          - 'profile'
          - 'groups'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

To configure [Pinepods] there are two methods, using
[Environment Variables](#environment-variables), or using the [Web GUI](#web-gui).

#### Environment Variables

To configure [Pinepods] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
OIDC_PROVIDER_NAME=Authelia
OIDC_CLIENT_ID=pinepods
OIDC_CLIENT_SECRET=insecure_secret
OIDC_DISCOVERY_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration
OIDC_BUTTON_TEXT=Login with Authelia
OIDC_SCOPE=openid email profile groups
OIDC_NAME_CLAIM=name
OIDC_EMAIL_CLAIM=email
OIDC_USERNAME_CLAIM=preferred_username
OIDC_ROLES_CLAIM=groups
OIDC_USER_ROLE=pinepods-users
OIDC_ADMIN_ROLE=pinepods-admins
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  pinepods:
    environment:
      OIDC_PROVIDER_NAME: 'Authelia'
      OIDC_CLIENT_ID: 'pinepods'
      OIDC_CLIENT_SECRET: 'insecure_secret'
      OIDC_DISCOVERY_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration'
      OIDC_BUTTON_TEXT: 'Login with Authelia'
      OIDC_SCOPE: 'openid email profile groups'
      OIDC_NAME_CLAIM: 'name'
      OIDC_EMAIL_CLAIM: 'email'
      OIDC_USERNAME_CLAIM: 'preferred_username'
      OIDC_ROLES_CLAIM: 'groups'
      OIDC_USER_ROLE: 'pinepods-users'
      OIDC_ADMIN_ROLE: 'pinepods-admins'
```

#### Web GUI

{{< callout context="tip" title="Help Wanted" icon="outline/rocket" >}}
We would love screenshots of this configuration!
{{< /callout >}}

To configure [Pinepods] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [Pinepods] as an administrator.
2. Navigate to Settings.
3. Navigate to Admin Settings.
4. Navigate to OIDC/SSO Settings.
5. Click `Add Provider`.
6. Configure the following options:
   - Provider Name: `Authelia`
   - Client ID: `pinepods`
   - Client Secret: `insecure_secret`
   - Discovery URL (optional): `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration`
7. Click `Discover` and ensure the following options are populated:
   - Authorization URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization`
   - Token URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token`
   - User Info URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo`
8. Configure the following options:
   - Button Text: `Login with Authelia`
   - Scopes: `openid email profile groups`
   - Name Claim: `name`
   - Email Claim: `email`
   - Username Claim: `preferred_username`
   - Roles Claim: `groups`
   - User Role: `pinepods-users`
   - Admin Role: `pinepods-admins`
9. Click `Add`.

## See Also

- [Pinepods OIDC Setup Documentation](https://www.pinepods.online/docs/tutorial-extras/OIDC-setup)
- [Pinepods Environment Variables Documentation](https://www.pinepods.online/docs/tutorial-basics/environment-variables#oidc-openid-connect-configuration)

[Authelia]: https://www.authelia.com
[Pinepods]: https://www.pinepods.online/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
