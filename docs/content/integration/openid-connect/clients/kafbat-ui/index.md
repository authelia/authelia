---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Kafbat UI"
description: "A guide on integrating Kafbat UI with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Kafbat UI | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Kafbat UI with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Kafbat UI]
  - [v1.5.0](https://github.com/kafbat/kafka-ui/releases/tag/v1.5.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://kafbat-ui.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://kafbat-ui.{{< sitevar name="domain" nojs="example.com" >}}/login/oauth2/code/authelia`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `kafbat-ui`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Kafbat UI] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'kafbat-ui'
        client_name: 'Kafbat UI'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://kafbat-ui.{{< sitevar name="domain" nojs="example.com" >}}/login/oauth2/code/authelia'
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

To configure [Kafbat UI] there are two methods, using the [Configuration File](#configuration-file), or using
[Environment Variables](#environment-variables).

#### Configuration File

To configure [Kafbat UI] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```yaml {title="config.yml"}
auth:
  type: 'OAUTH2'
  oauth2:
    client:
      authelia:
        clientId: 'kafbat-ui'
        clientSecret: 'insecure_secret'
        scope:
          - 'openid'
          - 'profile'
          - 'email'
          - 'groups'
        client-name: 'Authelia'
        provider: 'authelia'
        redirect-uri: 'https://kafbat-ui.{{< sitevar name="domain" nojs="example.com" >}}/login/oauth2/code/authelia'
        authorization-grant-type: 'authorization_code'
        issuer-uri: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
        user-name-attribute: 'preferred_username'
        custom-params:
          type: 'oauth'
          roles-field: 'groups'
```

#### Environment Variables

To configure [Kafbat UI] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
AUTH_TYPE=OAUTH2
AUTH_OAUTH2_CLIENT_AUTHELIA_CLIENTID=kafbat-ui
AUTH_OAUTH2_CLIENT_AUTHELIA_CLIENTSECRET=insecure_secret
AUTH_OAUTH2_CLIENT_AUTHELIA_SCOPE=openid,profile,email,groups
AUTH_OAUTH2_CLIENT_AUTHELIA_CLIENTNAME=Authelia
AUTH_OAUTH2_CLIENT_AUTHELIA_PROVIDER=authelia
AUTH_OAUTH2_CLIENT_AUTHELIA_REDIRECTURI=https://kafbat-ui.{{< sitevar name="domain" nojs="example.com" >}}/login/oauth2/code/authelia
AUTH_OAUTH2_CLIENT_AUTHELIA_AUTHORIZATIONGRANTTYPE=authorization_code
AUTH_OAUTH2_CLIENT_AUTHELIA_ISSUERURI=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
AUTH_OAUTH2_CLIENT_AUTHELIA_USERNAMEATTRIBUTE=preferred_username
AUTH_OAUTH2_CLIENT_AUTHELIA_CUSTOMPARAMS_TYPE=oauth
AUTH_OAUTH2_CLIENT_AUTHELIA_CUSTOMPARAMS_ROLESFIELD=groups
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  kafbat-ui:
    environment:
      AUTH_TYPE: 'OAUTH2'
      AUTH_OAUTH2_CLIENT_AUTHELIA_CLIENTID: 'kafbat-ui'
      AUTH_OAUTH2_CLIENT_AUTHELIA_CLIENTSECRET: 'insecure_secret'
      AUTH_OAUTH2_CLIENT_AUTHELIA_SCOPE: 'openid,profile,email,groups'
      AUTH_OAUTH2_CLIENT_AUTHELIA_CLIENTNAME: 'Authelia'
      AUTH_OAUTH2_CLIENT_AUTHELIA_PROVIDER: 'authelia'
      AUTH_OAUTH2_CLIENT_AUTHELIA_REDIRECTURI: 'https://kafbat-ui.{{< sitevar name="domain" nojs="example.com" >}}/login/oauth2/code/authelia'
      AUTH_OAUTH2_CLIENT_AUTHELIA_AUTHORIZATIONGRANTTYPE: 'authorization_code'
      AUTH_OAUTH2_CLIENT_AUTHELIA_ISSUERURI: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      AUTH_OAUTH2_CLIENT_AUTHELIA_USERNAMEATTRIBUTE: 'preferred_username'
      AUTH_OAUTH2_CLIENT_AUTHELIA_CUSTOMPARAMS_TYPE: 'oauth'
      AUTH_OAUTH2_CLIENT_AUTHELIA_CUSTOMPARAMS_ROLESFIELD: 'groups'
```

## See Also

- [Kafbat UI OAuth2 Documentation](https://ui.docs.kafbat.io/configuration/authentication/oauth2)

[Authelia]: https://www.authelia.com
[Kafbat UI]: https://github.com/kafbat/kafka-ui
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
