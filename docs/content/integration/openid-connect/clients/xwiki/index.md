---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "XWiki"
description: "A guide on integrating XWiki with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "XWiki | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring XWiki with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [XWiki] OpenID Connect Authenticator
  - [v2.30.0](https://github.com/xwiki-contrib/oidc/releases/tag/oidc-2.30.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://xwiki.{{< sitevar name="domain" nojs="example.com" >}}/xwiki/`
  - This option determines the redirect URI in the format of
    `https://xwiki.{{< sitevar name="domain" nojs="example.com" >}}/xwiki/oidc/authenticator/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `xwiki`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [XWiki] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'xwiki'
        client_name: 'XWiki'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://xwiki.{{< sitevar name="domain" nojs="example.com" >}}/xwiki/oidc/authenticator/callback'
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
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [XWiki] there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

To configure [XWiki] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Install the `OpenID Connect Authenticator` extension using the Extension Manager.
2. Add the following option to the `xwiki.cfg` configuration file:

   ```ini {title="xwiki.cfg"}
   xwiki.authentication.authclass=org.xwiki.contrib.oidc.auth.OIDCAuthServiceImpl
   ```

3. Add the following options to the `xwiki.properties` configuration file:

   ```ini {title="xwiki.properties"}
   oidc.provider=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
   oidc.clientid=xwiki
   oidc.secret=insecure_secret
   oidc.scope=openid,profile,email,groups
   oidc.endpoint.token.auth_method=client_secret_basic
   oidc.groups.claim=groups
   ```

4. Restart [XWiki].

The `oidc.groups.claim` option is only required if you want to synchronize Authelia groups with [XWiki] groups. See the
`oidc.groups.mapping` option in the [XWiki OpenID Connect Authenticator Documentation] for more information.

## See Also

- [XWiki OpenID Connect Authenticator Documentation]

[Authelia]: https://www.authelia.com
[XWiki]: https://www.xwiki.org/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
[XWiki OpenID Connect Authenticator Documentation]: https://extensions.xwiki.org/xwiki/bin/view/Extension/OpenID%20Connect/OpenID%20Connect%20Authenticator/
