---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "DockMon"
description: "A guide on integrating DockMon with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "DockMon | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring DockMon with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [DockMon]
  - [v2.5.0](https://github.com/darthnorse/dockmon/releases/tag/v2.5.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://dockmon.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://dockmon.{{< sitevar name="domain" nojs="example.com" >}}/api/v2/auth/oidc/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `dockmon`
- **Client Secret:** `insecure_secret`
- **Admin Group Name:** `dockmon-admins`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [DockMon] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'dockmon'
        client_name: 'DockMon'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://dockmon.{{< sitevar name="domain" nojs="example.com" >}}/api/v2/auth/oidc/callback'
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

To configure [DockMon] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [DockMon] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [DockMon] as an administrator.
2. Navigate to `Settings` > `OIDC`.
3. Toggle `Enable OIDC` to the on position.
4. Configure the following options:
   - Provider URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
   - Client ID: `dockmon`
   - Client Secret: `insecure_secret`
   - Scopes: `openid profile email groups`
   - Groups Claim: `groups`
   - Default Group: `Read Only`
5. Click `Save Configuration`.
6. Navigate to `Settings` > `OIDC` > `Group Mappings`.
7. Click `Add Mapping`.
8. Configure the following options:
   - OIDC Group Name: `dockmon-admins`
   - DockMon Group: `Administrators`
   - Priority: `1`
9. Click `Save`.

## See Also

- [DockMon Multi-User & OIDC Documentation](https://github.com/darthnorse/dockmon/wiki/Multi-User-and-OIDC)

[Authelia]: https://www.authelia.com
[DockMon]: https://github.com/darthnorse/dockmon
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
