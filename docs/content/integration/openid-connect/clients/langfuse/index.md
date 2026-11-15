---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Langfuse"
description: "A guide on integrating Langfuse with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Langfuse | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Langfuse with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Langfuse]
  - [v4.79.0](https://github.com/langfuse/langfuse/releases/tag/v4.79.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://langfuse.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://langfuse.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/callback/custom`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `langfuse`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Langfuse] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'langfuse'
        client_name: 'Langfuse'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://langfuse.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/callback/custom'
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

To configure [Langfuse] there is one method, using [Environment Variables](#environment-variables).

#### Environment Variables

To configure [Langfuse] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
NEXTAUTH_URL=https://langfuse.{{< sitevar name="domain" nojs="example.com" >}}
AUTH_CUSTOM_NAME=Authelia
AUTH_CUSTOM_ISSUER=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
AUTH_CUSTOM_CLIENT_ID=langfuse
AUTH_CUSTOM_CLIENT_SECRET=insecure_secret
AUTH_CUSTOM_SCOPE=openid email profile
AUTH_CUSTOM_CLIENT_AUTH_METHOD=client_secret_basic
AUTH_CUSTOM_CHECKS=pkce,state
AUTH_CUSTOM_FETCH_USERINFO=true
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  langfuse-web:
    environment:
      NEXTAUTH_URL: 'https://langfuse.{{< sitevar name="domain" nojs="example.com" >}}'
      AUTH_CUSTOM_NAME: 'Authelia'
      AUTH_CUSTOM_ISSUER: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      AUTH_CUSTOM_CLIENT_ID: 'langfuse'
      AUTH_CUSTOM_CLIENT_SECRET: 'insecure_secret'
      AUTH_CUSTOM_SCOPE: 'openid email profile'
      AUTH_CUSTOM_CLIENT_AUTH_METHOD: 'client_secret_basic'
      AUTH_CUSTOM_CHECKS: 'pkce,state'
      AUTH_CUSTOM_FETCH_USERINFO: 'true'
```

Optionally `AUTH_DISABLE_USERNAME_PASSWORD=true` can be set to disable the local email and password login.

## See Also

- [Langfuse Authentication and SSO Documentation](https://langfuse.com/self-hosting/security/authentication-and-sso)

[Authelia]: https://www.authelia.com
[Langfuse]: https://langfuse.com
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
