---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Lychee"
description: "A guide on integrating Lychee with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Lychee | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Lychee with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Lychee]
  - [v7.13.0](https://github.com/LycheeOrg/Lychee/releases/tag/v7.13.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://lychee.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://lychee.{{< sitevar name="domain" nojs="example.com" >}}/auth/authelia/redirect`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `lychee`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Lychee] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'lychee'
        client_name: 'Lychee'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://lychee.{{< sitevar name="domain" nojs="example.com" >}}/auth/authelia/redirect'
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

To configure [Lychee] there is one method, using [Environment Variables](#environment-variables).

#### Environment Variables

To configure [Lychee] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
AUTHELIA_BASE_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
AUTHELIA_CLIENT_ID=lychee
AUTHELIA_CLIENT_SECRET=insecure_secret
AUTHELIA_REDIRECT_URI=https://lychee.{{< sitevar name="domain" nojs="example.com" >}}/auth/authelia/redirect
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  lychee:
    environment:
      AUTHELIA_BASE_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      AUTHELIA_CLIENT_ID: 'lychee'
      AUTHELIA_CLIENT_SECRET: 'insecure_secret'
      AUTHELIA_REDIRECT_URI: 'https://lychee.{{< sitevar name="domain" nojs="example.com" >}}/auth/authelia/redirect'
```

#### Linking Accounts

By default [Lychee] only allows users to sign in with Authelia after they have linked their existing [Lychee] account.
To link an account, sign in to [Lychee] with a username and password, navigate to the profile page, and click the
Authelia button in the OAuth section. Alternatively enable the `oauth_create_user_on_first_attempt` setting in the
[Lychee] settings to automatically create users on their first sign in.

## See Also

- [Lychee Configuration Documentation](https://lycheeorg.dev/docs/getting-started/configuration/)

[Authelia]: https://www.authelia.com
[Lychee]: https://lycheeorg.dev/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
