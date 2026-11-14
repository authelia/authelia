---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Quay"
description: "A guide on integrating Quay with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Quay | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Quay with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Quay]
  - [v3.18.2](https://github.com/quay/quay/releases/tag/v3.18.2)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://quay.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://quay.{{< sitevar name="domain" nojs="example.com" >}}/oauth2/authelia/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `quay`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Quay] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'quay'
        client_name: 'Quay'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://quay.{{< sitevar name="domain" nojs="example.com" >}}/oauth2/authelia/callback'
          - 'https://quay.{{< sitevar name="domain" nojs="example.com" >}}/oauth2/authelia/callback/attach'
          - 'https://quay.{{< sitevar name="domain" nojs="example.com" >}}/oauth2/authelia/callback/cli'
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
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

To configure [Quay] there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
The `AUTHELIA` prefix of the `AUTHELIA_LOGIN_CONFIG` key determines the `authelia` portion of the redirect URI. If you
change this prefix you need to update the redirect URIs.
{{< /callout >}}

To configure [Quay] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```yaml {title="config.yaml"}
AUTHELIA_LOGIN_CONFIG:
  SERVICE_NAME: 'Authelia'
  OIDC_SERVER: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/'
  CLIENT_ID: 'quay'
  CLIENT_SECRET: 'insecure_secret'
  LOGIN_SCOPES:
    - 'openid'
    - 'profile'
    - 'email'
    - 'groups'
  PREFERRED_USERNAME_CLAIM_NAME: 'preferred_username'
  PREFERRED_GROUP_CLAIM_NAME: 'groups'
  USE_PKCE: true
  PKCE_METHOD: 'S256'
```

## See Also

- [Project Quay OIDC Configuration Fields Documentation](https://docs.projectquay.io/config_quay.html)

[Authelia]: https://www.authelia.com
[Quay]: https://www.projectquay.io/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
