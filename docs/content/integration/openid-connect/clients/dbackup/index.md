---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "DBackup"
description: "A guide on integrating DBackup with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "DBackup | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring DBackup with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [DBackup]
  - [v4.0.2](https://github.com/Skyfay/DBackup/releases/tag/v4.0.2)

{{% oidc-common bugs="client-credentials-encoding" %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://dbackup.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://dbackup.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/sso/callback/authelia`.
    This means if you change this value, you need to update the redirect URI.
  - This value must match the `BETTER_AUTH_URL` environment variable of [DBackup].
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `dbackup`
- **Client Secret:** `insecure_secret`
- **Provider ID:** `authelia`
  - This option determines the redirect URI in the format of
    `https://dbackup.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/sso/callback/authelia`.
    This means if you change this value, you need to update the redirect URI.

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [DBackup] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'dbackup'
        client_name: 'DBackup'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://dbackup.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/sso/callback/authelia'
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

To configure [DBackup] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [DBackup] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [DBackup] as a SuperAdmin.
2. Navigate to `Users & Groups` and select the `SSO` tab.
3. Click `New provider`.
4. Select `Authelia`.
5. Configure the following options:
   - Name: `Authelia`
   - Provider ID: `authelia`
   - Authelia URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
   - Client ID: `dbackup`
   - Client Secret: `insecure_secret`
6. Configure `Add new people on their first sign-in` and the group new users are added to as desired.
7. Click `Create provider`.

## See Also

- [DBackup SSO / OIDC Documentation](https://docs.dbackup.app/user-guide/admin/sso)

[Authelia]: https://www.authelia.com
[DBackup]: https://dbackup.app/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
