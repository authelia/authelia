---
title: "Oidcwarden"
description: "Integrating Oidcwarden with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-06-10T10:51:47+10:00
draft: false
images: []
weight: 620
toc: true
support:
  level: community
  versions: true
  integration: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.4](https://github.com/authelia/authelia/releases/tag/v4.39.4)
- [Oidcwarden]
  - [v2025.5.1-5](https://github.com/Timshel/OIDCWarden/releases/tag/v2025.5.1-5)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://vault.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://vault.{{< sitevar name="domain" nojs="example.com" >}}/identity/connect/oidc-signin`.
    This means if you change this value, you need to update the redirect URI.
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `oidcwarden`
- __Client Secret:__ `insecure_secret`
- __Groups:__ `vaultwarden_users` for users and `vaultwarden_admins` for admins

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Oidcwarden] which
will operate with the application example:

```yaml {title="configuration.yml"}
definitions:
  user_attributes:
    vault_roles:
      expression: '"vaultwarden_admins" in groups ? ["admin"] : "vaultwarden_users" in groups ? ["user"] : [""]'
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    claims_policies:
      vaultwarden:
        id_token: ['oidcwarden_roles']
        custom_claims:
          oidcwarden_roles: {}
    scopes:
      oidcwarden:
        claims: ['oidcwarden_roles']
    clients:
    - client_id: 'oidcwarden'
      client_name: 'Oidcwarden'
      client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
      public: false
      authorization_policy: 'two_factor'
      require_pkce: true
      pkce_challenge_method: 'S256'
      claims_policy: 'oidcwarden'
      redirect_uris:
        - 'https://vault.{{< sitevar name="domain" nojs="example.com" >}}/identity/connect/oidc-signin'
      scopes:
        - 'openid'
        - 'profile'
        - 'email'
        - 'offline_access'
        - 'oidcwarden'
      response_types:
        - 'code'
      grant_types:
        - 'authorization_code'
      access_token_signed_response_alg: 'none'
      userinfo_signed_response_alg: 'none'
      token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [Oidcwarden] there is one method, using the [Environment Variables](#environment-variables).

#### Environment Variables

To configure [Oidcwarden] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
SSO_ENABLED=true
SSO_ONLY=false
SSO_AUTHORITY=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
SSO_SCOPES=profile email offline_access oidcwarden
SSO_PKCE=true
SSO_CLIENT_ID=oidcwarden
SSO_CLIENT_SECRET=insecure_secret
SSO_ROLES_ENABLED=true
SSO_ROLES_DEFAULT_TO_USER=true
SSO_ROLES_TOKEN_PATH=/oidcwarden_roles
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  oidcwarden:
    environment:
      - SSO_ENABLED=true
      - SSO_ONLY=false
      - SSO_AUTHORITY=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
      - SSO_SCOPES=profile email offline_access oidcwarden
      - SSO_PKCE=true
      - SSO_CLIENT_ID=oidcwarden
      - SSO_CLIENT_SECRET=insecure_secret
      - SSO_ROLES_ENABLED=true
      - SSO_ROLES_DEFAULT_TO_USER=true
      - SSO_ROLES_TOKEN_PATH=/oidcwarden_roles
```


## See Also

- [SSO using OpenId Connect](https://github.com/Timshel/OIDCWarden/blob/main/SSO.md)

[Authelia]: https://www.authelia.com
[Oidcwarden]: https://github.com/Timshel/OIDCWarden/
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
