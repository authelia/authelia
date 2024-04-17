---
title: "Paperless"
description: "Integrating Paperless with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-04-13T13:46:05+10:00
draft: false
images: []
weight: 620
toc: true
community: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

* [Authelia]
  * [v4.38.0](https://github.com/authelia/authelia/releases/tag/v4.38.0)
* [Paperless]
  * [v2.7.2](https://github.com/paperless-ngx/paperless-ngx/releases/tag/v2.7.2)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://paperless.example.com/`
* __Authelia Root URL:__ `https://auth.example.com/`
* __Client ID:__ `paperless`
* __Client Secret:__ `insecure_secret`

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
          - 'https://paperless.example.com/accounts/authelia/login/callback'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
          - 'groups'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [Paperless] to utilize Authelia as an [OpenID Connect 1.0] Provider:

1. Set the following environment variables:

```env
PAPERLESS_APPS=allauth.socialaccount.providers.openid_connect
PAPERLESS_SOCIALACCOUNT_PROVIDERS={"openid_connect":{"SCOPE":["openid","profile","email"],"OAUTH_PKCE_ENABLED":true,"APPS":[{"provider_id":"authelia","name":"Authelia","client_id":"paperless","secret":"insecure_secret","settings":{"server_url":"https://auth.example.com","token_auth_method":"client_secret_basic"}}]}}
```

The `PAPERLESS_SOCIALACCOUNT_PROVIDERS` environment variable is the minified version of the following:

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
          "server_url": "https://auth.example.com",
          "token_auth_method": "client_secret_basic"
        }
      }
    ]
  }
}
```

## See Also

- [Paperless Advanced Usage OpenID Connect Documentation](https://docs.paperless-ngx.com/advanced_usage/#openid-connect-and-social-authentication)

[Paperless]: https://docs.paperless-ngx.com/
[Authelia]: https://www.authelia.com
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
