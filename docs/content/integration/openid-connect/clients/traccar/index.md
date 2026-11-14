---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Traccar"
description: "A guide on integrating Traccar with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Traccar | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Traccar with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Traccar]
  - [v6.17.0](https://github.com/traccar/traccar/releases/tag/v6.17.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://traccar.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://traccar.{{< sitevar name="domain" nojs="example.com" >}}/api/session/openid/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `traccar`
- **Client Secret:** `insecure_secret`
- **Admin Group Name:** `traccar-admins`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Traccar] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'traccar'
        client_name: 'Traccar'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://traccar.{{< sitevar name="domain" nojs="example.com" >}}/api/session/openid/callback'
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

To configure [Traccar] there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

To configure [Traccar] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```xml {title="traccar.xml"}
<?xml version='1.0' encoding='UTF-8'?>
<!DOCTYPE properties SYSTEM 'http://java.sun.com/dtd/properties.dtd'>
<properties>
    <entry key='web.url'>https://traccar.{{< sitevar name="domain" nojs="example.com" >}}</entry>
    <entry key='openid.issuerUrl'>https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}</entry>
    <entry key='openid.clientId'>traccar</entry>
    <entry key='openid.clientSecret'>insecure_secret</entry>
    <entry key='openid.adminGroup'>traccar-admins</entry>
</properties>
```

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
[Traccar] only requests the `groups` scope when the `openid.adminGroup` option is configured. The optional
`openid.allowGroup` option can be used to restrict access to members of a specific group.
{{< /callout >}}

## See Also

- [Traccar OpenID Connect Documentation](https://www.traccar.org/openid-sso/)

[Authelia]: https://www.authelia.com
[Traccar]: https://www.traccar.org/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
