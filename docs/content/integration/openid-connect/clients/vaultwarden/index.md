---
title: "Vaultwarden"
description: "Integrating Vaultwarden with the Authelia OpenID Connect 1.0 Provider."
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
  - [v4.39.6](https://github.com/authelia/authelia/releases/tag/v4.39.6)
- [Vaultwarden]
  - [oidcwarden development fork v2025.5.1-5](https://github.com/Timshel/OIDCWarden/releases/tag/v2025.5.1-5)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://vault.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://vault.{{< sitevar name="domain" nojs="example.com" >}}/identity/connect/oidc-signin`.
    This means if you change this value, you need to update the redirect URI.
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `vaultwarden`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

{{< callout context="tip" title="Did you know?" icon="outline/rocket" >}}
The `vaultwarden_roles` user attribute renders the value `["admin"]` if the user is in the `vaultwarden_admins` group
within Authelia, renders the value `["user"]` if they are in the `vaultwarden_users` group, otherwise it renders `""`.
You can adjust this to your preference to assign a role to the appropriate user groups.
{{< /callout >}}

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Vaultwarden] which
will operate with the application example:

```yaml {title="configuration.yml"}
definitions:
  user_attributes:
    vaultwarden_roles:
      expression: '"vaultwarden_admins" in groups ? ["admin"] : "vaultwarden_users" in groups ? ["user"] : [""]'
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    claims_policies:
      vaultwarden:
        id_token: ['vaultwarden_roles']
        custom_claims:
          vaultwarden_roles: {}
    scopes:
      vaultwarden:
        claims: ['vaultwarden_roles']
    clients:
    - client_id: 'vaultwarden'
      client_name: 'Vaultwarden'
      client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
      public: false
      authorization_policy: 'two_factor'
      require_pkce: true
      pkce_challenge_method: 'S256'
      claims_policy: 'vaultwarden'
      redirect_uris:
        - 'https://vault.{{< sitevar name="domain" nojs="example.com" >}}/identity/connect/oidc-signin'
      scopes:
        - 'openid'
        - 'offline_access'
        - 'profile'
        - 'email'
        - 'vaultwarden'
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
SSO_SCOPES=profile email offline_access vaultwarden
SSO_PKCE=true
SSO_CLIENT_ID=vaultwarden
SSO_CLIENT_SECRET=insecure_secret
SSO_ROLES_ENABLED=true
SSO_ROLES_DEFAULT_TO_USER=true
SSO_ROLES_TOKEN_PATH=/vaultwarden_roles
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  vaultwarden:
    environment:
      - SSO_ENABLED=true
      - SSO_ONLY=false
      - SSO_AUTHORITY=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
      - SSO_SCOPES=profile email offline_access vaultwarden
      - SSO_PKCE=true
      - SSO_CLIENT_ID=vaultwarden
      - SSO_CLIENT_SECRET=insecure_secret
      - SSO_ROLES_ENABLED=true
      - SSO_ROLES_DEFAULT_TO_USER=true
      - SSO_ROLES_TOKEN_PATH=/vaultwarden_roles
```


## See Also

- [SSO using OpenId Connect](https://github.com/Timshel/OIDCWarden/blob/main/SSO.md)

[Authelia]: https://www.authelia.com
[Vaultwarden]: https://github.com/Timshel/OIDCWarden/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
