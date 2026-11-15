---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Fluxer"
description: "A guide on integrating Fluxer with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Fluxer | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Fluxer with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Fluxer]
  - [2026.1112.20122](https://github.com/fluxerapp/fluxer/releases/tag/fluxer-release%402026.1112.20122)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://fluxer.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://fluxer.{{< sitevar name="domain" nojs="example.com" >}}/auth/sso/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `fluxer`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Fluxer] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'fluxer'
        client_name: 'Fluxer'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://fluxer.{{< sitevar name="domain" nojs="example.com" >}}/auth/sso/callback'
        scopes:
          - 'openid'
          - 'email'
          - 'profile'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

To configure [Fluxer] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
[Fluxer] only accepts identity provider URLs that use HTTPS and resolve to a public address. If the Authelia domain
resolves to a private address, set `FLUXER_SSO_ALLOW_PRIVATE_ADDRESSES=true` in the [Fluxer] environment (or run
`rexctl config set sso.allow_private_addresses true` on an instance managed by `rexctl`) and restart [Fluxer].
{{< /callout >}}

To configure [Fluxer] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to the [Fluxer] admin panel at `https://fluxer.{{< sitevar name="domain" nojs="example.com" >}}/admin` as an
   administrator.
2. Navigate to `Configuration` > `Instance Config`.
3. Locate the `Single Sign-On (SSO)` card.
4. Configure the following options:
   - Enable SSO: Enabled
   - Automatically provision users on first SSO login: Enable this option if users who do not already exist in
     [Fluxer] should be created on their first login.
   - Issuer: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
   - Display Name: `Authelia`
   - Client ID: `fluxer`
   - Client Secret: `insecure_secret`
   - Scope: `openid email profile`
5. Ensure the `Redirect URI` field shows `https://fluxer.{{< sitevar name="domain" nojs="example.com" >}}/auth/sso/callback`.
6. Click `Save SSO Settings`.

## See Also

- [Fluxer Single Sign-On Documentation](https://docs.fluxer.com/operator/sign-in/#single-sign-on)

[Authelia]: https://www.authelia.com
[Fluxer]: https://fluxer.app/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
