---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Keycloak"
description: "A guide on integrating Keycloak with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Keycloak | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Keycloak with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Keycloak]
  - [v26.8.2](https://github.com/keycloak/keycloak/releases/tag/26.8.2)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://keycloak.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://keycloak.{{< sitevar name="domain" nojs="example.com" >}}/realms/example/broker/authelia/endpoint`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `keycloak`
- **Client Secret:** `insecure_secret`
- **Keycloak Realm:** `example`
  - This option determines the redirect URI. This means if you change this value, you need to update the redirect URI.
- **Keycloak Identity Provider Alias:** `authelia`
  - This option determines the redirect URI. This means if you change this value, you need to update the redirect URI.

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Keycloak] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'keycloak'
        client_name: 'Keycloak'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://keycloak.{{< sitevar name="domain" nojs="example.com" >}}/realms/example/broker/authelia/endpoint'
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

To configure [Keycloak] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [Keycloak] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to the [Keycloak] Admin Console as an administrator.
2. Select the `example` realm.
3. Navigate to `Identity providers`.
4. Select `OpenID Connect v1.0`.
5. Configure the following options:
   - Alias: `authelia`
   - Display name: `Authelia`
   - Use discovery endpoint: Toggle to the on position
   - Discovery endpoint: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration`
   - Client authentication: `Client secret sent as basic auth`
   - Client ID: `keycloak`
   - Client Secret: `insecure_secret`
6. Ensure the `Redirect URI` displayed in the form matches the redirect URI configured in Authelia.
7. Click `Add`.
8. Configure the following options:
   - Use PKCE: Toggle to the on position
   - PKCE Method: `S256`
   - Scopes (under `Advanced`): `openid email profile`
9. Click `Save`.

## See Also

- [Keycloak Integrating Identity Providers Documentation](https://www.keycloak.org/docs/latest/server_admin/index.html#_identity_broker)
- [Keycloak OpenID Connect v1.0 Identity Providers Documentation](https://www.keycloak.org/docs/latest/server_admin/index.html#_identity_broker_oidc)

[Authelia]: https://www.authelia.com
[Keycloak]: https://www.keycloak.org/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
