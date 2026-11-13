---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Docspell"
description: "A guide on integrating Docspell with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Docspell | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Docspell with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Docspell]
  - [v0.43.0](https://github.com/eikek/docspell/releases/tag/v0.43.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://docspell.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://docspell.{{< sitevar name="domain" nojs="example.com" >}}/api/v1/open/auth/openid/authelia/resume`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `docspell`
- **Client Secret:** `insecure_secret`
- **Provider ID:** `authelia`
- **Collective:** All users who log in via Authelia are placed in the `family` collective.

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Docspell] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'docspell'
        client_name: 'Docspell'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://docspell.{{< sitevar name="domain" nojs="example.com" >}}/api/v1/open/auth/openid/authelia/resume'
        scopes:
          - 'openid'
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

To configure [Docspell] there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

To configure [Docspell] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration in the
REST server configuration file. The `openid` option is a list and therefore can't be configured using environment
variables. The user information is retrieved from the UserInfo Endpoint, which requires the `sign-key` option to be an
empty string.

```hocon {title="docspell-server.conf"}
docspell.server {
  base-url = "https://docspell.{{< sitevar name="domain" nojs="example.com" >}}"

  openid = [
    {
      enabled = true
      display = "Authelia"
      provider = {
        provider-id = "authelia"
        client-id = "docspell"
        client-secret = "insecure_secret"
        scope = "openid profile"
        authorize-url = "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization"
        token-url = "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token"
        user-url = "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo"
        sign-key = ""
        sig-algo = "RS256"
      }
      collective-key = "fixed:family"
      user-key = "preferred_username"
    }
  ]
}
```

## See Also

- [Docspell Authentication Documentation](https://docspell.org/docs/configure/authentication/)

[Authelia]: https://www.authelia.com
[Docspell]: https://docspell.org/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
