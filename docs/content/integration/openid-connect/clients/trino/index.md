---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Trino"
description: "A guide on integrating Trino with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Trino | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Trino with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Trino]
  - [484](https://github.com/trinodb/trino/releases/tag/484)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://trino.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://trino.{{< sitevar name="domain" nojs="example.com" >}}/oauth2/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `trino`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Trino] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'trino'
        client_name: 'Trino'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://trino.{{< sitevar name="domain" nojs="example.com" >}}/oauth2/callback'
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

To configure [Trino] there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
[Trino] requires TLS to be enabled on the coordinator (either directly or via a TLS terminating reverse proxy with
`http-server.process-forwarded=true`) and requires a shared secret for internal communication when OAuth 2.0
authentication is enabled. Only the coordinator needs the OAuth 2.0 configuration.
{{< /callout >}}

To configure [Trino] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration on the
coordinator:

```properties {title="etc/config.properties"}
http-server.process-forwarded=true
internal-communication.shared-secret=a_long_random_shared_secret
http-server.authentication.type=oauth2
web-ui.authentication.type=oauth2
http-server.authentication.oauth2.issuer=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
http-server.authentication.oauth2.client-id=trino
http-server.authentication.oauth2.client-secret=insecure_secret
http-server.authentication.oauth2.scopes=openid,email,profile
http-server.authentication.oauth2.principal-field=preferred_username
http-server.authentication.oauth2.oidc.discovery=true
http-server.authentication.oauth2.oidc.use-userinfo-endpoint=true
```

## See Also

- [Trino OAuth 2.0 Authentication Documentation](https://trino.io/docs/current/security/oauth2.html)

[Authelia]: https://www.authelia.com
[Trino]: https://trino.io
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
