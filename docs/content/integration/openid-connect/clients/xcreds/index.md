---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "XCreds"
description: "A guide on integrating XCreds with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "XCreds | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring XCreds with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [XCreds]
  - [v5.9](https://github.com/twocanoes/xcreds/releases/tag/tag-5.9%289148%29)

{{% oidc-common bugs="client-credentials-encoding,claims-hydration" %}}

### Assumptions

This example makes the following assumptions:

- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `xcreds`
- **Client Secret:** `insecure_secret`
- **Redirect URI:** `https://127.0.0.1/xcreds`
  - This value is intercepted by [XCreds] and is never actually requested, it must however match the `redirectURI`
    preference.

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [XCreds] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'xcreds'
        client_name: 'XCreds'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'native'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://127.0.0.1/xcreds'
        scopes:
          - 'openid'
          - 'offline_access'
          - 'profile'
          - 'email'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
          - 'refresh_token'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_post'
```

#### Configuration Escape Hatch

{{% oidc-escape-hatch-claims-hydration client_id="xcreds" claims="email,preferred_username,name,given_name,family_name" %}}

### Application

To configure [XCreds] there is one method, using a [Configuration Profile](#configuration-profile).

#### Configuration Profile

[XCreds] uses the refresh token to detect password changes, which is why the `offline_access` scope and the
`refresh_token` grant type are required.

To configure [XCreds] to utilize Authelia as an [OpenID Connect 1.0] Provider, create a configuration profile with a
payload for the `com.twocanoes.xcreds` preference domain, for example using
[ProfileCreator](https://github.com/ProfileCreator/ProfileCreator), which contains the following preferences:

```xml {title="com.twocanoes.xcreds.plist"}
<key>clientID</key>
<string>xcreds</string>
<key>clientSecret</key>
<string>insecure_secret</string>
<key>discoveryURL</key>
<string>https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration</string>
<key>redirectURI</key>
<string>https://127.0.0.1/xcreds</string>
<key>scopes</key>
<string>openid offline_access profile email</string>
<key>map_username</key>
<string>preferred_username</string>
<key>map_fullname</key>
<string>name</string>
<key>map_firstname</key>
<string>given_name</string>
<key>map_lastname</key>
<string>family_name</string>
```

Install the configuration profile on the target Mac with a `System` scope, either using your device management
solution or by navigating to `System Settings` > `General` > `Device Management`.

## See Also

- [XCreds Admin Guide](https://twocanoes.com/knowledge-base/xcreds-admin-guide/)
- [XCreds Identity Provider Integration Guide](https://twocanoes.com/knowledge-base/xcreds-identity-provider-integration-guide/)

[Authelia]: https://www.authelia.com
[XCreds]: https://twocanoes.com/products/mac/xcreds/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
