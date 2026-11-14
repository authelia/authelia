---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "zot"
description: "A guide on integrating zot with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "zot | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring zot with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [zot]
  - [v2.1.23](https://github.com/project-zot/zot/releases/tag/v2.1.23)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://zot.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://zot.{{< sitevar name="domain" nojs="example.com" >}}/zot/auth/callback/oidc`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `zot`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

### Special Notes

1. The [zot] [OpenID Connect 1.0] integration is for the web interface. Command-line clients should use an API key
   generated in [zot] after logging in via [OpenID Connect 1.0].

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [zot] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'zot'
        client_name: 'zot'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://zot.{{< sitevar name="domain" nojs="example.com" >}}/zot/auth/callback/oidc'
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
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [zot] there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

To configure [zot] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```json {title="/etc/zot/config.json"}
{
  "http": {
    "externalUrl": "https://zot.{{< sitevar name="domain" nojs="example.com" >}}",
    "auth": {
      "openid": {
        "providers": {
          "oidc": {
            "name": "Authelia",
            "credentialsFile": "/etc/zot/oidc-credentials.json",
            "issuer": "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}",
            "keypath": "",
            "scopes": ["openid", "profile", "email", "groups"]
          }
        }
      }
    }
  }
}
```

```json {title="/etc/zot/oidc-credentials.json"}
{
  "clientid": "zot",
  "clientsecret": "insecure_secret"
}
```

## See Also

- [zot User Authentication and Authorization Documentation](https://zotregistry.dev/v2.1.23/articles/authn-authz/#social-login-using-openidoauth2)

[Authelia]: https://www.authelia.com
[zot]: https://zotregistry.dev/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
