---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "OpenList"
description: "A guide on integrating OpenList with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "OpenList | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring OpenList with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [OpenList]
  - [v4.2.9](https://github.com/OpenListTeam/OpenList/releases/tag/v4.2.9)

{{% oidc-common bugs="claims-hydration" %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://openlist.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URIs in the format of
    `https://openlist.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/sso_callback?method=sso_get_token` and
    `https://openlist.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/sso_callback?method=get_sso_id`.
    This means if you change this value, you need to update the redirect URIs.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `openlist`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [OpenList] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'openlist'
        client_name: 'OpenList'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://openlist.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/sso_callback?method=sso_get_token'
          - 'https://openlist.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/sso_callback?method=get_sso_id'
        scopes:
          - 'openid'
          - 'profile'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

#### Configuration Escape Hatch

{{% oidc-escape-hatch-claims-hydration client_id="openlist" claims="preferred_username" %}}

### Application

To configure [OpenList] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [OpenList] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [OpenList] as an administrator.
2. Navigate to Manage.
3. Navigate to Settings.
4. Navigate to SSO.
5. Configure the following options:
   - Sso login enabled: Enabled
   - Sso login platform: `OIDC`
   - Sso client id: `openlist`
   - Sso client secret: `insecure_secret`
   - Sso oidc username key: `preferred_username`
   - Sso endpoint name: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
   - Sso compatibility mode: Disabled
   - Sso auto register: Enable if you want users to automatically be created in [OpenList].
6. Click `Save`.
7. Navigate to Profile and click the button to bind the current account to the SSO account.

## See Also

- [OpenList Single Sign-On Documentation](https://doc.oplist.org/guide/advanced/sso)

[Authelia]: https://www.authelia.com
[OpenList]: https://github.com/OpenListTeam/OpenList
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
