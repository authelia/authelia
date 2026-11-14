---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "BuildBuddy"
description: "A guide on integrating BuildBuddy with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "BuildBuddy | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring BuildBuddy with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [BuildBuddy]
  - [v2.324.0](https://github.com/buildbuddy-io/buildbuddy/releases/tag/v2.324.0)

{{% oidc-common bugs="claims-hydration" %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://buildbuddy.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://buildbuddy.{{< sitevar name="domain" nojs="example.com" >}}/auth/`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `buildbuddy`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [BuildBuddy] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'buildbuddy'
        client_name: 'BuildBuddy'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://buildbuddy.{{< sitevar name="domain" nojs="example.com" >}}/auth/'
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

#### Configuration Escape Hatch

{{% oidc-escape-hatch-claims-hydration client_id="buildbuddy" claims="email,name,given_name,family_name" %}}

### Application

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
Authentication via [OpenID Connect 1.0] is only available in the self-hosted Enterprise version of [BuildBuddy].
{{< /callout >}}

To configure [BuildBuddy] there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

To configure [BuildBuddy] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration,
replacing the `jwt_key` value with a securely generated random value which is the same on every [BuildBuddy] server
replica:

```yaml {title="config.yaml"}
app:
  build_buddy_url: 'https://buildbuddy.{{< sitevar name="domain" nojs="example.com" >}}'
auth:
  jwt_key: 'a_securely_generated_random_value'
  oauth_providers:
    - issuer_url: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      client_id: 'buildbuddy'
      client_secret: 'insecure_secret'
```

## See Also

- [BuildBuddy Auth Configuration Documentation](https://www.buildbuddy.io/docs/config-auth/)
- [BuildBuddy Enterprise Configuration Documentation](https://www.buildbuddy.io/docs/enterprise-config/)

[Authelia]: https://www.authelia.com
[BuildBuddy]: https://www.buildbuddy.io/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
