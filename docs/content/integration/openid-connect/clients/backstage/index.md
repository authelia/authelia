---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Backstage"
description: "A guide on integrating Backstage with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Backstage | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Backstage with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Backstage]
  - [v1.56.4](https://github.com/backstage/backstage/releases/tag/v1.56.4)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://backstage.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://backstage.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/oidc/handler/frame`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `backstage`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Backstage] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'backstage'
        client_name: 'Backstage'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://backstage.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/oidc/handler/frame'
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

To configure [Backstage] there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
The [Backstage] OpenID Connect 1.0 provider is not enabled by default. The
`@backstage/plugin-auth-backend-module-oidc-provider` module must be added to the backend, and a sign-in page using
the `oidc` provider must be added to the frontend. See the
[Backstage OIDC provider documentation](https://backstage.io/docs/auth/oidc) for the code changes required.
{{< /callout >}}

To configure [Backstage] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```yaml {title="app-config.yaml"}
app:
  baseUrl: 'https://backstage.{{< sitevar name="domain" nojs="example.com" >}}'
backend:
  baseUrl: 'https://backstage.{{< sitevar name="domain" nojs="example.com" >}}'
auth:
  environment: 'production'
  session:
    secret: '${AUTH_SESSION_SECRET}'
  providers:
    oidc:
      production:
        metadataUrl: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration'
        clientId: 'backstage'
        clientSecret: 'insecure_secret'
        tokenEndpointAuthMethod: 'client_secret_basic'
        tokenSignedResponseAlg: 'RS256'
        additionalScopes:
          - 'offline_access'
        prompt: 'auto'
        signIn:
          resolvers:
            - resolver: 'emailMatchingUserEntityProfileEmail'
```

## See Also

- [Backstage OIDC Provider Documentation](https://backstage.io/docs/auth/oidc)
- [Backstage Sign-in Resolvers Documentation](https://backstage.io/docs/auth/identity-resolver)

[Authelia]: https://www.authelia.com
[Backstage]: https://backstage.io/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
