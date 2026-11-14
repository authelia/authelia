---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Dockhand"
description: "A guide on integrating Dockhand with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Dockhand | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Dockhand with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Dockhand]
  - [v1.0.57](https://github.com/Finsys/dockhand/releases/tag/v1.0.57)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://dockhand.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://dockhand.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/oidc/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `dockhand`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Dockhand] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'dockhand'
        client_name: 'Dockhand'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://dockhand.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/oidc/callback'
        scopes: ['openid', 'email', 'profile']
        response_types: ['code']
        grant_types: ['authorization_code']
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

To configure [Dockhand] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

{{< callout context="tip" title="Help Wanted" icon="outline/rocket" >}}
We would love screenshots of this configuration!
{{< /callout >}}

To configure [Dockhand] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [Dockhand] as an administrator.
2. Navigate to Settings.
3. Navigate to Authentication.
4. Navigate to the SSO / OIDC section.
5. Click `Add provider`.
6. Configure the following options:
   - Name: `Authelia`
   - Issuer URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
   - Client ID: `dockhand`
   - Client secret: `insecure_secret`
   - Redirect URI: `https://dockhand.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/oidc/callback`
   - Scopes: `openid profile email`
   - Username claim: `preferred_username`
   - Email claim: `email`
   - Display name claim: `name`
7. Click Save.

## See Also

- [Dockhand Manual](https://dockhand.pro/manual/)

[Authelia]: https://www.authelia.com
[Dockhand]: https://dockhand.pro/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
