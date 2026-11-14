---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Concourse"
description: "A guide on integrating Concourse with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Concourse | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Concourse with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Concourse]
  - [v8.3.3](https://github.com/concourse/concourse/releases/tag/v8.3.3)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://concourse.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://concourse.{{< sitevar name="domain" nojs="example.com" >}}/sky/issuer/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `concourse`
- **Client Secret:** `insecure_secret`
- **Administrator Group:** `concourse-admins`
  - Members of this group are added to the `main` team.

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Concourse] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'concourse'
        client_name: 'Concourse'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://concourse.{{< sitevar name="domain" nojs="example.com" >}}/sky/issuer/callback'
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

### Application

To configure [Concourse] there is one method, using [Environment Variables](#environment-variables).

#### Environment Variables

To configure [Concourse] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables on the `web` node:

##### Standard

```shell {title=".env"}
CONCOURSE_EXTERNAL_URL=https://concourse.{{< sitevar name="domain" nojs="example.com" >}}
CONCOURSE_OIDC_DISPLAY_NAME=Authelia
CONCOURSE_OIDC_ISSUER=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
CONCOURSE_OIDC_CLIENT_ID=concourse
CONCOURSE_OIDC_CLIENT_SECRET=insecure_secret
CONCOURSE_OIDC_SCOPE=profile,email,groups
CONCOURSE_OIDC_GROUPS_KEY=groups
CONCOURSE_OIDC_USER_NAME_KEY=preferred_username
CONCOURSE_MAIN_TEAM_OIDC_GROUP=concourse-admins
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  web:
    environment:
      CONCOURSE_EXTERNAL_URL: 'https://concourse.{{< sitevar name="domain" nojs="example.com" >}}'
      CONCOURSE_OIDC_DISPLAY_NAME: 'Authelia'
      CONCOURSE_OIDC_ISSUER: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      CONCOURSE_OIDC_CLIENT_ID: 'concourse'
      CONCOURSE_OIDC_CLIENT_SECRET: 'insecure_secret'
      CONCOURSE_OIDC_SCOPE: 'profile,email,groups'
      CONCOURSE_OIDC_GROUPS_KEY: 'groups'
      CONCOURSE_OIDC_USER_NAME_KEY: 'preferred_username'
      CONCOURSE_MAIN_TEAM_OIDC_GROUP: 'concourse-admins'
```

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
When the `CONCOURSE_OIDC_SCOPE` option is set it replaces the default `profile` and `email` scopes rather than
adding to them, which is why they are included alongside the `groups` scope.
{{< /callout >}}

## See Also

- [Concourse Generic OIDC Auth Documentation](https://concourse-ci.org/docs/auth-and-teams/configuring/generic-oidc/)

[Authelia]: https://www.authelia.com
[Concourse]: https://concourse-ci.org
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
