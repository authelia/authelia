---
title: "audiobookshelf"
description: "Integrating audiobookshelf with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-03-22T03:16:02+00:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/audiobookshelf/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "audiobookshelf | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring audiobookshelf with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.0](https://github.com/authelia/authelia/releases/tag/v4.39.0)
- [audiobookshelf]
  - [v2.26.3](https://github.com/advplyr/audiobookshelf/releases/tag/v2.26.3)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://audiobookshelf.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
        `https://audiobookshelf.{{< sitevar name="domain" nojs="example.com" >}}/login`.
        This means if you change this value, you need to update the redirect URI.
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `audiobookshelf`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [audiobookshelf] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    clients:
      - client_id: 'audiobookshelf'
        client_name: 'audiobookshelf'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://audiobookshelf.{{< sitevar name="domain" nojs="example.com" >}}/auth/openid/callback'
          - 'https://audiobookshelf.{{< sitevar name="domain" nojs="example.com" >}}/auth/openid/mobile-redirect'
          - 'audiobookshelf://oauth'
        scopes:
          - 'openid'
          - 'profile'
          - 'groups'
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

To configure [audiobookshelf] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [audiobookshelf] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Navigate to Settings.
2. Navigate to Authentication.
3. Configure the following options (some of these options can be automated by filling the `Issuer URL` and clicking
   `Auto-populate` and just verifying the value is correct):
   - OpenID Connect Authentication: Enabled
   - Issuer URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
   - Authorize URL (Auto-populate): `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization`
   - Token URL (Auto-populate): `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token`
   - Userinfo URL (Auto-populate): `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo`
   - JWKS URL (Auto-populate): `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/jwks.json`
   - Logout URL (Auto-populate): empty
   - Client ID: `audiobookshelf`
   - Client Secret: `insecure_secret`
   - Signing Algorithm: `RS256`
   - Allowed Mobile Redirect URIs: `audiobookshelf://oauth`
   - Subfolder for Redirect URLs: `None`
   - Button Text: `Login with Authelia`
   - Match existing users by: `Match by username`
   - Auto Launch: Disabled
   - Auto Register: Disabled

In addition to the configuration above you may want to consider enabling the Auto Launch and Auto Register features. It's
important to note that if you enable Auto Launch you will automatically be redirected to Authelia for consent regardless
if you have an account or not, and [audiobookshelf] does not seem to provide errors to the user when this happens.

Auto Registration is probably fine but if you only want some users to have access to [audiobookshelf] we suggest leaving
it off.

The groups claim can be configured as `groups` but you must make sure the groups expected by [audiobookshelf] exist for
the users you want to have access. This will also mean the group management will occur in Authelia, not [audiobookshelf]
presumably.

{{< figure src="audiobookshelf_1.png" alt="audiobookshelf_1" width="300" >}}

## See Also

- [audiobookshelf Authenticating With an OpenID Provider Documentation](https://www.audiobookshelf.org/guides/oidc_authentication/)

[Authelia]: https://www.authelia.com
[audiobookshelf]: https://www.audiobookshelf.org/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
