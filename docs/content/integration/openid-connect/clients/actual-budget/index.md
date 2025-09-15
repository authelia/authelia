---
title: "Actual Budget"
description: "Integrating Actual Budget with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-01-25T12:36:00+11:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/actual-budget/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Actual Budget | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Actual Budget with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.38.18](https://github.com/authelia/authelia/releases/tag/v4.38.18)
- [Actual Budget]
  - [v25.1.0](https://github.com/actualbudget/actual/releases/tag/v25.1.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://actual-budget.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
        `https://actual-budget.{{< sitevar name="domain" nojs="example.com" >}}/login`.
        This means if you change this value, you need to update the redirect URI.
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `actual-budget`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Actual Budget] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'actual-budget'
        client_name: 'Actual Budget'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://actual-budget.{{< sitevar name="domain" nojs="example.com" >}}/openid/callback'
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

To configure [Actual Budget] there are three methods, using the [Configuration File](#configuration-file), using
[Environment Variables](#environment-variables), or using the [Web GUI](#web-gui).

#### Configuration File

To configure [Actual Budget] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```json
{
  "openId": {
    "discoveryURL": "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}",
    "client_id": "actual-budget",
    "client_secret": "insecure_secret",
    "server_hostname": "https://actual-budget.{{< sitevar name="domain" nojs="example.com" >}}",
    "authMethod": "oauth2"
  }
}
```

#### Environment Variables

To configure [Actual Budget] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
ACTUAL_OPENID_DISCOVERY_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
ACTUAL_OPENID_CLIENT_ID=actual-budget
ACTUAL_OPENID_CLIENT_SECRET=insecure_secret
ACTUAL_OPENID_SERVER_HOSTNAME=https://actual-budget.{{< sitevar name="domain" nojs="example.com" >}}
ACTUAL_OPENID_AUTH_METHOD=oauth2
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  actual-budget:
    environment:
      ACTUAL_OPENID_DISCOVERY_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration'
      ACTUAL_OPENID_CLIENT_ID: 'actual-budget'
      ACTUAL_OPENID_CLIENT_SECRET: 'insecure_secret'
      ACTUAL_OPENID_SERVER_HOSTNAME: 'https://actual-budget.{{< sitevar name="domain" nojs="example.com" >}}'
      ACTUAL_OPENID_AUTH_METHOD: 'oauth2'
```

#### Web GUI

To configure [Actual Budget] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Navigate to any Budget file.
2. Navigate to Settings.
3. Click on Start Using OpenID.
4. Configure the following options:
   - OpenID Provider: `Other`
   - OpenID Provider URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
   - Client ID: `actual-budget`
   - Client Secret: `insecure_secret`
5. Click OK.

## See Also

- [Actual Budget Authenticating With an OpenID Provider Documentation](https://actualbudget.org/docs/config/oauth-auth)

[Authelia]: https://www.authelia.com
[Actual Budget]: https://actualbudget.org/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
