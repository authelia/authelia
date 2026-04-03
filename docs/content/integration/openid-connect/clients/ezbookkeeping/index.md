---
title: "ezBookkeeping"
description: "Integrating ezBookkeeping with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-10-25T13:41:05+08:00
draft: false
images: []
weight: 620
toc: true
aliases: []
support:
  level: community
  versions: true
  integration: true
seo:
  title: "ezBookkeeping | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring ezBookkeeping with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.13](https://github.com/authelia/authelia/releases/tag/v4.39.13)
- [ezBookkeeping]
  - [v1.2.0](https://github.com/mayswind/ezbookkeeping/releases/tag/v1.2.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://ezbookkeeping.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `ezbookkeeping`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [ezbookkeeping] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'ezbookkeeping'
        client_name: 'ezBookkeeping'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://ezbookkeeping.{{< sitevar name="domain" nojs="example.com" >}}/oauth2/callback'
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
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [ezBookkeeping] there are two methods, using the [Configuration File](#configuration-file), or using
[Environment Variables](#environment-variables).

#### Configuration File

To configure [ezBookkeeping] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```ini {title="ezbookkeeping.ini"}
[server]
domain = ezbookkeeping.{{< sitevar name="domain" nojs="example.com" >}}
root_url = https://ezbookkeeping.{{< sitevar name="domain" nojs="example.com" >}}/

[auth]
enable_oauth2_auth = true
oauth2_provider = oidc
oauth2_client_id = ezbookkeeping
oauth2_client_secret = insecure_secret
oauth2_use_pkce = true
oidc_provider_base_url = https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
enable_oidc_display_name = true
oidc_custom_display_name = Authelia
```

#### Environment Variables

To configure [ezBookkeeping] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

##### Standard

```shell {title=".env"}
EBK_SERVER_DOMAIN=ezbookkeeping.{{< sitevar name="domain" nojs="example.com" >}}
EBK_SERVER_ROOT_URL=https://ezbookkeeping.{{< sitevar name="domain" nojs="example.com" >}}/
EBK_AUTH_ENABLE_OAUTH2_AUTH=true
EBK_AUTH_OAUTH2_PROVIDER=oidc
EBK_AUTH_OAUTH2_CLIENT_ID=ezbookkeeping
EBK_AUTH_OAUTH2_CLIENT_SECRET=insecure_secret
EBK_AUTH_OAUTH2_USE_PKCE=true
EBK_AUTH_OIDC_PROVIDER_BASE_URL='https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
EBK_AUTH_ENABLE_OIDC_DISPLAY_NAME=true
EBK_AUTH_OIDC_CUSTOM_DISPLAY_NAME=Authelia
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  ezbookkeeping:
    environment:
      EBK_SERVER_DOMAIN=ezbookkeeping.{{< sitevar name="domain" nojs="example.com" >}}
      EBK_SERVER_ROOT_URL=https://ezbookkeeping.{{< sitevar name="domain" nojs="example.com" >}}/
      EBK_AUTH_ENABLE_OAUTH2_AUTH=true
      EBK_AUTH_OAUTH2_PROVIDER=oidc
      EBK_AUTH_OAUTH2_CLIENT_ID=ezbookkeeping
      EBK_AUTH_OAUTH2_CLIENT_SECRET=insecure_secret
      EBK_AUTH_OAUTH2_USE_PKCE=true
      EBK_AUTH_OIDC_PROVIDER_BASE_URL='https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      EBK_AUTH_ENABLE_OIDC_DISPLAY_NAME=true
      EBK_AUTH_OIDC_CUSTOM_DISPLAY_NAME=Authelia
```

## See Also

- [ezBookkeeping Configuration Documentation](https://ezbookkeeping.mayswind.net/configuration)

[Authelia]: https://www.authelia.com
[ezBookkeeping]: https://ezbookkeeping.mayswind.net
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
