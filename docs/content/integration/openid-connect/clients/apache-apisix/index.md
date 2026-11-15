---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Apache APISIX"
description: "A guide on integrating Apache APISIX with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Apache APISIX | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Apache APISIX with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Apache APISIX]
  - [v3.20.0](https://github.com/apache/apisix/releases/tag/3.20.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://app.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This is the URL of the application protected by the [Apache APISIX] `openid-connect` plugin.
- **Redirect URI:** `https://app.{{< sitevar name="domain" nojs="example.com" >}}/oidc/callback`
  - This must be a path handled by the route the plugin is configured on, and must not be the same as a path of the
    protected application.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `apisix`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Apache APISIX] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'apisix'
        client_name: 'Apache APISIX'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://app.{{< sitevar name="domain" nojs="example.com" >}}/oidc/callback'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
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

To configure [Apache APISIX] there are two methods, using the [Admin API](#admin-api) or the
[Standalone Configuration File](#standalone-configuration-file).

The `session.secret` value is used to encrypt the [Apache APISIX] session and must be a random value of at least 16
characters.

#### Admin API

To configure [Apache APISIX] to utilize Authelia as an [OpenID Connect 1.0] Provider, create a route with the
`openid-connect` plugin using the Admin API:

```shell
curl "http://127.0.0.1:9180/apisix/admin/routes/app" -X PUT \
  -H "X-API-KEY: ${admin_key}" \
  -d '{
    "uri": "/*",
    "host": "app.{{< sitevar name="domain" nojs="example.com" >}}",
    "plugins": {
      "openid-connect": {
        "discovery": "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration",
        "client_id": "apisix",
        "client_secret": "insecure_secret",
        "scope": "openid profile email groups",
        "redirect_uri": "https://app.{{< sitevar name="domain" nojs="example.com" >}}/oidc/callback",
        "token_endpoint_auth_method": "client_secret_basic",
        "use_pkce": true,
        "session": {
          "secret": "insecure_session_secret"
        }
      }
    },
    "upstream": {
      "type": "roundrobin",
      "nodes": {
        "app:80": 1
      }
    }
  }'
```

#### Standalone Configuration File

To configure [Apache APISIX] to utilize Authelia as an [OpenID Connect 1.0] Provider when running in standalone mode,
add the following route to the `apisix.yaml` configuration file:

```yaml {title="apisix.yaml"}
routes:
  - id: 'app'
    uri: '/*'
    host: 'app.{{< sitevar name="domain" nojs="example.com" >}}'
    plugins:
      openid-connect:
        discovery: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration'
        client_id: 'apisix'
        client_secret: 'insecure_secret'
        scope: 'openid profile email groups'
        redirect_uri: 'https://app.{{< sitevar name="domain" nojs="example.com" >}}/oidc/callback'
        token_endpoint_auth_method: 'client_secret_basic'
        use_pkce: true
        session:
          secret: 'insecure_session_secret'
    upstream:
      type: 'roundrobin'
      nodes:
        'app:80': 1
#END
```

## See Also

- [Apache APISIX openid-connect Plugin Documentation](https://apisix.apache.org/docs/apisix/plugins/openid-connect/)

[Authelia]: https://www.authelia.com
[Apache APISIX]: https://apisix.apache.org/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
