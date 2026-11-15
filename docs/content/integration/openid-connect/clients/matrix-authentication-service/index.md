---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Matrix Authentication Service"
description: "A guide on integrating Matrix Authentication Service with the Authelia OpenID Connect 1.0 Provider for SSO, with configuration examples and available options."
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
  title: "Matrix Authentication Service | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Matrix Authentication Service with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Matrix Authentication Service]
  - [v1.29.0](https://github.com/element-hq/matrix-authentication-service/releases/tag/v1.29.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://mas.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://mas.{{< sitevar name="domain" nojs="example.com" >}}/upstream/callback/01M63F1HX0HD2YZ0XR6CPJ6KRM`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `matrix-authentication-service`
- **Client Secret:** `insecure_secret`
- **Upstream Provider ID:** `01M63F1HX0HD2YZ0XR6CPJ6KRM`
  - This value is a [ULID](https://github.com/ulid/spec) which uniquely identifies the provider within
    [Matrix Authentication Service] and is part of the redirect URI. You should generate your own value.

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with
[Matrix Authentication Service] which will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'matrix-authentication-service'
        client_name: 'Matrix Authentication Service'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://mas.{{< sitevar name="domain" nojs="example.com" >}}/upstream/callback/01M63F1HX0HD2YZ0XR6CPJ6KRM'
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

### Application

To configure [Matrix Authentication Service] there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

To configure [Matrix Authentication Service] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following
configuration:

```yaml {title="config.yaml"}
upstream_oauth2:
  providers:
    - id: '01M63F1HX0HD2YZ0XR6CPJ6KRM'
      human_name: 'Authelia'
      issuer: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      client_id: 'matrix-authentication-service'
      client_secret: 'insecure_secret'
      token_endpoint_auth_method: 'client_secret_basic'
      scope: 'openid profile email'
      fetch_userinfo: true
      claims_imports:
        localpart:
          action: 'require'
          template: '{{ user.preferred_username }}'
        displayname:
          action: 'suggest'
          template: '{{ user.name }}'
        email:
          action: 'suggest'
          template: '{{ user.email }}'
```

After modifying the configuration run `mas-cli config sync` to synchronize the upstream providers with the database.

## See Also

- [Matrix Authentication Service Configure an upstream SSO provider Documentation](https://element-hq.github.io/matrix-authentication-service/setup/sso.html)
- [Matrix Authentication Service Configuration Reference](https://element-hq.github.io/matrix-authentication-service/reference/configuration.html#upstream_oauth2)

[Authelia]: https://www.authelia.com
[Matrix Authentication Service]: https://element-hq.github.io/matrix-authentication-service/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
