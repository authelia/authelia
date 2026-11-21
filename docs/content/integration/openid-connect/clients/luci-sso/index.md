---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "LuCI SSO"
description: "A guide on integrating LuCI SSO with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "LuCI SSO | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring LuCI SSO with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [LuCI SSO]
  - [v0.10.0](https://github.com/m00qek/luci-sso/releases/tag/v0.10.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://openwrt.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://openwrt.{{< sitevar name="domain" nojs="example.com" >}}/cgi-bin/luci-sso/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `luci-sso`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [LuCI SSO] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'luci-sso'
        client_name: 'LuCI SSO'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://openwrt.{{< sitevar name="domain" nojs="example.com" >}}/cgi-bin/luci-sso/callback'
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
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

To configure [LuCI SSO] there are two methods, using the [Web GUI](#web-gui), or using the [CLI](#cli).

#### Web GUI

To configure [LuCI SSO] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [LuCI].
2. Navigate to Services.
3. Navigate to Single Sign-On.
4. Configure the following options in the Identity provider section:
   - Issuer URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
   - Client ID: `luci-sso`
   - Client Secret: `insecure_secret`
   - Redirect URI: `https://openwrt.{{< sitevar name="domain" nojs="example.com" >}}/cgi-bin/luci-sso/callback`
   - Scopes: `openid profile email groups`
   - Enable SSO: Enabled
5. Click `Test connection` and ensure every check passes.
6. Click `Save & Apply`.
7. In the Roles section, edit the `admin` role and add the email addresses or groups of the users who should be
   granted access.
8. Click `Save & Apply`.

#### CLI

To configure [LuCI SSO] to utilize Authelia as an [OpenID Connect 1.0] Provider, run the following commands on the
router:

```shell
uci set luci-sso.default.issuer_url='https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
uci set luci-sso.default.client_id='luci-sso'
uci set luci-sso.default.client_secret='insecure_secret'
uci set luci-sso.default.redirect_uri='https://openwrt.{{< sitevar name="domain" nojs="example.com" >}}/cgi-bin/luci-sso/callback'
uci set luci-sso.default.scope='openid profile email groups'
uci set luci-sso.default.enabled='1'
uci add_list luci-sso.admin.group='admins'
uci commit luci-sso
```

## See Also

- [LuCI SSO Authelia Documentation](https://m00qek.github.io/luci-sso/latest/how-to/providers/authelia/)

[Authelia]: https://www.authelia.com
[LuCI SSO]: https://github.com/m00qek/luci-sso
[LuCI]: https://openwrt.org/docs/guide-user/luci/start
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
