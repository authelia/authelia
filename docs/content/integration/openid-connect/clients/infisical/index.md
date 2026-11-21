---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Infisical"
description: "A guide on integrating Infisical with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Infisical | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Infisical with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Infisical]
  - [v0.167.8](https://github.com/Infisical/infisical/releases/tag/v0.167.8)

{{% oidc-common bugs="claims-hydration" %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://infisical.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://infisical.{{< sitevar name="domain" nojs="example.com" >}}/api/v1/sso/oidc/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `infisical`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

### Requirements

[Infisical] has the following requirements for [OpenID Connect 1.0] single sign-on:

- An Infisical Cloud Pro plan or an enterprise license for a self-hosted installation.
- The email domain of the organization must be verified in [Infisical].
- A self-hosted installation must have the `SITE_URL` and `AUTH_SECRET` environment variables configured, where
  `SITE_URL` is the Application Root URL without the trailing slash.

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Infisical] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'infisical'
        client_name: 'Infisical'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://infisical.{{< sitevar name="domain" nojs="example.com" >}}/api/v1/sso/oidc/callback'
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
        token_endpoint_auth_method: 'client_secret_basic'
```

#### Configuration Escape Hatch

{{% oidc-escape-hatch-claims-hydration client_id="infisical" claims="email,name,given_name,family_name" %}}

### Application

To configure [Infisical] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [Infisical] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [Infisical] as an organization administrator.
2. Navigate to `Single Sign-On (SSO)`.
3. Navigate to the `General` tab.
4. Click `Connect` for `OIDC`.
5. Configure the following options:
   - Configuration Type: `Discovery URL`
   - Discovery Document URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration`
   - JWT Signature Algorithm: `RS256`
   - Client ID: `infisical`
   - Client Secret: `insecure_secret`
6. Click `Update`.
7. Enable `OIDC SSO` for the organization.

## See Also

- [Infisical General OIDC Documentation](https://infisical.com/docs/documentation/platform/sso/general-oidc/overview)
- [Infisical Email Domain Verification Documentation](https://infisical.com/docs/documentation/platform/email-domain)
- [Infisical Environment Variables Documentation](https://infisical.com/docs/self-hosting/configuration/envars)

[Authelia]: https://www.authelia.com
[Infisical]: https://infisical.com/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
