---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Zerobyte"
description: "A guide on integrating Zerobyte with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Zerobyte | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Zerobyte with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Zerobyte]
  - [v0.44.2](https://github.com/nicotsx/zerobyte/releases/tag/v0.44.2)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://zerobyte.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://zerobyte.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/sso/callback/authelia`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `zerobyte`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Zerobyte] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'zerobyte'
        client_name: 'Zerobyte'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://zerobyte.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/sso/callback/authelia'
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

### Application

To configure [Zerobyte] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [Zerobyte] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [Zerobyte] as an organization owner.
2. Navigate to `Settings` > `Organization` > `Single Sign-On`.
3. Click `Register new`.
4. Configure the following options:
   - Provider ID: `authelia`
   - Organization Domain: `{{< sitevar name="domain" nojs="example.com" >}}`
   - Issuer URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
   - Discovery Endpoint: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration`
   - Client ID: `zerobyte`
   - Client Secret: `insecure_secret`
   - Link matching emails to existing accounts: Enable this option if existing [Zerobyte] accounts in the
     organization should be linked to Authelia users with a matching email address.
5. Click `Register Provider`.

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
[Zerobyte] single sign-on is invite only. Users who do not already belong to the organization must be invited before
they are able to login using Authelia.
{{< /callout >}}

## See Also

- [Zerobyte SSO Documentation](https://github.com/nicotsx/zerobyte/blob/main/apps/docs/content/docs/guides/oidc-sso.mdx)

[Authelia]: https://www.authelia.com
[Zerobyte]: https://github.com/nicotsx/zerobyte
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
