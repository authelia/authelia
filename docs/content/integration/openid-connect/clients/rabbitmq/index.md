---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "RabbitMQ"
description: "A guide on integrating RabbitMQ with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "RabbitMQ | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring RabbitMQ with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [RabbitMQ]
  - [v4.4.0](https://github.com/rabbitmq/rabbitmq-server/releases/tag/v4.4.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://rabbitmq.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://rabbitmq.{{< sitevar name="domain" nojs="example.com" >}}/js/oidc-oauth/login-callback.html`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `rabbitmq`
- **Resource Server ID:** `rabbitmq`
- **Groups:**
  - `rabbitmq-administrator`: Users in this group are granted full administrator access.
  - `rabbitmq-monitoring`: Users in this group are granted read-only monitoring access.

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [RabbitMQ] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    cors:
      allowed_origins_from_client_redirect_uris: true
      endpoints:
        - 'token'
        - 'revocation'
    claims_policies:
      rabbitmq:
        access_token: ['groups']
    clients:
      - client_id: 'rabbitmq'
        client_name: 'RabbitMQ'
        public: true
        application_type: 'web'
        authorization_policy: 'two_factor'
        claims_policy: 'rabbitmq'
        require_pkce: true
        pkce_challenge_method: 'S256'
        audience:
          - 'rabbitmq'
        requested_audience_mode: 'implicit'
        redirect_uris:
          - 'https://rabbitmq.{{< sitevar name="domain" nojs="example.com" >}}/js/oidc-oauth/login-callback.html'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
          - 'groups'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'RS256'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'none'
```

Notes:

- The [RabbitMQ] management UI is a browser based public client which calls the token endpoint directly, so the `token`
  and `revocation` endpoints must permit cross-origin requests.
- [RabbitMQ] validates JWT access tokens, so the `access_token_signed_response_alg` option must be configured with a
  signing algorithm. The `audience` must contain the `auth_oauth2.resource_server_id` value, and the
  [requested_audience_mode] value of `implicit` grants this audience without the client explicitly requesting it.
- The `groups` claim is included in the access token by the claims policy so [RabbitMQ] can map the groups to its
  permissions.

### Application

To configure [RabbitMQ] there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

To configure [RabbitMQ] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Enable the OAuth 2.0 authentication backend plugin:

   ```shell
   rabbitmq-plugins enable rabbitmq_auth_backend_oauth2
   ```

2. Add the following configuration:

   ```ini {title="rabbitmq.conf"}
   auth_backends.1 = rabbit_auth_backend_oauth2
   auth_backends.2 = rabbit_auth_backend_internal

   auth_oauth2.resource_server_id = rabbitmq
   auth_oauth2.issuer = https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
   auth_oauth2.additional_scopes_key = groups
   auth_oauth2.scope_aliases.1.alias = rabbitmq-administrator
   auth_oauth2.scope_aliases.1.scope = rabbitmq.tag:administrator rabbitmq.read:*/* rabbitmq.write:*/* rabbitmq.configure:*/*
   auth_oauth2.scope_aliases.2.alias = rabbitmq-monitoring
   auth_oauth2.scope_aliases.2.scope = rabbitmq.tag:monitoring rabbitmq.read:*/*

   management.oauth_enabled = true
   management.oauth_client_id = rabbitmq
   management.oauth_scopes = openid profile email groups
   ```

3. Restart [RabbitMQ] for the changes to take effect.

## See Also

- [RabbitMQ OAuth 2.0 Authentication Backend Documentation](https://www.rabbitmq.com/docs/oauth2)
- [RabbitMQ Management Plugin Documentation](https://www.rabbitmq.com/docs/management)

[Authelia]: https://www.authelia.com
[RabbitMQ]: https://www.rabbitmq.com/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
[requested_audience_mode]: ../../../../configuration/identity-providers/openid-connect/clients.md#requested_audience_mode
