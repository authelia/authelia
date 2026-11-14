---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Plane"
description: "A guide on integrating Plane with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Plane | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Plane with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Plane] (Commercial Edition)
  - [v3.3.3](https://plane.so/changelog/commercial)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://plane.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://plane.{{< sitevar name="domain" nojs="example.com" >}}/auth/oidc/callback/`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `plane`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

### Special Notes

1. The [Plane] [OpenID Connect 1.0] integration requires a self-hosted [Plane] Pro or Business plan. This guide covers
   self-hosted [Plane] instances only.

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Plane] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'plane'
        client_name: 'Plane'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://plane.{{< sitevar name="domain" nojs="example.com" >}}/auth/oidc/callback/'
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

To configure [Plane] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [Plane] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [Plane] as an instance administrator.
2. Navigate to `God Mode` by visiting `https://plane.{{< sitevar name="domain" nojs="example.com" >}}/god-mode/`.
3. Navigate to `Authentication` > `OIDC`.
4. Configure the following options:
   - Client ID: `plane`
   - Client secret: `insecure_secret`
   - Authorize URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization`
   - Token URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token`
   - Users' info URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo`
   - IdP's name: `Authelia`
5. Click `Save changes`.
6. Toggle OIDC to the enabled position at the top of the page.

## See Also

- [Plane OIDC SSO Documentation](https://developers.plane.so/self-hosting/govern/oidc-sso)
- [Plane Single Sign-On Documentation](https://docs.plane.so/authentication/sso)

[Authelia]: https://www.authelia.com
[Plane]: https://plane.so/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
