---
title: "Firezone"
description: "Integrating Firezone with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/firezone/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Firezone | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Firezone with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.38.0](https://github.com/authelia/authelia/releases/tag/v4.38.0)
- [Firezone]
  - [v0.7.25](https://github.com/firezone/firezone/releases/tag/0.7.25)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://firezone.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `firezone`
- __Client Secret:__ `insecure_secret`
- __Config ID (Firezone):__ `authelia`:
    - This option determines the redirect URI in the format of
      `https://firezone.{{< sitevar name="domain" nojs="example.com" >}}/auth/oidc/<Config ID>/callback`.
      This means if you change this value you need to update the redirect URI.

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Firezone] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'firezone'
        client_name: 'Firezone'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://firezone.{{< sitevar name="domain" nojs="example.com" >}}/auth/oidc/authelia/callback'
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

To configure [Firezone] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [Firezone] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following
instructions:

1. Visit your [Firezone] site
2. Sign in as an admin
3. Visit:
    1. Settings
    2. Security
4. In the `Single Sign-On` section, click on the `Add OpenID Connect Provider` button
5. Configure the following options:
   - Config ID: `authelia`
   - Label: `Authelia`
   - Scope: `openid email profile`
   - Client ID: `firezone`
   - Client secret: `insecure_secret`
   - Discovery Document URI: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration`
   - Redirect URI (optional): `https://firezone.{{< sitevar name="domain" nojs="example.com" >}}/auth/oidc/authelia/callback`
   - Auto-create users (checkbox): `true`

{{< figure src="firezone.png" alt="Firezone" width="500" >}}

## See Also

- [Firezone OIDC documentation](https://www.firezone.dev/docs/authenticate/oidc/)

[Authelia]: https://www.authelia.com
[Firezone]: https://www.firezone.dev
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
