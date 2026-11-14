---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Velociraptor"
description: "A guide on integrating Velociraptor with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Velociraptor | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Velociraptor with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Velociraptor]
  - [v0.77.3](https://github.com/Velocidex/velociraptor/releases/tag/v0.77.3)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://velociraptor.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://velociraptor.{{< sitevar name="domain" nojs="example.com" >}}/auth/oidc/Authelia/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `velociraptor`
- **Client Secret:** `insecure_secret`
- **OIDC Name:** `Authelia`
  - This value is part of the redirect URI. If you change this value, you need to update the redirect URI.

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Velociraptor] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'velociraptor'
        client_name: 'Velociraptor'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://velociraptor.{{< sitevar name="domain" nojs="example.com" >}}/auth/oidc/Authelia/callback'
        scopes:
          - 'openid'
          - 'email'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [Velociraptor] there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

To configure [Velociraptor] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```yaml {title="server.config.yaml"}
GUI:
  public_url: 'https://velociraptor.{{< sitevar name="domain" nojs="example.com" >}}/'
  authenticator:
    type: 'oidc'
    oidc_issuer: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
    oidc_name: 'Authelia'
    oauth_client_id: 'velociraptor'
    oauth_client_secret: 'insecure_secret'
```

## See Also

- [Velociraptor GUI Authenticator Configuration Reference](https://docs.velociraptor.app/docs/deployment/references/#GUI.authenticator)
- [Velociraptor Security Configuration Documentation](https://docs.velociraptor.app/docs/deployment/security/#authentication)

[Authelia]: https://www.authelia.com
[Velociraptor]: https://docs.velociraptor.app/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
