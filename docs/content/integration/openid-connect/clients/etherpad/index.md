---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Etherpad"
description: "A guide on integrating Etherpad with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Etherpad | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Etherpad with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Etherpad]
  - [v3.3.17](https://github.com/ether/etherpad-lite/releases/tag/v3.3.17)
- [ep_openid_connect]
  - [v3.0.54](https://github.com/ether/ep_openid_connect/tags)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://etherpad.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://etherpad.{{< sitevar name="domain" nojs="example.com" >}}/ep_openid_connect/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `etherpad`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Etherpad] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'etherpad'
        client_name: 'Etherpad'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://etherpad.{{< sitevar name="domain" nojs="example.com" >}}/ep_openid_connect/callback'
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

To configure [Etherpad] there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
This configuration requires the [ep_openid_connect] plugin to be installed in [Etherpad].
{{< /callout >}}

To configure [Etherpad] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```json {title="settings.json"}
{
  "requireAuthentication": true,
  "ep_openid_connect": {
    "issuer": "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}",
    "client_id": "etherpad",
    "client_secret": "insecure_secret",
    "base_url": "https://etherpad.{{< sitevar name="domain" nojs="example.com" >}}",
    "scope": ["openid", "profile", "email"],
    "token_endpoint_auth_method": "client_secret_post"
  }
}
```

## See Also

- [Etherpad ep_openid_connect Documentation](https://github.com/ether/ep_openid_connect#readme)

[Authelia]: https://www.authelia.com
[Etherpad]: https://etherpad.org/
[ep_openid_connect]: https://github.com/ether/ep_openid_connect
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
