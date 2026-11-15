---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Vaultwarden"
description: "A guide on integrating Vaultwarden with the Authelia OpenID Connect 1.0 Provider with configuration examples and an outline of the available options for SSO."
summary: ""
date: 2025-06-15T09:27:11+00:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/vaultwarden/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Vaultwarden | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Vaultwarden with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Vaultwarden]
  - [1.35.0](https://github.com/dani-garcia/vaultwarden/releases/tag/1.35.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://vault.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://vault.{{< sitevar name="domain" nojs="example.com" >}}/identity/connect/oidc-signin`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `vaultwarden`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Vaultwarden] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
    - client_id: 'vaultwarden'
      client_name: 'Vaultwarden'
      client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
      public: false
      application_type: 'web'
      authorization_policy: 'two_factor'
      require_pkce: true
      pkce_challenge_method: 'S256'
      redirect_uris:
        - 'https://vault.{{< sitevar name="domain" nojs="example.com" >}}/identity/connect/oidc-signin'
      scopes:
        - 'openid'
        - 'offline_access'
        - 'profile'
        - 'email'
      response_types:
        - 'code'
      grant_types:
        - 'authorization_code'
        - 'refresh_token'
      access_token_signed_response_alg: 'none'
      userinfo_signed_response_alg: 'none'
      token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [Vaultwarden] there is one method, using the [Environment Variables](#environment-variables).

#### Environment Variables

To configure [Vaultwarden] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
SSO_ENABLED=true
SSO_ONLY=false
SSO_AUTHORITY=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
SSO_SCOPES=profile email offline_access
SSO_PKCE=true
SSO_CLIENT_ID=vaultwarden
SSO_CLIENT_SECRET=insecure_secret
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  vaultwarden:
    environment:
      - SSO_ENABLED=true
      - SSO_ONLY=false
      - SSO_AUTHORITY=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
      - SSO_SCOPES=profile email offline_access
      - SSO_PKCE=true
      - SSO_CLIENT_ID=vaultwarden
      - SSO_CLIENT_SECRET=insecure_secret
```

### OIDCWarden Role Mapping

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
Role mapping via the `SSO_ROLES_*` environment variables is only available in the [OIDCWarden] fork and is not
supported by [Vaultwarden] itself. Skip this section if you are using [Vaultwarden].
{{< /callout >}}

{{< callout context="tip" title="Did you know?" icon="outline/rocket" >}}
The `vaultwarden_roles` user attribute renders the value `["admin"]` if the user is in the `vaultwarden_admins` group
within Authelia, renders the value `["user"]` if they are in the `vaultwarden_users` group, otherwise it renders `[""]`.
You can adjust this to your preference to assign a role to the appropriate user groups.
{{< /callout >}}

To map roles with [OIDCWarden], add the following to the **Authelia** configuration above, then add
`claims_policy: 'vaultwarden'` and the `vaultwarden` scope to the client:

```yaml {title="configuration.yml"}
definitions:
  user_attributes:
    vaultwarden_roles:
      expression: '"vaultwarden_admins" in groups ? ["admin"] : "vaultwarden_users" in groups ? ["user"] : [""]'
identity_providers:
  oidc:
    claims_policies:
      vaultwarden:
        id_token: ['vaultwarden_roles']
        custom_claims:
          vaultwarden_roles: {}
    scopes:
      vaultwarden:
        claims: ['vaultwarden_roles']
```

Then add the `vaultwarden` scope to `SSO_SCOPES` and set the following additional environment variables:

```shell {title=".env"}
SSO_SCOPES=profile email offline_access vaultwarden
SSO_ROLES_ENABLED=true
SSO_ROLES_DEFAULT_TO_USER=true
SSO_ROLES_TOKEN_PATH=/vaultwarden_roles
```

## See Also

- [Vaultwarden SSO using OpenID Connect](https://github.com/dani-garcia/vaultwarden/wiki/Enabling-SSO-support-using-OpenId-Connect)
- [OIDCWarden SSO using OpenID Connect](https://github.com/Timshel/OIDCWarden/blob/main/SSO.md)

[Authelia]: https://www.authelia.com
[Vaultwarden]: https://github.com/dani-garcia/vaultwarden
[OIDCWarden]: https://github.com/Timshel/OIDCWarden
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
