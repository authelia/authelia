---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "AList"
description: "A guide on integrating AList with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "AList | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring AList with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [AList]
  - [v3.66.0](https://github.com/AlistGo/alist/releases/tag/v3.66.0)

{{% oidc-common bugs="claims-hydration" %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://alist.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URIs in the format of
    `https://alist.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/sso_callback?method=sso_get_token` and
    `https://alist.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/sso_callback?method=get_sso_id`.
    This means if you change this value, you need to update the redirect URIs.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `alist`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [AList] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'alist'
        client_name: 'AList'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://alist.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/sso_callback?method=sso_get_token'
          - 'https://alist.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/sso_callback?method=get_sso_id'
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

{{% oidc-escape-hatch-claims-hydration client_id="alist" claims="preferred_username" %}}

### Application

To configure [AList] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [AList] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [AList] as an administrator.
2. Navigate to Manage.
3. Navigate to Settings.
4. Navigate to SSO.
5. Configure the following options:
   - SSO login enabled: Enabled
   - SSO login platform: `OIDC`
   - SSO client id: `alist`
   - SSO client secret: `insecure_secret`
   - SSO oidc username key: `preferred_username`
   - SSO endpoint name: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
   - SSO compatibility mode: Disabled
   - SSO auto register: Enable if you want users to automatically be created in [AList].
6. Click `Save`.
7. Navigate to Profile and click the button to bind the current account to the SSO account.

## See Also

- [AList Single Sign-On Documentation](https://alistgo.com/guide/advanced/sso.html)

[Authelia]: https://www.authelia.com
[AList]: https://alistgo.com/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
