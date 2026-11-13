---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "SOGo"
description: "A guide on integrating SOGo with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "SOGo | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring SOGo with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [SOGo]
  - [v5.12.13](https://github.com/Alinto/sogo/releases/tag/SOGo-5.12.13)

{{% oidc-common bugs="client-credentials-encoding" %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://sogo.{{< sitevar name="domain" nojs="example.com" >}}/SOGo/`
  - This option determines the redirect URI in the format of
    `https://sogo.{{< sitevar name="domain" nojs="example.com" >}}/SOGo/`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `sogo`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [SOGo] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'sogo'
        client_name: 'SOGo'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://sogo.{{< sitevar name="domain" nojs="example.com" >}}/SOGo/'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

To configure [SOGo] there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

To configure [SOGo] to utilize Authelia as an [OpenID Connect 1.0] Provider, add the following options to the
`/etc/sogo/sogo.conf` configuration file:

```text {title="sogo.conf"}
  SOGoAuthenticationType = openid;
  SOGoXSRFValidationEnabled = NO;
  OCSOpenIdURL = "mysql://sogo:sogo@127.0.0.1:3306/sogo/sogo_openid";
  SOGoOpenIdConfigUrl = "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration";
  SOGoOpenIdClient = "sogo";
  SOGoOpenIdClientSecret = "insecure_secret";
  SOGoOpenIdScope = "openid profile email";
  SOGoOpenIdEmailParam = "email";
```

The `OCSOpenIdURL` option is the database URL used to store the OpenID Connect sessions and should be adjusted to match
your database.

[SOGo] only has the access token of the user and not their password, so the IMAP and SMTP servers must be configured to
accept the `xoauth2` mechanism using the `NGImap4AuthMechanism` and `SOGoSMTPAuthenticationType` options. See the
[SOGo Installation Guide](https://www.sogo.nu/files/docs/SOGoInstallationGuide.html) for more information.

## See Also

- [SOGo Installation Guide](https://www.sogo.nu/files/docs/SOGoInstallationGuide.html)

[Authelia]: https://www.authelia.com
[SOGo]: https://www.sogo.nu/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
