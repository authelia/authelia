---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Unraid"
description: "A guide on integrating Unraid with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Unraid | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Unraid with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Unraid]
  - [v7.4.0](https://docs.unraid.net/unraid-os/release-notes/7.4.0/)

{{% oidc-common bugs="claims-hydration" %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://unraid.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://unraid.{{< sitevar name="domain" nojs="example.com" >}}/graphql/api/auth/oidc/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `unraid`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

### Special Notes

1. [OpenID Connect 1.0] support was introduced in [Unraid] 7.2.0.

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Unraid] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'unraid'
        client_name: 'Unraid'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://unraid.{{< sitevar name="domain" nojs="example.com" >}}/graphql/api/auth/oidc/callback'
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
        token_endpoint_auth_method: 'client_secret_post'
```

#### Configuration Escape Hatch

{{% oidc-escape-hatch-claims-hydration client_id="unraid" claims="email,name,groups" %}}

### Application

To configure [Unraid] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [Unraid] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [Unraid].
2. Navigate to `Settings` > `Management Access` > `API` > `OIDC`.
3. Click the plus button in the `OIDC Providers` section.
4. Configure the following options:
   - Provider ID: `authelia`
   - Provider Name: `Authelia`
   - Client ID: `unraid`
   - Client Secret: `insecure_secret`
   - Issuer URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
   - Scopes: `openid`, `profile`, `email`, `groups`
   - Authorization Mode: `simple`
   - Allowed Email Domains: `{{< sitevar name="domain" nojs="example.com" >}}`
5. Click `Apply` at the bottom of the page.

{{< callout context="tip" title="Did you know?" icon="outline/rocket" >}}
The `advanced` Authorization Mode can be used to restrict access based on the `groups` claim, for example by adding an
Authorization Rule with the Claim `groups`, the Operator `equals`, and the Value `unraid-admins`.
{{< /callout >}}

## See Also

- [Unraid OIDC Provider Setup Documentation](https://docs.unraid.net/API/oidc-provider-setup/)

[Authelia]: https://www.authelia.com
[Unraid]: https://unraid.net/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
