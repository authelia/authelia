---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Gameyfin"
description: "A guide on integrating Gameyfin with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Gameyfin | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Gameyfin with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Gameyfin]
  - [v2.6.1](https://github.com/gameyfin/gameyfin/releases/tag/v2.6.1)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://gameyfin.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://gameyfin.{{< sitevar name="domain" nojs="example.com" >}}/login/oauth2/code/oidc`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `gameyfin`
- **Client Secret:** `insecure_secret`
- **Groups:**
  - `GAMEYFIN_SUPERADMIN`: Users in this group are granted the Super Admin role.
  - `GAMEYFIN_ADMIN`: Users in this group are granted the Admin role.
  - Users in neither group are granted the User role.

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Gameyfin] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'gameyfin'
        client_name: 'Gameyfin'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://gameyfin.{{< sitevar name="domain" nojs="example.com" >}}/login/oauth2/code/oidc'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
          - 'groups'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [Gameyfin] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [Gameyfin] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [Gameyfin] as an administrator.
2. Navigate to `Administration`.
3. Navigate to `Security`.
4. Under `Single Sign-On` configure the following options:
   - Enable SSO via OIDC/OAuth2: Enabled
   - Match existing users by: `username`
   - Username claim: `preferred_username`
   - Role claim: `groups`
   - OAuth2 scopes to request: `openid`, `profile`, `email`, `groups`
   - Client ID: `gameyfin`
   - Client secret: `insecure_secret`
   - Issuer URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
5. Click `Auto-populate` and ensure the following options are configured:
   - Authorize URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization`
   - Token URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token`
   - Userinfo URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo`
   - JWKS URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/jwks.json`
6. Manually configure the following option as it is not populated by `Auto-populate`:
   - Logout URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/logout`
7. Click `Save`.
8. Restart [Gameyfin] for the changes to take effect.

## See Also

- [Gameyfin SSO Documentation](https://gameyfin.org/configuration/sso/)

[Authelia]: https://www.authelia.com
[Gameyfin]: https://gameyfin.org/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
