---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "LubeLogger"
description: "A guide on integrating LubeLogger with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "LubeLogger | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring LubeLogger with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [LubeLogger]
  - [v1.7.4](https://github.com/hargata/lubelog/releases/tag/v1.7.4)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://lubelogger.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://lubelogger.{{< sitevar name="domain" nojs="example.com" >}}/Login/RemoteAuth`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `lubelogger`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [LubeLogger] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'lubelogger'
        client_name: 'LubeLogger'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://lubelogger.{{< sitevar name="domain" nojs="example.com" >}}/Login/RemoteAuth'
        scopes: ['openid', 'email']
        response_types: ['code']
        grant_types: ['authorization_code']
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

To configure [LubeLogger] there are two methods, using [Environment Variables](#environment-variables), or using the
[Web GUI](#web-gui).

#### Environment Variables

To configure [LubeLogger] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
OpenIDConfig__Name=Authelia
OpenIDConfig__ClientId=lubelogger
OpenIDConfig__ClientSecret=insecure_secret
OpenIDConfig__AuthURL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization
OpenIDConfig__TokenURL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token
OpenIDConfig__UserInfoURL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo
OpenIDConfig__JwksURL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/jwks.json
OpenIDConfig__RedirectURL=https://lubelogger.{{< sitevar name="domain" nojs="example.com" >}}/Login/RemoteAuth
OpenIDConfig__Scope=openid email
OpenIDConfig__ValidateState=true
OpenIDConfig__UsePKCE=true
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  lubelogger:
    environment:
      OpenIDConfig__Name: 'Authelia'
      OpenIDConfig__ClientId: 'lubelogger'
      OpenIDConfig__ClientSecret: 'insecure_secret'
      OpenIDConfig__AuthURL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization'
      OpenIDConfig__TokenURL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token'
      OpenIDConfig__UserInfoURL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo'
      OpenIDConfig__JwksURL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/jwks.json'
      OpenIDConfig__RedirectURL: 'https://lubelogger.{{< sitevar name="domain" nojs="example.com" >}}/Login/RemoteAuth'
      OpenIDConfig__Scope: 'openid email'
      OpenIDConfig__ValidateState: 'true'
      OpenIDConfig__UsePKCE: 'true'
```

#### Web GUI

{{< callout context="tip" title="Help Wanted" icon="outline/rocket" >}}
We would love screenshots of this configuration!
{{< /callout >}}

To configure [LubeLogger] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [LubeLogger] as the root user.
2. Click the user icon and select `Configure`.
3. Click `Acknowledge and Continue`.
4. Click `Next` until the `Single Sign On` page is displayed.
5. Configure the following options:
   - OIDC Provider: `Authelia`
   - OIDC Client ID: `lubelogger`
   - OIDC Client Secret: `insecure_secret`
   - OIDC Auth URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization`
   - OIDC Token URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token`
   - OIDC UserInfo URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo`
   - OIDC JWKS URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/jwks.json`
   - OIDC Redirect URL: `https://lubelogger.{{< sitevar name="domain" nojs="example.com" >}}/Login/RemoteAuth`
   - OIDC Scope: `openid email`
   - OIDC Validate State: Enabled
   - OIDC Use PKCE: Enabled
6. Click `Next`.
7. Optionally enable `Enable OIDC for Root User` and enter the email address of the Authelia user who should be the root
   user.
8. Click `Save`.

## See Also

- [LubeLogger Authenticating via OpenID Connect Documentation](https://docs.lubelogger.com/Advanced/OpenID/)

[Authelia]: https://www.authelia.com
[LubeLogger]: https://lubelogger.com/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
