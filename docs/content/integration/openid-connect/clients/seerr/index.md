---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Seerr"
description: "A guide on integrating Seerr with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Seerr | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Seerr with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Seerr]
  - [development version tag:preview-new-oidc](https://github.com/seerr-team/seerr/tree/preview-new-oidc)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://seerr.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URIs in the format of
    `https://seerr.{{< sitevar name="domain" nojs="example.com" >}}/login` and
    `https://seerr.{{< sitevar name="domain" nojs="example.com" >}}/profile/settings/linked-accounts`.
    This means if you change this value, you need to update the redirect URIs.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `seerr`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Seerr] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'seerr'
        client_name: 'Seerr'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://seerr.{{< sitevar name="domain" nojs="example.com" >}}/login'
          - 'https://seerr.{{< sitevar name="domain" nojs="example.com" >}}/profile/settings/linked-accounts'
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

To configure [Seerr] there is one method, using the [Configuration File](#configuration-file).

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
At the time of this writing [OpenID Connect 1.0] support in [Seerr] is only available in the `preview-new-oidc` preview
build. Back up your Seerr configuration before changing it. If Seerr is accessed via a reverse proxy over HTTPS you
should also either set the `Application URL` or enable `Enable Proxy Support` in the Seerr settings so the correct
redirect URI is sent to Authelia.
{{< /callout >}}

#### Configuration File

To configure [Seerr] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Stop [Seerr].
2. Edit the [Seerr] `settings.json` file which is typically located in the `/app/config` directory when using Docker.
3. Set the `main.oidcLogin` option to `true` and add the following provider to the `oidc.providers` list:

```json {title="settings.json"}
{
  "main": {
    "oidcLogin": true
  },
  "oidc": {
    "providers": [
      {
        "slug": "authelia",
        "name": "Authelia",
        "issuerUrl": "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}",
        "clientId": "seerr",
        "clientSecret": "insecure_secret",
        "logo": "https://www.authelia.com/images/branding/logo-cropped.png",
        "scopes": "openid profile email",
        "newUserLogin": true
      }
    ]
  }
}
```

4. Start [Seerr].

An `Authelia` button is then displayed on the Seerr login page. Existing users can link their Authelia account by
navigating to `Profile`, `Settings`, `Linked Accounts`, and clicking `Link Account`.

## See Also

- [Seerr OpenID Connect Documentation](https://github.com/seerr-team/seerr/blob/preview-new-oidc/docs/using-seerr/settings/users/oidc.md)
- [Seerr OpenID Connect Support Discussion](https://github.com/seerr-team/seerr/discussions/2721)

[Authelia]: https://www.authelia.com
[Seerr]: https://seerr.dev/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
