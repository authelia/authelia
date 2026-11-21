---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Gravity"
description: "A guide on integrating Gravity with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Gravity | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Gravity with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Gravity]
  - [v0.34.0](https://github.com/BeryJu/gravity/releases/tag/v0.34.0)

{{% oidc-common bugs="claims-hydration" %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://gravity.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://gravity.{{< sitevar name="domain" nojs="example.com" >}}/auth/oidc/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `gravity`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Gravity] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'gravity'
        client_name: 'Gravity'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://gravity.{{< sitevar name="domain" nojs="example.com" >}}/auth/oidc/callback'
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

#### Configuration Escape Hatch

{{% oidc-escape-hatch-claims-hydration client_id="gravity" claims="email" %}}

### Application

To configure [Gravity] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

{{< callout context="tip" title="Help Wanted" icon="outline/rocket" >}}
We would love screenshots of this configuration!
{{< /callout >}}

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
[Gravity] grants every user who successfully authenticates via [OpenID Connect 1.0] full access to the API. It is
strongly recommended that you restrict access to this client using an appropriate
[authorization policy](../../../../configuration/identity-providers/openid-connect/clients.md#authorization_policy).
{{< /callout >}}

To configure [Gravity] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to the [Gravity] administrative interface.
2. Navigate to Cluster.
3. Navigate to Roles.
4. Click `API`.
5. Configure the following options in the `OIDC` section:
   - Issuer: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
   - Client ID: `gravity`
   - Client Secret: `insecure_secret`
   - Redirect URL: `https://gravity.{{< sitevar name="domain" nojs="example.com" >}}/auth/oidc/callback`
   - Scopes: `openid`, `email`
6. Click `Update`.

When [OpenID Connect 1.0] is configured [Gravity] automatically starts the single sign-on flow. To use local
authentication instead, add the `local` query parameter to the URL, for example
`https://gravity.{{< sitevar name="domain" nojs="example.com" >}}/ui/?local`.

## See Also

- [Gravity API Role Configuration Documentation](https://gravity.beryju.io/docs/api/role_config/)

[Authelia]: https://www.authelia.com
[Gravity]: https://gravity.beryju.io/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
