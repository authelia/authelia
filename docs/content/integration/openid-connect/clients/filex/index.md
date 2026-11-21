---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "filex"
description: "A guide on integrating filex with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "filex | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring filex with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [filex]
  - [v0.60.0](https://github.com/BRF-Tech/filex/releases/tag/v0.60.0)

{{% oidc-common bugs="claims-hydration" %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://filex.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://filex.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/oidc/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `filex`
- **Client Secret:** `insecure_secret`
- **Administrators Group:** `filex-admins`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [filex] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'filex'
        client_name: 'filex'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://filex.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/oidc/callback'
        scopes:
          - 'openid'
          - 'email'
          - 'profile'
          - 'groups'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

#### Configuration Escape Hatch

{{% oidc-escape-hatch-claims-hydration client_id="filex" claims="email,email_verified,groups" %}}

### Application

To configure [filex] there are two methods, using [Environment Variables](#environment-variables) or using the
[Web GUI](#web-gui).

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
[filex] builds the redirect URI from the `FILEX_PUBLIC_URL` environment variable, which must be set to
`https://filex.{{< sitevar name="domain" nojs="example.com" >}}` regardless of the method used. In addition [filex]
only trusts the email address of a user when the `email_verified` claim is `true`, otherwise new accounts are created
disabled until an administrator enables them.
{{< /callout >}}

#### Environment Variables

To configure [filex] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
FILEX_PUBLIC_URL=https://filex.{{< sitevar name="domain" nojs="example.com" >}}
FILEX_AUTH_DRIVERS=local,oidc
FILEX_OIDC_ISSUER=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
FILEX_OIDC_CLIENT_ID=filex
FILEX_OIDC_CLIENT_SECRET=insecure_secret
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  filex:
    environment:
      FILEX_PUBLIC_URL: 'https://filex.{{< sitevar name="domain" nojs="example.com" >}}'
      FILEX_AUTH_DRIVERS: 'local,oidc'
      FILEX_OIDC_ISSUER: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      FILEX_OIDC_CLIENT_ID: 'filex'
      FILEX_OIDC_CLIENT_SECRET: 'insecure_secret'
```

#### Web GUI

To configure [filex] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Ensure the `FILEX_SECRET_KEY` environment variable is set, as it is required to store the client secret.
2. Login to [filex] as an administrator and visit `https://filex.{{< sitevar name="domain" nojs="example.com" >}}/admin`.
3. Navigate to `Settings`, `Identity providers`.
4. Select the `OIDC` tab and click `OIDC (Keycloak, Auth0, Authentik…)`.
5. Configure the following options:
   - Enabled: Enabled
   - Issuer URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
   - Client ID: `filex`
   - Client secret: `insecure_secret`
   - Extra scopes: `groups`
   - Role claim: `groups`
   - Admin group / role: `filex-admins`
6. Press `Test now` to check the configuration.
7. Press `Save and apply`.

## See Also

- [filex Single Sign-On (OIDC) Documentation](https://docs.filex.sh/SSO)
- [filex Configuration Documentation](https://docs.filex.sh/CONFIGURATION)

[Authelia]: https://www.authelia.com
[filex]: https://filex.sh/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
