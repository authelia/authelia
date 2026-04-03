---
title: "FreshRSS"
description: "Integrating FreshRSS with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/freshrss/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "FreshRSS | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring FreshRSS with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.38.0](https://github.com/authelia/authelia/releases/tag/v4.38.0)
- [FreshRSS]
  - [v1.23.1](https://github.com/FreshRSS/FreshRSS/releases/tag/1.23.1)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://freshrss.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `freshrss`
- __Client Secret:__ `insecure_secret`
- __Port:__ '443'
  - This is the port [FreshRSS] is served over (usually 80 for http and 443 for https) NOT the port of the container.

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

### Special Notes

1. The [FreshRSS] implementation seems to always include the port in the requested `redirect_uri`. As Authelia strictly
   conforms to the specifications this means the client registration **_MUST_** include the port for the requested
   `redirect_uri` to match.

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [FreshRSS] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'freshrss'
        client_name: 'FreshRSS'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://freshrss.{{< sitevar name="domain" nojs="example.com" >}}:443/i/oidc/'
        scopes:
          - 'openid'
          - 'groups'
          - 'email'
          - 'profile'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [FreshRSS] there is one method, using the [Environment Variables](#environment-variables).

#### Environment Variables

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
The below example uses `insecure_crypto_key` for one of the values. It's recommended that this value is configured
according to the FreshRSS recommendations. At minimum this should be a reasonably long random string.
{{< /callout >}}

To configure [FreshRSS] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment variables:

##### Standard

```shell {title=".env"}
OIDC_ENABLED=1
OIDC_PROVIDER_METADATA_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration
OIDC_CLIENT_ID=freshrss
OIDC_CLIENT_SECRET=insecure_secret
OIDC_CLIENT_CRYPTO_KEY=insecure_crypto_key
OIDC_REMOTE_USER_CLAIM=preferred_username
OIDC_SCOPES=openid groups email profile
OIDC_X_FORWARDED_HEADERS=X-Forwarded-Host X-Forwarded-Port X-Forwarded-Proto
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  freshrss:
    environment:
      OIDC_ENABLED: '1'
      OIDC_PROVIDER_METADATA_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration'
      OIDC_CLIENT_ID: 'freshrss'
      OIDC_CLIENT_SECRET: 'insecure_secret'
      OIDC_CLIENT_CRYPTO_KEY: 'insecure_crypto_key'
      OIDC_REMOTE_USER_CLAIM: 'preferred_username'
      OIDC_SCOPES: 'openid groups email profile'
      OIDC_X_FORWARDED_HEADERS: 'X-Forwarded-Host X-Forwarded-Port X-Forwarded-Proto'
```

In addition, the following steps may be required:

1. Open the newly created [FreshRSS] instance.
2. During initial config, select "HTTP" during the user creation.

## See Also

- [FreshRSS OIDC documentation](https://freshrss.github.io/FreshRSS/en/admins/16_OpenID-Connect.html)

[Authelia]: https://www.authelia.com
[FreshRSS]: https://freshrss.github.io/FreshRSS/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
