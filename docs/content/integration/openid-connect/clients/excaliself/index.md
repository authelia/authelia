---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Excaliself"
description: "A guide on integrating Excaliself with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
summary: ""
date: 2026-11-14T12:00:00+00:00
draft: false
images: []
weight: 620
toc: true
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Excaliself | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Excaliself with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Excaliself]
  - [v0.12.0](https://github.com/maciekzygmunt/excaliself/releases/tag/v0.12.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://excaliself.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://excaliself.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/callback/oidc`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `excaliself`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Excaliself] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'excaliself'
        client_name: 'Excaliself'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://excaliself.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/callback/oidc'
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

Notes:

- [Excaliself] sends the client credentials in the request body, so the `token_endpoint_auth_method` must be
  `client_secret_post`. It always uses PKCE with the `S256` challenge method.
- [Excaliself] reads the `email` and `name` claims from the UserInfo Endpoint when the ID Token does not contain the
  `email` claim, so no claims policy is needed.

### Application

To configure [Excaliself] there is one method, using [Environment Variables](#environment-variables).

#### Environment Variables

To configure [Excaliself] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
BASE_URL=https://excaliself.{{< sitevar name="domain" nojs="example.com" >}}
OIDC_ISSUER=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
OIDC_CLIENT_ID=excaliself
OIDC_CLIENT_SECRET=insecure_secret
OIDC_NAME=Authelia
SSO_ALLOWED_DOMAINS={{< sitevar name="domain" nojs="example.com" >}}
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  excaliself:
    environment:
      BASE_URL: 'https://excaliself.{{< sitevar name="domain" nojs="example.com" >}}'
      OIDC_ISSUER: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      OIDC_CLIENT_ID: 'excaliself'
      OIDC_CLIENT_SECRET: 'insecure_secret'
      OIDC_NAME: 'Authelia'
      SSO_ALLOWED_DOMAINS: '{{< sitevar name="domain" nojs="example.com" >}}'
```

Restart [Excaliself] after you change these variables. The sign-in page then shows a **Continue with Authelia** button.
`OIDC_NAME` sets the label of this button.

#### Who Can Sign In

- **Existing members:** a member whose email matches the email of the Authelia user is signed in to their existing
  account. The first sign-in links the account to Authelia and removes its password.
- **New users:** an account is created on the first sign-in only if the domain of the email is in
  `SSO_ALLOWED_DOMAINS`, which is a comma-separated list. Without this variable, new users need an invite link from an
  [Excaliself] admin.

## See Also

- [Excaliself Single Sign-On Documentation](https://excaliself.com/docs/single-sign-on/)

[Authelia]: https://www.authelia.com
[Excaliself]: https://excaliself.com
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
