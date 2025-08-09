---
title: "Paperless"
description: "Integrating Paperless with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-04-13T13:46:05+10:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/paperless/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Paperless | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Paperless with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.38.17](https://github.com/authelia/authelia/releases/tag/v4.38.17)
- [Paperless]
  - [v2.13.5](https://github.com/paperless-ngx/paperless-ngx/releases/tag/v2.13.5)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://paperless.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `paperless`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Paperless] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'paperless'
        client_name: 'Paperless'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://paperless.{{< sitevar name="domain" nojs="example.com" >}}/accounts/oidc/authelia/login/callback/'
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

To configure [Paperless] there is one method, using the [Environment Variables](#environment-variables).

#### Environment Variables

For reference purposes, the below `PAPERLESS_SOCIALACCOUNT_PROVIDERS` environment variable examples are the minified
format of the following:

```json
{
  "openid_connect": {
    "SCOPE": ["openid", "profile", "email"],
    "OAUTH_PKCE_ENABLED": true,
    "APPS": [
      {
        "provider_id": "authelia",
        "name": "Authelia",
        "client_id": "paperless",
        "secret": "insecure_secret",
        "settings": {
          "server_url": "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}",
          "token_auth_method": "client_secret_basic"
        }
      }
    ]
  }
}
```

To configure [Paperless] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment variables:

##### Standard

```shell {title=".env"}
PAPERLESS_APPS=allauth.socialaccount.providers.openid_connect
PAPERLESS_SOCIALACCOUNT_PROVIDERS={"openid_connect":{"SCOPE":["openid","profile","email"],"OAUTH_PKCE_ENABLED":true,"APPS":[{"provider_id":"authelia","name":"Authelia","client_id":"paperless","secret":"insecure_secret","settings":{"server_url":"https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}","token_auth_method":"client_secret_basic"}}]}}
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  paperless:
    environment:
      PAPERLESS_APPS: 'allauth.socialaccount.providers.openid_connect'
      PAPERLESS_SOCIALACCOUNT_PROVIDERS: '{"openid_connect":{"SCOPE":["openid","profile","email"],"OAUTH_PKCE_ENABLED":true,"APPS":[{"provider_id":"authelia","name":"Authelia","client_id":"paperless","secret":"insecure_secret","settings":{"server_url":"https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}","token_auth_method":"client_secret_basic"}}]}}'
```

## See Also

- [Paperless Advanced Usage OpenID Connect Documentation](https://docs.paperless-ngx.com/advanced_usage/#openid-connect-and-social-authentication)

[Paperless]: https://docs.paperless-ngx.com/
[Authelia]: https://www.authelia.com
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
