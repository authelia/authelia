---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "AKHQ"
description: "A guide on integrating AKHQ with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "AKHQ | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring AKHQ with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [AKHQ]
  - [0.28.0](https://github.com/tchiotludo/akhq/releases/tag/0.28.0)

{{% oidc-common bugs="client-credentials-encoding,claims-hydration" %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://akhq.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://akhq.{{< sitevar name="domain" nojs="example.com" >}}/oauth/callback/authelia`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `akhq`
- **Client Secret:** `insecure_secret`
- **Admin Group Name:** `akhq-admins`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [AKHQ] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'akhq'
        client_name: 'AKHQ'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://akhq.{{< sitevar name="domain" nojs="example.com" >}}/oauth/callback/authelia'
        scopes:
          - 'openid'
          - 'email'
          - 'profile'
          - 'groups'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

#### Configuration Escape Hatch

{{% oidc-escape-hatch-claims-hydration client_id="akhq" claims="preferred_username,groups" %}}

### Application

To configure [AKHQ] there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

To configure [AKHQ] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```yaml {title="application.yml"}
micronaut:
  security:
    enabled: true
    oauth2:
      enabled: true
      clients:
        authelia:
          client-id: 'akhq'
          client-secret: 'insecure_secret'
          openid:
            issuer: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
akhq:
  security:
    default-group: 'no-roles'
    oidc:
      enabled: true
      providers:
        authelia:
          label: 'Login with Authelia'
          username-field: 'preferred_username'
          groups-field: 'groups'
          default-group: 'reader'
          groups:
            - name: 'akhq-admins'
              groups:
                - 'admin'
```

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
The name of the client under `micronaut.security.oauth2.clients` and the name of the provider under
`akhq.security.oidc.providers` must match, and this name determines the last segment of the redirect URI.
{{< /callout >}}

## See Also

- [AKHQ OIDC Authentication Documentation](https://akhq.io/docs/configuration/authentifications/oidc.html)

[Authelia]: https://www.authelia.com
[AKHQ]: https://akhq.io/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
