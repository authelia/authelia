---
title: "NetBird"
description: "Integrating NetBird with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-01-25T10:04:53+11:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/netbird/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "NetBird | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring NetBird with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.38.8](https://github.com/authelia/authelia/releases/tag/v4.38.8)
- [NetBird]
  - [v0.36.3](https://github.com/netbirdio/netbird/releases/tag/v0.36.3)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://netbird.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
        `https://netbird.{{< sitevar name="domain" nojs="example.com" >}}/login`.
        This means if you change this value, you need to update the redirect URI.
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `netbird`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [NetBird] which will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    cors:
      allowed_origins_from_client_redirect_uris: true
      endpoints:
        - 'userinfo'
        - 'authorization'
        - 'token'
        - 'revocation'
        - 'introspection'
    clients:
      - client_id: 'netbird'
        client_name: 'NetBird'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://netbird.{{< sitevar name="domain" nojs="example.com" >}}/peers'
          - 'https://netbird.{{< sitevar name="domain" nojs="example.com" >}}/add-peers'
          - 'http://localhost'
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

To configure [NetBird] to utilize Authelia as an [OpenID Connect 1.0] Provider you have to update a number of areas to
configure it for Authelia.

#### NetBird Dashboard

To configure [NetBird] Dashboard to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
NETBIRD_MGMT_API_ENDPOINT=https://netbird.{{< sitevar name="domain" nojs="example.com" >}}
NETBIRD_MGMT_GRPC_API=https://netbird.{{< sitevar name="domain" nojs="example.com" >}}
AUTH_AUDIENCE=none
AUTH_CLIENT_ID=netbird
AUTH_CLIENT_SECRET=insecure_secret
AUTH_AUTHORITY=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
USE_AUTH0=false
AUTH_SUPPORTED_SCOPES=openid email profile
AUTH_REDIRECT_URI=/peers
AUTH_SILENT_REDIRECT_URI=/add-peers
NETBIRD_TOKEN_SOURCE=idToken
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  netbird-dashboard:
    environment:
      NETBIRD_MGMT_API_ENDPOINT: 'https://netbird.{{< sitevar name="domain" nojs="example.com" >}}'
      NETBIRD_MGMT_GRPC_API: 'https://netbird.{{< sitevar name="domain" nojs="example.com" >}}'
      AUTH_AUDIENCE: 'none'
      AUTH_CLIENT_ID: 'netbird'
      AUTH_CLIENT_SECRET: 'insecure_secret'
      AUTH_AUTHORITY: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      USE_AUTH0: 'false'
      AUTH_SUPPORTED_SCOPES: 'openid email profile'
      AUTH_REDIRECT_URI: '/peers'
      AUTH_SILENT_REDIRECT_URI: '/add-peers'
      NETBIRD_TOKEN_SOURCE: 'idToken'
```

#### NetBird Management

{{< callout context="tip" title="Did you know?" icon="outline/rocket" >}}
Generally the configuration file is named `management.json`.
{{< /callout >}}

To configure [NetBird] Management to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following
configuration:

```json {title="management.json"}
{
  "HttpConfig": {
    "AuthIssuer": "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}",
    "AuthAudience": "netbird",
    "AuthKeysLocation": "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/jwks.json",
    "AuthUserIDClaim": "",
    "CertFile": "",
    "CertKey": "",
    "IdpSignKeyRefreshEnabled": true,
    "OIDCConfigEndpoint": "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration"
  },
  "IdpManagerConfig": {},
  "DeviceAuthorizationFlow": {},
  "PKCEAuthorizationFlow": {
    "ProviderConfig": {
      "Audience": "netbird",
      "ClientID": "netbird",
      "ClientSecret": "insecure_secret",
      "Domain": "",
      "AuthorizationEndpoint": "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization",
      "TokenEndpoint": "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token",
      "Scope": "openid email profile",
      "RedirectURLs": [
        "http://localhost:53000"
      ],
      "UseIDToken": true
    }
  }
}
```

## See Also

- [NetBird Identity Providers Documentation](https://docs.netbird.io/selfhosted/identity-providers)

[Authelia]: https://www.authelia.com
[NetBird]: https://netbird.io/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
