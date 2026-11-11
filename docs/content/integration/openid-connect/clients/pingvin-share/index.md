---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Pingvin Share"
description: "A guide on integrating Pingvin Share with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Pingvin Share | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Pingvin Share with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Pingvin Share]
  - [v1.13.0](https://github.com/stonith404/pingvin-share/releases/tag/v1.13.0)

{{% oidc-common bugs="claims-hydration" %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://pingvin-share.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://pingvin-share.{{< sitevar name="domain" nojs="example.com" >}}/api/oauth/callback/oidc`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `pingvin-share`
- **Client Secret:** `insecure_secret`
- **User Group:** `pingvin-share`
- **Admin Group:** `pingvin-share-admins`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Pingvin Share] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'pingvin-share'
        client_name: 'Pingvin Share'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://pingvin-share.{{< sitevar name="domain" nojs="example.com" >}}/api/oauth/callback/oidc'
        scopes: ['openid', 'email', 'profile', 'groups']
        response_types: ['code']
        grant_types: ['authorization_code']
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_post'
```

#### Configuration Escape Hatch

{{% oidc-escape-hatch-claims-hydration client_id="pingvin-share" claims="email,preferred_username,name,groups" %}}

### Application

To configure [Pingvin Share] there are two methods, using the [Configuration File](#configuration-file), or using the
[Web GUI](#web-gui).

#### Configuration File

To configure [Pingvin Share] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```yaml {title="config.yaml"}
general:
  appUrl: 'https://pingvin-share.{{< sitevar name="domain" nojs="example.com" >}}'
oauth:
  oidc-enabled: 'true'
  oidc-discoveryUri: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration'
  oidc-signOut: 'false'
  oidc-scope: 'openid email profile groups'
  oidc-usernameClaim: 'preferred_username'
  oidc-rolePath: 'groups'
  oidc-roleGeneralAccess: 'pingvin-share'
  oidc-roleAdminAccess: 'pingvin-share-admins'
  oidc-clientId: 'pingvin-share'
  oidc-clientSecret: 'insecure_secret'
```

#### Web GUI

{{< callout context="tip" title="Help Wanted" icon="outline/rocket" >}}
We would love screenshots of this configuration!
{{< /callout >}}

To configure [Pingvin Share] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [Pingvin Share] as an administrator.
2. Navigate to Administration.
3. Navigate to Configuration.
4. Navigate to Social Login.
5. Configure the following options:
   - OpenID Connect: Enabled
   - OpenID Connect Discovery URI: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration`
   - Sign out from OpenID Connect: Disabled
   - OpenID Connect scope: `openid email profile groups`
   - OpenID Connect username claim: `preferred_username`
   - Path to roles in OpenID Connect token: `groups`
   - OpenID Connect role for general access: `pingvin-share`
   - OpenID Connect role for admin access: `pingvin-share-admins`
   - OpenID Connect Client ID: `pingvin-share`
   - OpenID Connect Client secret: `insecure_secret`
6. Click Save.

## See Also

- [Pingvin Share OAuth 2 Login Documentation](https://stonith404.github.io/pingvin-share/setup/oauth2login/)

[Authelia]: https://www.authelia.com
[Pingvin Share]: https://github.com/stonith404/pingvin-share
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
