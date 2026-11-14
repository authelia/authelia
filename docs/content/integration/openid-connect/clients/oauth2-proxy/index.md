---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "OAuth2 Proxy"
description: "A guide on integrating OAuth2 Proxy with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "OAuth2 Proxy | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring OAuth2 Proxy with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [OAuth2 Proxy]
  - [v7.15.6](https://github.com/oauth2-proxy/oauth2-proxy/releases/tag/v7.15.6)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://oauth2-proxy.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://oauth2-proxy.{{< sitevar name="domain" nojs="example.com" >}}/oauth2/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `oauth2-proxy`
- **Client Secret:** `insecure_secret`
- **Upstream URL:** `http://app:8080/`
  - This is the application that [OAuth2 Proxy] protects.

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [OAuth2 Proxy] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'oauth2-proxy'
        client_name: 'OAuth2 Proxy'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://oauth2-proxy.{{< sitevar name="domain" nojs="example.com" >}}/oauth2/callback'
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

To configure [OAuth2 Proxy] there are two methods, using the [Configuration File](#configuration-file), or using
[Environment Variables](#environment-variables).

#### Configuration File

{{< callout context="tip" title="Did you know?" icon="outline/rocket" >}}
Generally the configuration file is passed to [OAuth2 Proxy] using the `--config` flag, for example
`oauth2-proxy --config /etc/oauth2-proxy/oauth2-proxy.cfg`.
{{< /callout >}}

To configure [OAuth2 Proxy] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```toml {title="oauth2-proxy.cfg"}
http_address = "0.0.0.0:4180"
upstreams = ["http://app:8080/"]
provider = "oidc"
provider_display_name = "Authelia"
oidc_issuer_url = "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}"
client_id = "oauth2-proxy"
client_secret = "insecure_secret"
redirect_url = "https://oauth2-proxy.{{< sitevar name="domain" nojs="example.com" >}}/oauth2/callback"
code_challenge_method = "S256"
scope = "openid email profile groups"
email_domains = ["*"]
## Generate a random value with 'openssl rand -base64 32 | tr -- '+/' '-_''.
cookie_secret = "a_random_32_byte_cookie_secret_!"
cookie_secure = true
```

#### Environment Variables

To configure [OAuth2 Proxy] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
OAUTH2_PROXY_HTTP_ADDRESS=0.0.0.0:4180
OAUTH2_PROXY_UPSTREAMS=http://app:8080/
OAUTH2_PROXY_PROVIDER=oidc
OAUTH2_PROXY_PROVIDER_DISPLAY_NAME=Authelia
OAUTH2_PROXY_OIDC_ISSUER_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
OAUTH2_PROXY_CLIENT_ID=oauth2-proxy
OAUTH2_PROXY_CLIENT_SECRET=insecure_secret
OAUTH2_PROXY_REDIRECT_URL=https://oauth2-proxy.{{< sitevar name="domain" nojs="example.com" >}}/oauth2/callback
OAUTH2_PROXY_CODE_CHALLENGE_METHOD=S256
OAUTH2_PROXY_SCOPE=openid email profile groups
OAUTH2_PROXY_EMAIL_DOMAINS=*
OAUTH2_PROXY_COOKIE_SECRET=a_random_32_byte_cookie_secret_!
OAUTH2_PROXY_COOKIE_SECURE=true
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  oauth2-proxy:
    environment:
      OAUTH2_PROXY_HTTP_ADDRESS: '0.0.0.0:4180'
      OAUTH2_PROXY_UPSTREAMS: 'http://app:8080/'
      OAUTH2_PROXY_PROVIDER: 'oidc'
      OAUTH2_PROXY_PROVIDER_DISPLAY_NAME: 'Authelia'
      OAUTH2_PROXY_OIDC_ISSUER_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      OAUTH2_PROXY_CLIENT_ID: 'oauth2-proxy'
      OAUTH2_PROXY_CLIENT_SECRET: 'insecure_secret'
      OAUTH2_PROXY_REDIRECT_URL: 'https://oauth2-proxy.{{< sitevar name="domain" nojs="example.com" >}}/oauth2/callback'
      OAUTH2_PROXY_CODE_CHALLENGE_METHOD: 'S256'
      OAUTH2_PROXY_SCOPE: 'openid email profile groups'
      OAUTH2_PROXY_EMAIL_DOMAINS: '*'
      OAUTH2_PROXY_COOKIE_SECRET: 'a_random_32_byte_cookie_secret_!'
      OAUTH2_PROXY_COOKIE_SECURE: 'true'
```

To restrict access to members of specific groups, set the `allowed_groups` configuration option
(`OAUTH2_PROXY_ALLOWED_GROUPS` environment variable) to a list of group names.

## See Also

- [OAuth2 Proxy OpenID Connect Provider Documentation](https://oauth2-proxy.github.io/oauth2-proxy/configuration/providers/openid_connect)
- [OAuth2 Proxy Configuration Overview](https://oauth2-proxy.github.io/oauth2-proxy/configuration/overview)

[Authelia]: https://www.authelia.com
[OAuth2 Proxy]: https://oauth2-proxy.github.io/oauth2-proxy/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
