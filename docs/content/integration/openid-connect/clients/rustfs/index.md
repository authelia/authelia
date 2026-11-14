---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "RustFS"
description: "A guide on integrating RustFS with the Authelia OpenID Connect 1.0 Provider including example configurations and an overview of available options for SSO."
summary: ""
date: 2026-11-11T17:00:00+01:00
draft: false
images: []
weight: 620
toc: true
support:
  level: community
  versions: true
  integration: true
seo:
  title: "RustFS | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring RustFS with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [RustFS]
  - [1.0.4](https://github.com/rustfs/rustfs/releases/tag/1.0.4)

{{% oidc-common bugs="claims-hydration" %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://rustfs.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `rustfs`
- **Client Secret:** `insecure_secret`
- **Authorized Group:** `rustfs`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [RustFS] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    authorization_policies:
      rustfs:
        default_policy: 'deny'
        rules:
          - policy: 'two_factor'
            subject: 'group:rustfs'
    clients:
      - client_id: 'rustfs'
        client_name: 'RustFS'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'rustfs'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://rustfs.{{< sitevar name="domain" nojs="example.com" >}}/rustfs/admin/v3/oidc/callback/default'
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

The `token_endpoint_auth_method` must be `client_secret_post`. [RustFS] always sends the client credentials in the
request body and has no option to use `client_secret_basic`.

The `rustfs` [authorization policy] only decides who may log in to [RustFS] at all. What a user may do once logged in
is decided by [RustFS] from the `groups` claim, see [Group to Policy Mapping](#group-to-policy-mapping).

#### Configuration Escape Hatch

[RustFS] does not request the UserInfo Endpoint and reads the `groups`, `email`, and `preferred_username` claims solely
from the ID Token. Without the following adaptation the ID Token only contains the standard claims and the login fails
with `OIDC policy mapping did not resolve to current policies`.

{{% oidc-escape-hatch-claims-hydration client_id="rustfs" claims="groups,email,preferred_username" %}}

### Application

To configure [RustFS] there is one method, using [Environment Variables](#environment-variables).

#### Environment Variables

To configure [RustFS] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment variables:

##### Standard

```shell {title=".env"}
RUSTFS_BROWSER_REDIRECT_URL=https://rustfs.{{< sitevar name="domain" nojs="example.com" >}}
RUSTFS_IDENTITY_OPENID_ENABLE=on
RUSTFS_IDENTITY_OPENID_CONFIG_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
RUSTFS_IDENTITY_OPENID_CLIENT_ID=rustfs
RUSTFS_IDENTITY_OPENID_CLIENT_SECRET=insecure_secret
RUSTFS_IDENTITY_OPENID_SCOPES=openid,profile,email,groups
RUSTFS_IDENTITY_OPENID_REDIRECT_URI=https://rustfs.{{< sitevar name="domain" nojs="example.com" >}}/rustfs/admin/v3/oidc/callback/default
RUSTFS_IDENTITY_OPENID_REDIRECT_URI_DYNAMIC=off
RUSTFS_IDENTITY_OPENID_DISPLAY_NAME=Authelia
RUSTFS_IDENTITY_OPENID_GROUPS_CLAIM=groups
RUSTFS_IDENTITY_OPENID_EMAIL_CLAIM=email
RUSTFS_IDENTITY_OPENID_USERNAME_CLAIM=preferred_username
RUSTFS_OUTBOUND_ALLOW_ORIGINS=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  rustfs:
    environment:
      RUSTFS_BROWSER_REDIRECT_URL: 'https://rustfs.{{< sitevar name="domain" nojs="example.com" >}}'
      RUSTFS_IDENTITY_OPENID_ENABLE: 'on'
      RUSTFS_IDENTITY_OPENID_CONFIG_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      RUSTFS_IDENTITY_OPENID_CLIENT_ID: 'rustfs'
      RUSTFS_IDENTITY_OPENID_CLIENT_SECRET: 'insecure_secret'
      RUSTFS_IDENTITY_OPENID_SCOPES: 'openid,profile,email,groups'
      RUSTFS_IDENTITY_OPENID_REDIRECT_URI: 'https://rustfs.{{< sitevar name="domain" nojs="example.com" >}}/rustfs/admin/v3/oidc/callback/default'
      RUSTFS_IDENTITY_OPENID_REDIRECT_URI_DYNAMIC: 'off'
      RUSTFS_IDENTITY_OPENID_DISPLAY_NAME: 'Authelia'
      RUSTFS_IDENTITY_OPENID_GROUPS_CLAIM: 'groups'
      RUSTFS_IDENTITY_OPENID_EMAIL_CLAIM: 'email'
      RUSTFS_IDENTITY_OPENID_USERNAME_CLAIM: 'preferred_username'
      RUSTFS_OUTBOUND_ALLOW_ORIGINS: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
```

The `RUSTFS_OUTBOUND_ALLOW_ORIGINS` variable is only required when the Authelia Root URL resolves to a private address
as seen from [RustFS], which is common for self-hosted deployments. [RustFS] refuses outbound requests to private
addresses unless their origin is allowed, and logs `OIDC provider discovery blocked by outbound policy` at startup
otherwise.

#### Reverse Proxy

[RustFS] serves the S3 API on port `9000` and the console on port `9001` by default. The console port also serves the
OIDC authorization and callback endpoints, so the reverse proxy for the Application Root URL only needs to forward to
the console port.

#### Group to Policy Mapping

[RustFS] looks up every value of the `groups` claim as the name of a [RustFS] policy. Groups without a matching
policy are skipped. If no group matches a policy the login fails with
`OIDC policy mapping did not resolve to current policies`.

For example, adding a user to both the `rustfs` group (to pass the authorization policy) and the `consoleAdmin` group
(a built-in [RustFS] policy) grants them full administrative access. Custom policies can be created in the [RustFS]
console while logged in with the root credentials, and are then granted by adding users to an Authelia group with the
same name.

The built-in policies are `consoleAdmin`, `readonly`, `readwrite`, `writeonly`, `diagnostics`, `KMSAuditor`,
`KMSKeyAdministrator`, and `KMSKeyUser`. Avoid using these names for groups that are used by other applications, as
any member of the `rustfs` group who is also in such a group receives the matching [RustFS] policy.

## See Also

- [RustFS External Identity (OIDC) Documentation](https://docs.rustfs.com/administration/iam/oidc)

[RustFS]: https://rustfs.com/
[Authelia]: https://www.authelia.com
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
[authorization policy]: ../../../../configuration/identity-providers/openid-connect/provider.md#authorization_policies
