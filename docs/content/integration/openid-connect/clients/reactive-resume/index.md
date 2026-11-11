---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Reactive Resume"
description: "A guide on integrating Reactive Resume with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Reactive Resume | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Reactive Resume with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Reactive Resume]
  - [v6.1.0](https://github.com/reactive-resume/reactive-resume/releases/tag/v6.1.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://resume.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://resume.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/callback/custom`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `reactive-resume`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Reactive Resume]
which will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'reactive-resume'
        client_name: 'Reactive Resume'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://resume.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/callback/custom'
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
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

To configure [Reactive Resume] there is one method, using [Environment Variables](#environment-variables).

#### Environment Variables

To configure [Reactive Resume] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
APP_URL=https://resume.{{< sitevar name="domain" nojs="example.com" >}}
OAUTH_PROVIDER_NAME=Authelia
OAUTH_CLIENT_ID=reactive-resume
OAUTH_CLIENT_SECRET=insecure_secret
OAUTH_DISCOVERY_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration
OAUTH_SCOPES=openid profile email
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  reactive-resume:
    environment:
      APP_URL: 'https://resume.{{< sitevar name="domain" nojs="example.com" >}}'
      OAUTH_PROVIDER_NAME: 'Authelia'
      OAUTH_CLIENT_ID: 'reactive-resume'
      OAUTH_CLIENT_SECRET: 'insecure_secret'
      OAUTH_DISCOVERY_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration'
      OAUTH_SCOPES: 'openid profile email'
```

## See Also

- [Reactive Resume Single Sign-On Documentation](https://docs.rxresu.me/self-hosting/sso)

[Authelia]: https://www.authelia.com
[Reactive Resume]: https://rxresu.me/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
