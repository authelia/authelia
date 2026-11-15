---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Dependency-Track"
description: "A guide on integrating Dependency-Track with the Authelia OpenID Connect 1.0 Provider for SSO, with configuration examples and available options."
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
  title: "Dependency-Track | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Dependency-Track with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Dependency-Track]
  - [v5.2.1](https://github.com/DependencyTrack/dependency-track/releases/tag/5.2.1)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://dependency-track.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://dependency-track.{{< sitevar name="domain" nojs="example.com" >}}/static/oidc-callback.html`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `dependency-track`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Dependency-Track]
which will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'dependency-track'
        client_name: 'Dependency-Track'
        public: true
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://dependency-track.{{< sitevar name="domain" nojs="example.com" >}}/static/oidc-callback.html'
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
        token_endpoint_auth_method: 'none'
```

### Application

To configure [Dependency-Track] there is one method, using [Environment Variables](#environment-variables).

#### Environment Variables

To configure [Dependency-Track] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables on both the API Server and the Frontend containers:

##### Standard

API Server:

```shell {title=".env"}
DT_OIDC_ENABLED=true
DT_OIDC_ISSUER=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
DT_OIDC_CLIENT_ID=dependency-track
DT_OIDC_USERNAME_CLAIM=preferred_username
DT_OIDC_USER_PROVISIONING=true
DT_OIDC_TEAM_SYNCHRONIZATION=true
DT_OIDC_TEAMS_CLAIM=groups
```

Frontend:

```shell {title=".env"}
OIDC_ISSUER=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
OIDC_CLIENT_ID=dependency-track
OIDC_SCOPE=openid profile email groups
OIDC_LOGIN_BUTTON_TEXT=Sign in with Authelia
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  apiserver:
    environment:
      DT_OIDC_ENABLED: 'true'
      DT_OIDC_ISSUER: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      DT_OIDC_CLIENT_ID: 'dependency-track'
      DT_OIDC_USERNAME_CLAIM: 'preferred_username'
      DT_OIDC_USER_PROVISIONING: 'true'
      DT_OIDC_TEAM_SYNCHRONIZATION: 'true'
      DT_OIDC_TEAMS_CLAIM: 'groups'
  frontend:
    environment:
      OIDC_ISSUER: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      OIDC_CLIENT_ID: 'dependency-track'
      OIDC_SCOPE: 'openid profile email groups'
      OIDC_LOGIN_BUTTON_TEXT: 'Sign in with Authelia'
```

#### Team Synchronization

When `DT_OIDC_TEAM_SYNCHRONIZATION` is enabled, the groups from Authelia must be mapped to teams in
[Dependency-Track]. To configure the mapping, use the following instructions:

1. Login to [Dependency-Track] as an administrator.
2. Navigate to Administration.
3. Navigate to Access Management.
4. Navigate to OpenID Connect Groups.
5. Create a group with the exact name of the Authelia group (case-sensitive).
6. Add the teams the members of the group should belong to under Mapped Teams.

## See Also

- [Dependency-Track Configuring OpenID Connect Documentation](https://dependencytrack.github.io/docs/next/guides/administration/configuring-oidc/)

[Authelia]: https://www.authelia.com
[Dependency-Track]: https://dependencytrack.org/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
