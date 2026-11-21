---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "PostgreSQL"
description: "A guide on integrating PostgreSQL with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "PostgreSQL | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring PostgreSQL with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [PostgreSQL]
  - [v18.7](https://www.postgresql.org/docs/release/18.7/)
- [pg_oidc_validator]
  - [v1.1.0](https://github.com/percona/pg_oidc_validator/tree/1.1.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `postgresql.{{< sitevar name="domain" nojs="example.com" >}}`
  - This option is the hostname of the [PostgreSQL] server. The [OAuth 2.0 Device Authorization Grant] is used which
    does not require a redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `postgresql`
- **Client Secret:** This client is a public client and does not have a client secret.

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [PostgreSQL] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    claims_policies:
      postgresql:
        access_token: ['email']
    clients:
      - client_id: 'postgresql'
        client_name: 'PostgreSQL'
        public: true
        application_type: 'native'
        authorization_policy: 'two_factor'
        claims_policy: 'postgresql'
        require_pkce: false
        pkce_challenge_method: ''
        scopes:
          - 'openid'
          - 'email'
        grant_types:
          - 'urn:ietf:params:oauth:grant-type:device_code'
        access_token_signed_response_alg: 'RS256'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'none'
```

The `access_token_signed_response_alg` option must be set to `RS256` as [PostgreSQL] validates the access token as a
JWT, and the `postgresql` claims policy adds the `email` claim to the access token so it can be used to identify the
user.

### Application

To configure [PostgreSQL] there is one method, using the [Configuration File](#configuration-file).

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
[PostgreSQL] 18 or later is required on both the server and the client. The server must have TLS configured and the
[pg_oidc_validator] module installed, as [PostgreSQL] does not include an OAuth validator. The client must have the
optional `libpq-oauth` module installed, which is packaged separately as `libpq-oauth` on Debian and Ubuntu.
{{< /callout >}}

#### Configuration File

To configure [PostgreSQL] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Enable the validator by adding the following to the `postgresql.conf` file:

```ini {title="postgresql.conf"}
oauth_validator_libraries = 'pg_oidc_validator'
pg_oidc_validator.authn_field = 'email'
```

2. Add the following entry to the `pg_hba.conf` file above any other entry which matches the same connections:

```text {title="pg_hba.conf"}
# TYPE   DATABASE  USER  ADDRESS  METHOD  OPTIONS
hostssl  all       all   all      oauth   issuer="https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}" scope="openid email" map=authelia
```

3. Map the email addresses of Authelia users to [PostgreSQL] roles by adding the following to the `pg_ident.conf`
   file, which for example maps `john@{{< sitevar name="domain" nojs="example.com" >}}` to the `john` role:

```text {title="pg_ident.conf"}
# MAPNAME  SYSTEM-USERNAME          PG-USERNAME
authelia   /^(.*)@example\.com$     \1
```

The `SYSTEM-USERNAME` is a regular expression, so `example\.com` must be replaced with your email domain with each `.`
escaped as `\.`.

4. Create the mapped roles with the `LOGIN` attribute, for example `CREATE ROLE john LOGIN;`, and grant the required
   privileges.
5. Reload [PostgreSQL], for example by running `SELECT pg_reload_conf();` as an administrator.

To connect using `psql` run the following command, then visit the displayed URL and enter the displayed code:

```shell
psql "host=postgresql.{{< sitevar name="domain" nojs="example.com" >}} dbname=postgres user=john sslmode=require oauth_issuer=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}} oauth_client_id=postgresql"
```

## See Also

- [PostgreSQL OAuth Authorization/Authentication Documentation](https://www.postgresql.org/docs/18/auth-oauth.html)
- [PostgreSQL OAuth Support in libpq Documentation](https://www.postgresql.org/docs/18/libpq-oauth.html)
- [pg_oidc_validator Configuration Documentation](https://github.com/percona/pg_oidc_validator#configuration)

[Authelia]: https://www.authelia.com
[PostgreSQL]: https://www.postgresql.org/
[pg_oidc_validator]: https://github.com/percona/pg_oidc_validator
[OAuth 2.0 Device Authorization Grant]: https://datatracker.ietf.org/doc/html/rfc8628
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
