---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Appsmith"
description: "A guide on integrating Appsmith with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Appsmith | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Appsmith with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Appsmith]
  - [v2.6.0](https://github.com/appsmithorg/appsmith/releases/tag/v2.6.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://appsmith.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://appsmith.{{< sitevar name="domain" nojs="example.com" >}}/login/oauth2/code/oidc`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `appsmith`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

### Special Notes

1. The [Appsmith] [OpenID Connect 1.0] integration is only available in the paid editions of [Appsmith] and requires a
   self-hosted instance with a custom domain.

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Appsmith] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'appsmith'
        client_name: 'Appsmith'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://appsmith.{{< sitevar name="domain" nojs="example.com" >}}/login/oauth2/code/oidc'
        scopes:
          - 'openid'
          - 'offline_access'
          - 'email'
          - 'profile'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
          - 'refresh_token'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [Appsmith] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [Appsmith] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [Appsmith] as an administrator.
2. Navigate to `Admin Settings` > `Authentication`.
3. Under `OIDC`, click `Enable`.
4. Confirm the `Redirect URL` is `https://appsmith.{{< sitevar name="domain" nojs="example.com" >}}/login/oauth2/code/oidc`.
5. Configure the following options:
   - Client ID: `appsmith`
   - Client Secret: `insecure_secret`
   - Authorization URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization`
   - Token URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token`
   - User Info URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo`
   - JWK Set URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/jwks.json`
   - Scope: `openid`, `email`, `profile`, `offline_access`
   - Username Attribute: `email`
6. Click `Save & Restart` and wait for [Appsmith] to restart.

## See Also

- [Appsmith OpenID Connect (OIDC) Documentation](https://docs.appsmith.com/getting-started/setup/instance-configuration/authentication/openid-connect-oidc)

[Authelia]: https://www.authelia.com
[Appsmith]: https://www.appsmith.com/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
