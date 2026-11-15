---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "wg-access-server"
description: "A guide on integrating wg-access-server with the Authelia OpenID Connect 1.0 Provider for SSO, with configuration examples and available options."
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
  title: "wg-access-server | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring wg-access-server with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [wg-access-server]
  - [v1.3.0](https://github.com/freifunkMUC/wg-access-server/releases/tag/v1.3.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://wg-access-server.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://wg-access-server.{{< sitevar name="domain" nojs="example.com" >}}/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `wg-access-server`
- **Client Secret:** `insecure_secret`
- **Admin Group:** `wg-access-server-admins`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [wg-access-server]
which will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'wg-access-server'
        client_name: 'wg-access-server'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://wg-access-server.{{< sitevar name="domain" nojs="example.com" >}}/callback'
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

To configure [wg-access-server] there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

To configure [wg-access-server] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following
configuration:

```yaml {title="config.yaml"}
auth:
  oidc:
    name: 'Authelia'
    issuer: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
    clientID: 'wg-access-server'
    clientSecret: 'insecure_secret'
    redirectURL: 'https://wg-access-server.{{< sitevar name="domain" nojs="example.com" >}}/callback'
    scopes:
      - 'openid'
      - 'profile'
      - 'email'
      - 'groups'
    claimMapping:
      admin: "'wg-access-server-admins' in groups"
```

## See Also

- [wg-access-server Authentication Documentation](https://github.com/freifunkMUC/wg-access-server/blob/master/docs/auth.md)

[Authelia]: https://www.authelia.com
[wg-access-server]: https://github.com/freifunkMUC/wg-access-server
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
