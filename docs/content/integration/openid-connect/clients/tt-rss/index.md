---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Tiny Tiny RSS"
description: "A guide on integrating Tiny Tiny RSS with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Tiny Tiny RSS | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Tiny Tiny RSS with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Tiny Tiny RSS]
  - [59272c0](https://github.com/tt-rss/tt-rss/commit/59272c09926e7183249191f9c89b4b1bc772a878)
- [Tiny Tiny RSS OIDC Plugin]
  - [9ef6658](https://github.com/tt-rss/tt-rss-plugin-auth-oidc/commit/9ef6658d3d613268b1de3e383c4b487aee4b78ac)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://tt-rss.{{< sitevar name="domain" nojs="example.com" >}}/tt-rss/`
  - This option determines the redirect URI in the format of
    `https://tt-rss.{{< sitevar name="domain" nojs="example.com" >}}/tt-rss`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `tt-rss`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Tiny Tiny RSS] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'tt-rss'
        client_name: 'Tiny Tiny RSS'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://tt-rss.{{< sitevar name="domain" nojs="example.com" >}}/tt-rss'
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
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [Tiny Tiny RSS] there is one method, using [Environment Variables](#environment-variables).

#### Plugin Installation

The OpenID Connect 1.0 support for [Tiny Tiny RSS] is provided by the official [Tiny Tiny RSS OIDC Plugin] which is
not included in the container image. Clone it into the `plugins.local` directory of the installation as `auth_oidc`,
for example:

```shell
git clone https://github.com/tt-rss/tt-rss-plugin-auth-oidc.git plugins.local/auth_oidc
```

The plugin is a system plugin, so it has to be enabled globally via the `TTRSS_PLUGINS` environment variable as shown
below.

#### Environment Variables

To configure [Tiny Tiny RSS] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
TTRSS_SELF_URL_PATH=https://tt-rss.{{< sitevar name="domain" nojs="example.com" >}}/tt-rss/
TTRSS_PLUGINS=auth_internal,auth_oidc,note,nginx_xaccel
TTRSS_AUTH_OIDC_NAME=Authelia
TTRSS_AUTH_OIDC_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
TTRSS_AUTH_OIDC_CLIENT_ID=tt-rss
TTRSS_AUTH_OIDC_CLIENT_SECRET=insecure_secret
TTRSS_AUTH_OIDC_CLIENT_USERNAME_CLAIM=preferred_username
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  app:
    environment:
      TTRSS_SELF_URL_PATH: 'https://tt-rss.{{< sitevar name="domain" nojs="example.com" >}}/tt-rss/'
      TTRSS_PLUGINS: 'auth_internal,auth_oidc,note,nginx_xaccel'
      TTRSS_AUTH_OIDC_NAME: 'Authelia'
      TTRSS_AUTH_OIDC_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      TTRSS_AUTH_OIDC_CLIENT_ID: 'tt-rss'
      TTRSS_AUTH_OIDC_CLIENT_SECRET: 'insecure_secret'
      TTRSS_AUTH_OIDC_CLIENT_USERNAME_CLAIM: 'preferred_username'
```

## See Also

- [Tiny Tiny RSS OIDC Plugin Documentation](https://github.com/tt-rss/tt-rss-plugin-auth-oidc)

[Authelia]: https://www.authelia.com
[Tiny Tiny RSS]: https://tt-rss.org/
[Tiny Tiny RSS OIDC Plugin]: https://github.com/tt-rss/tt-rss-plugin-auth-oidc
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
