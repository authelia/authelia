---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "mod_auth_openidc"
description: "A guide on integrating mod_auth_openidc with the Authelia OpenID Connect 1.0 Provider for SSO, with configuration examples and available options."
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
  title: "mod_auth_openidc | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring mod_auth_openidc with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [mod_auth_openidc]
  - [v2.4.20.4](https://github.com/OpenIDC/mod_auth_openidc/releases/tag/v2.4.20.4)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://app.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This is the site served by the [Apache HTTP Server] which is protected by [mod_auth_openidc].
- **Redirect URI:** `https://app.{{< sitevar name="domain" nojs="example.com" >}}/oauth2/callback`
  - This is a vanity URL handled by [mod_auth_openidc] which must be within a location protected by the module and
    must not serve any content.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `mod-auth-openidc`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [mod_auth_openidc]
which will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'mod-auth-openidc'
        client_name: 'mod_auth_openidc'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://app.{{< sitevar name="domain" nojs="example.com" >}}/oauth2/callback'
        scopes:
          - 'openid'
          - 'email'
          - 'profile'
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

To configure [mod_auth_openidc] there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

To configure [mod_auth_openidc] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following
[Apache HTTP Server] configuration:

```apache {title="httpd.conf"}
LoadModule auth_openidc_module modules/mod_auth_openidc.so

<VirtualHost *:443>
    ServerName app.{{< sitevar name="domain" nojs="example.com" >}}

    OIDCProviderMetadataURL https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration
    OIDCClientID mod-auth-openidc
    OIDCClientSecret insecure_secret
    OIDCRedirectURI https://app.{{< sitevar name="domain" nojs="example.com" >}}/oauth2/callback
    OIDCCryptoPassphrase <a random secret>
    OIDCScope "openid email profile groups"
    OIDCPKCEMethod S256
    OIDCProviderTokenEndpointAuth client_secret_basic
    OIDCRemoteUserClaim preferred_username

    <Location />
        AuthType openid-connect
        Require valid-user
    </Location>
</VirtualHost>
```

{{< callout context="tip" title="Did you know?" icon="outline/rocket" >}}
Access can be restricted to members of a specific group by replacing `Require valid-user` with a claim requirement such
as `Require claim groups:admins`.
{{< /callout >}}

## See Also

- [mod_auth_openidc Documentation](https://github.com/OpenIDC/mod_auth_openidc/wiki)
- [mod_auth_openidc Configuration Reference](https://github.com/OpenIDC/mod_auth_openidc/blob/master/auth_openidc.conf)

[Authelia]: https://www.authelia.com
[mod_auth_openidc]: https://github.com/OpenIDC/mod_auth_openidc
[Apache HTTP Server]: https://httpd.apache.org/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
