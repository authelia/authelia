---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "FOG Project"
description: "A guide on integrating FOG Project with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "FOG Project | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring FOG Project with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [FOG Project]
  - [v1.6.0.158](https://github.com/FOGProject/fogproject/releases/tag/1.6.0.158)

{{% oidc-common bugs="claims-hydration" %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://fog.{{< sitevar name="domain" nojs="example.com" >}}/fog/`
  - This option determines the redirect URI in the format of
    `https://fog.{{< sitevar name="domain" nojs="example.com" >}}/fog/ext/oidc/callback`.
    This means if you change this value, you need to update the redirect URI. The value is derived from the
    `FOG_WEB_HOST` and `FOG_WEB_ROOT` settings, and the exact value is shown on the provider page in [FOG Project].
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `fog`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [FOG Project] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'fog'
        client_name: 'FOG Project'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://fog.{{< sitevar name="domain" nojs="example.com" >}}/fog/ext/oidc/callback'
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

{{% oidc-escape-hatch-claims-hydration client_id="fog" claims="preferred_username,name,groups" %}}

### Application

To configure [FOG Project] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [FOG Project] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to the [FOG Project] web management interface as an administrator.
2. Ensure the `OIDC` plugin is installed and activated.
3. Navigate to `OpenID Connect` and click `Create New OpenID Connect Provider`.
4. Configure the following options:
   - Name: `Authelia`
   - Issuer: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
   - Client ID: `fog`
   - Client Secret: `insecure_secret`
   - Scopes: `openid profile email groups`
   - Username Claim: `preferred_username`
   - Group Claim: `groups`
5. Save the provider.
6. Open the provider and configure the following options:
   - Enabled: Enabled
   - Create Users On First Login: Enable if you want users to automatically be created in [FOG Project].
7. Save the provider.
8. Ensure the redirect URI shown on the provider page is
   `https://fog.{{< sitevar name="domain" nojs="example.com" >}}/fog/ext/oidc/callback`.
9. Optionally navigate to `OpenID Connect Groups` and click `Create New Provider Group` to map Authelia groups to
   [FOG Project] roles and user groups using the `Role Association` and `User Group Association` tabs.

The Issuer must not include a trailing slash, as [FOG Project] requires the `iss` claim of the ID Token to exactly match
the configured value.

## See Also

- [FOG Project OpenID Connect Documentation](https://docs.fogproject.org/en/latest/management/web/oidc/)

[Authelia]: https://www.authelia.com
[FOG Project]: https://fogproject.org/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
