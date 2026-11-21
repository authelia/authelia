---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Termix"
description: "A guide on integrating Termix with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Termix | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Termix with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Termix]
  - [v26.11.1](https://github.com/Termix-SSH/Termix/releases/tag/v26.11.1)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://termix.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://termix.{{< sitevar name="domain" nojs="example.com" >}}/plugin-api/sso/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `termix`
- **Client Secret:** `insecure_secret`
- **Administrator Group:** `termix-admins`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Termix] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'termix'
        client_name: 'Termix'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://termix.{{< sitevar name="domain" nojs="example.com" >}}/plugin-api/sso/callback'
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

To configure [Termix] there are two methods, using the [Web GUI](#web-gui), or using
[Environment Variables](#environment-variables). Both methods require the [Termix] Single sign-on plugin to be
installed from the Plugins tab.

#### Web GUI

To configure [Termix] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [Termix] as an administrator.
2. Navigate to `Settings`.
3. Navigate to `Single sign-on`.
4. Click `Add Provider`.
5. Configure the following options:
   - Provider Type: `OIDC`
   - Display Name: `Authelia`
   - Client ID: `termix`
   - Client Secret: `insecure_secret`
   - Issuer URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
   - Authorization URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization`
   - Token URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token`
   - User Identifier Path: `sub`
   - Username Path: `preferred_username`
   - Scopes: `openid email profile groups`
   - Group Claim: `groups`
   - Admin Group: `termix-admins`
6. Ensure the `Redirect URI` displayed in the form matches the redirect URI configured in Authelia.
7. Toggle `Enabled` to the on position.
8. Click `Save Provider`.

#### Environment Variables

To configure [Termix] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
OIDC_CLIENT_ID=termix
OIDC_CLIENT_SECRET=insecure_secret
OIDC_ISSUER_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
OIDC_AUTHORIZATION_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization
OIDC_TOKEN_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token
OIDC_USERINFO_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo
OIDC_IDENTIFIER_PATH=sub
OIDC_NAME_PATH=preferred_username
OIDC_SCOPES=openid email profile groups
OIDC_GROUP_CLAIM=groups
OIDC_ADMIN_GROUP=termix-admins
OIDC_LEGACY_CALLBACK=false
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  termix:
    environment:
      OIDC_CLIENT_ID: 'termix'
      OIDC_CLIENT_SECRET: 'insecure_secret'
      OIDC_ISSUER_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      OIDC_AUTHORIZATION_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization'
      OIDC_TOKEN_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token'
      OIDC_USERINFO_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo'
      OIDC_IDENTIFIER_PATH: 'sub'
      OIDC_NAME_PATH: 'preferred_username'
      OIDC_SCOPES: 'openid email profile groups'
      OIDC_GROUP_CLAIM: 'groups'
      OIDC_ADMIN_GROUP: 'termix-admins'
      OIDC_LEGACY_CALLBACK: 'false'
```

The `OIDC_LEGACY_CALLBACK` option set to `false` ensures [Termix] uses the `/plugin-api/sso/callback` redirect URI
rather than the legacy `/users/oidc/callback` redirect URI used by versions prior to 2.9.

## See Also

- [Termix Single sign-on Documentation](https://docs.termix.site/plugins/sso)
- [Termix Single sign-on Reference Documentation](https://docs.termix.site/plugins/sso/reference)

[Authelia]: https://www.authelia.com
[Termix]: https://termix.site/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
