---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Yamtrack"
description: "A guide on integrating Yamtrack with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Yamtrack | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Yamtrack with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Yamtrack]
  - [v0.26.3](https://github.com/FuzzyGrim/Yamtrack/releases/tag/v0.26.3)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://yamtrack.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://yamtrack.{{< sitevar name="domain" nojs="example.com" >}}/accounts/oidc/authelia/login/callback/`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `yamtrack`
- **Client Secret:** `insecure_secret`
- **Provider ID:** `authelia`
  - This option determines the redirect URI in the format of
    `https://yamtrack.{{< sitevar name="domain" nojs="example.com" >}}/accounts/oidc/<provider_id>/login/callback/`.
    This means if you change this value, you need to update the redirect URI.

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Yamtrack] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'yamtrack'
        client_name: 'Yamtrack'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://yamtrack.{{< sitevar name="domain" nojs="example.com" >}}/accounts/oidc/authelia/login/callback/'
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

To configure [Yamtrack] there is one method, using [Environment Variables](#environment-variables).

#### Environment Variables

To configure [Yamtrack] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
SOCIAL_PROVIDERS=allauth.socialaccount.providers.openid_connect
SOCIALACCOUNT_PROVIDERS={"openid_connect":{"OAUTH_PKCE_ENABLED":true,"APPS":[{"provider_id":"authelia","name":"Authelia","client_id":"yamtrack","secret":"insecure_secret","settings":{"server_url":"https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration"}}]}}
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  yamtrack:
    environment:
      SOCIAL_PROVIDERS: 'allauth.socialaccount.providers.openid_connect'
      SOCIALACCOUNT_PROVIDERS: '{"openid_connect":{"OAUTH_PKCE_ENABLED":true,"APPS":[{"provider_id":"authelia","name":"Authelia","client_id":"yamtrack","secret":"insecure_secret","settings":{"server_url":"https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration"}}]}}'
```

Optionally the following environment variables can also be configured:

- `SOCIALACCOUNT_ONLY=True` disables the local username and password login.
- `REDIRECT_LOGIN_TO_SSO=True` automatically redirects users from the login page to Authelia.

## See Also

- [Yamtrack Social Authentication Documentation](https://github.com/FuzzyGrim/Yamtrack/blob/dev/docs/social-auth.md)

[Authelia]: https://www.authelia.com
[Yamtrack]: https://github.com/FuzzyGrim/Yamtrack
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
