---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Uptrace"
description: "A guide on integrating Uptrace with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Uptrace | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Uptrace with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Uptrace]
  - [v2.1.0-rc.2](https://github.com/uptrace/uptrace/releases/tag/v2.1.0-rc.2)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://uptrace.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://uptrace.{{< sitevar name="domain" nojs="example.com" >}}/api/v1/sso/authelia/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `uptrace`
- **Client Secret:** `insecure_secret`
- **SSO Domain:** `authelia`
  - This option determines the redirect URI in the format of
    `https://uptrace.{{< sitevar name="domain" nojs="example.com" >}}/api/v1/sso/authelia/callback`.
    This means if you change this value, you need to update the redirect URI.

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Uptrace] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'uptrace'
        client_name: 'Uptrace'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://uptrace.{{< sitevar name="domain" nojs="example.com" >}}/api/v1/sso/authelia/callback'
        scopes:
          - 'openid'
          - 'profile'
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

To configure [Uptrace] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
Single Sign-On in [Uptrace] requires a commercial license. The redirect URI displayed by [Uptrace] after creating the
SSO provider is the authoritative value and must be used in the Authelia `redirect_uris` option if it differs from the
example above.
{{< /callout >}}

To configure [Uptrace] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [Uptrace].
2. Navigate to `Organization` and then `Single Sign-On`.
3. Click `New SSO` and then `New OIDC`.
4. Configure the following options:
   - Domain: `authelia`
   - User team: Select the team new users should be assigned to.
   - User role: Select the role new users should be assigned.
5. Click `Create`.
6. Copy the displayed redirect URL and ensure it matches the `redirect_uris` option in the Authelia configuration.
7. Configure the following options:
   - Client ID: `uptrace`
   - Client Secret: `insecure_secret`
   - Issuer URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
8. Click `Save`.

Users can now login at `https://uptrace.{{< sitevar name="domain" nojs="example.com" >}}/auth/sso/authelia`.

## See Also

- [Uptrace Keycloak Single Sign-On Documentation](https://uptrace.dev/features/sso/keycloak)

[Authelia]: https://www.authelia.com
[Uptrace]: https://uptrace.dev
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
