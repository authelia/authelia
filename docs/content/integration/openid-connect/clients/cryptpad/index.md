---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "CryptPad"
description: "A guide on integrating CryptPad with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "CryptPad | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring CryptPad with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [CryptPad]
  - [2026.10.1](https://github.com/cryptpad/cryptpad/releases/tag/2026.10.1)
- [CryptPad SSO Plugin]
  - [1.0.0](https://github.com/cryptpad/sso/releases/tag/1.0.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://cryptpad.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://cryptpad.{{< sitevar name="domain" nojs="example.com" >}}/ssoauth`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `cryptpad`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [CryptPad] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'cryptpad'
        client_name: 'CryptPad'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://cryptpad.{{< sitevar name="domain" nojs="example.com" >}}/ssoauth'
        scopes:
          - 'openid'
          - 'profile'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        id_token_signed_response_alg: 'RS256'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [CryptPad] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [CryptPad] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Install the [CryptPad SSO Plugin] by cloning it into the `lib/plugins` directory of your [CryptPad] installation
   and running `npm ci`, then restart [CryptPad].
2. Login to [CryptPad] as an administrator.
3. Navigate to the `Administration` panel.
4. Navigate to the `Single Sign-On` category.
5. Enable `Enable SSO login on this CryptPad instance`.
6. Under `Add new SSO provider` configure the following options:
   - Type: `OpenID Connect`
   - Identifier: `Authelia`
   - Base URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
   - Client ID: `cryptpad`
   - Client secret: `insecure_secret`
   - ID token algorithm: `RS256`
   - Use PKCE code challenge: Enabled
   - Use nonce: Enabled
   - Username Scope: `profile`
   - Username Claim: `name`
7. Save the provider.

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
The `ID token algorithm` must match the `id_token_signed_response_alg` configured for the client in Authelia. When it
is left blank the plugin defaults to `PS256`, which will cause the ID Token validation to fail.
{{< /callout >}}

## See Also

- [CryptPad SSO Plugin Documentation](https://github.com/cryptpad/sso)

[Authelia]: https://www.authelia.com
[CryptPad]: https://cryptpad.org/
[CryptPad SSO Plugin]: https://github.com/cryptpad/sso
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
