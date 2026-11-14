---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Icinga Web 2"
description: "A guide on integrating Icinga Web 2 with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Icinga Web 2 | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Icinga Web 2 with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Icinga Web 2]
  - [v2.14.1](https://github.com/Icinga/icingaweb2/releases/tag/v2.14.1)
- [Icinga Web 2 OIDC Module]
  - [v0.7.7](https://github.com/RISE-GmbH/icingaweb2-module-oidc/releases/tag/v0.7.7)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://icinga.{{< sitevar name="domain" nojs="example.com" >}}/icingaweb2/`
  - This option determines the redirect URI in the format of
    `https://icinga.{{< sitevar name="domain" nojs="example.com" >}}/icingaweb2/oidc/authentication/realm?name=authelia`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `icinga-web-2`
- **Client Secret:** `insecure_secret`
- **Provider Name:** `authelia`
  - This option determines the value of the `name` query parameter in the redirect URI. This means if you change this
    value, you need to update the redirect URI.

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Icinga Web 2] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'icinga-web-2'
        client_name: 'Icinga Web 2'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://icinga.{{< sitevar name="domain" nojs="example.com" >}}/icingaweb2/oidc/authentication/realm?name=authelia'
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

To configure [Icinga Web 2] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
This guide assumes the [Icinga Web 2 OIDC Module] is already installed, enabled, and configured with a database
resource. See the [module installation documentation](https://github.com/RISE-GmbH/icingaweb2-module-oidc/blob/main/doc/02-Installation.md)
for more information.
{{< /callout >}}

To configure [Icinga Web 2] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [Icinga Web 2] as an administrator.
2. Navigate to `Oidc` > `Provider`.
3. Click `New Provider`.
4. Configure the following options:
   - Name: `authelia`
   - Url: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
   - Secret: `insecure_secret`
   - Appname: `icinga-web-2`
   - Caption: `Authelia`
   - Custom Username: `preferred_username`
   - Groups to sync: A comma separated list of the groups which should be imported into [Icinga Web 2] for role
     assignments, for example `icinga-*`.
   - PKCE client (S256): Enabled
   - Enabled: Enabled
   - Enforce Https on redirect urls: Enable this option if [Icinga Web 2] runs behind a reverse proxy which terminates
     TLS.
5. Click `Create Provider`.

Users and synchronized groups are created in [Icinga Web 2] without any permissions on their first login. Roles can be
assigned to them by navigating to `Configuration` > `Authentication` > `Roles`.

## See Also

- [Icinga Web 2 OIDC Module Configuration Documentation](https://github.com/RISE-GmbH/icingaweb2-module-oidc/blob/main/doc/03-Configuration.md)
- [Icinga Web 2 Access Control Documentation](https://icinga.com/docs/icinga-web/latest/doc/06-Security/)

[Authelia]: https://www.authelia.com
[Icinga Web 2]: https://icinga.com/
[Icinga Web 2 OIDC Module]: https://github.com/RISE-GmbH/icingaweb2-module-oidc
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
