---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Documenso"
description: "A guide on integrating Documenso with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Documenso | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Documenso with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Documenso]
  - [v2.21.1](https://github.com/documenso/documenso/releases/tag/v2.21.1)

{{% oidc-common bugs="client-credentials-encoding,claims-hydration" %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://documenso.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://documenso.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/callback/oidc`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `documenso`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Documenso] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'documenso'
        client_name: 'Documenso'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://documenso.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/callback/oidc'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

#### Configuration Escape Hatch

{{% oidc-escape-hatch-claims-hydration client_id="documenso" claims="email,name" %}}

### Application

To configure [Documenso] there is one method, using [Environment Variables](#environment-variables).

#### Environment Variables

To configure [Documenso] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
NEXT_PUBLIC_WEBAPP_URL=https://documenso.{{< sitevar name="domain" nojs="example.com" >}}
NEXT_PRIVATE_OIDC_WELL_KNOWN=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration
NEXT_PRIVATE_OIDC_CLIENT_ID=documenso
NEXT_PRIVATE_OIDC_CLIENT_SECRET=insecure_secret
NEXT_PRIVATE_OIDC_PROVIDER_LABEL=Authelia
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  documenso:
    environment:
      NEXT_PUBLIC_WEBAPP_URL: 'https://documenso.{{< sitevar name="domain" nojs="example.com" >}}'
      NEXT_PRIVATE_OIDC_WELL_KNOWN: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration'
      NEXT_PRIVATE_OIDC_CLIENT_ID: 'documenso'
      NEXT_PRIVATE_OIDC_CLIENT_SECRET: 'insecure_secret'
      NEXT_PRIVATE_OIDC_PROVIDER_LABEL: 'Authelia'
```

## See Also

- [Documenso Environment Variables Documentation](https://docs.documenso.com/docs/self-hosting/configuration/environment)

[Authelia]: https://www.authelia.com
[Documenso]: https://documenso.com/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
