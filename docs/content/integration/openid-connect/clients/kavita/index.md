---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Kavita"
description: "A guide on integrating Kavita with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Kavita | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Kavita with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Kavita]
  - [v0.9.1.4](https://github.com/Kareadita/Kavita/releases/tag/v0.9.1.4)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://kavita.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://kavita.{{< sitevar name="domain" nojs="example.com" >}}/signin-oidc`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `kavita`
- **Client Secret:** `insecure_secret`
- **Roles Prefix:** `kavita-`
  - Users who should be able to log in must be a member of the `kavita-login` group when
    [syncing user settings with OIDC roles](#web-gui).

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Kavita] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'kavita'
        client_name: 'Kavita'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://kavita.{{< sitevar name="domain" nojs="example.com" >}}/signin-oidc'
        scopes:
          - 'openid'
          - 'offline_access'
          - 'profile'
          - 'email'
          - 'groups'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
          - 'refresh_token'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

To configure [Kavita] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [Kavita] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [Kavita] as an administrator.
2. Navigate to `Settings`.
3. Navigate to `OpenID Connect` under `Server`.
4. Configure the following options:
   - Authority: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
   - Client ID: `kavita`
   - Secret: `insecure_secret`
   - Provider name: `Authelia`
   - Provision accounts: Enabled
   - Require verified emails: Enabled
   - Sync user settings with OIDC roles: Enabled
   - Custom scopes: `groups`
   - Roles claim: `groups`
   - Roles prefix: `kavita-`
5. Click `Save`.
6. Restart [Kavita] for the changes to take effect.

When `Sync user settings with OIDC roles` is enabled the roles, library access, and age restrictions for each user are
derived from the Authelia groups which begin with the configured prefix. For example a user who is a member of the
`kavita-login`, `kavita-download`, and `kavita-library-Comics` groups is granted the `Login` and `Download` roles and
access to the `Comics` library. The `Login` role is required for a user to log in. See the
[Kavita OpenID Connect Documentation](https://wiki.kavitareader.com/guides/admin-settings/open-id-connect/) for the
full list of roles.

## See Also

- [Kavita OpenID Connect Documentation](https://wiki.kavitareader.com/guides/admin-settings/open-id-connect/)

[Authelia]: https://www.authelia.com
[Kavita]: https://www.kavitareader.com/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
