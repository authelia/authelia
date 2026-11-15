---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Gokapi"
description: "A guide on integrating Gokapi with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Gokapi | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Gokapi with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Gokapi]
  - [v2.2.4](https://github.com/Forceu/Gokapi/releases/tag/v2.2.4)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://gokapi.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://gokapi.{{< sitevar name="domain" nojs="example.com" >}}/oauth-callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `gokapi`
- **Client Secret:** `insecure_secret`
- **Authorized Groups:** `gokapi-admins`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Gokapi] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'gokapi'
        client_name: 'Gokapi'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://gokapi.{{< sitevar name="domain" nojs="example.com" >}}/oauth-callback'
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
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [Gokapi] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [Gokapi] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Open the [Gokapi] setup wizard. This is shown on the first start, or after starting [Gokapi] with the
   `--reconfigure` argument which prints temporary credentials for the setup page to the console.
2. Proceed through the wizard until the authentication step.
3. Select `OAuth2 OpenID Connect` as the login method.
4. Configure the following options:
   - Provider URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
   - Client ID: `gokapi`
   - Client Secret: `insecure_secret`
   - Admin email address: The email address of the [Gokapi] super admin user.
   - Recheck identity every: `12 hours`
   - Restrict to Groups: Enabled
     - Scope identifier: `groups`
     - Authorised groups: `gokapi-admins`
5. Ensure the `Redirection URL` shown by the wizard is
   `https://gokapi.{{< sitevar name="domain" nojs="example.com" >}}/oauth-callback`.
6. Complete the setup wizard.

Multiple authorized groups can be separated with a semicolon, and `*` can be used as a wildcard, for example
`gokapi-admins;gokapi-*`.

## See Also

- [Gokapi OpenID Connect Examples Documentation](https://gokapi.readthedocs.io/en/stable/examples.html)
- [Gokapi Setup Documentation](https://gokapi.readthedocs.io/en/stable/setup.html)

[Authelia]: https://www.authelia.com
[Gokapi]: https://github.com/Forceu/Gokapi
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
