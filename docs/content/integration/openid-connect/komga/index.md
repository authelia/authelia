---
title: "Komga"
description: "Integrating Komga with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2022-08-26T11:39:00+10:00
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
* [Komga]
  * [v0.157.1](https://github.com/gotson/komga/releases/tag/v0.157.1)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://komga.example.com/`
* __Authelia Root URL:__ `https://auth.example.com/`
* __Client ID:__ `komga`
* __Client Secret:__ `insecure_secret`

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Komga] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'komga'
        client_name: 'Komga'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        redirect_uris:
          - 'https://komga.example.com/login/oauth2/code/authelia'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
        grant_types:
          - 'authorization_code'
        userinfo_signed_response_alg: 'none'
```

### Application

To configure [Komga] to utilize Authelia as an [OpenID Connect 1.0] Provider:

1. Configure the security section of the [Komga] configuration:
```yaml
komga:
  ## Comment if you don't want automatic account creation.
  oauth2-account-creation: true
spring:
  security:
    oauth2:
      client:
        registration:
          authelia:
            client-id: 'komga'
            client-secret: 'insecure_secret'
            client-name: 'Authelia'
            scope: 'openid,profile,email'
            authorization-grant-type: 'authorization_code'
            redirect-uri: "{baseScheme}://{baseHost}{basePort}{basePath}/login/oauth2/code/authelia"
        provider:
          authelia:
            issuer-uri: 'https://auth.example.com'
            user-name-attribute: 'preferred_username'
````

## See Also

* [Komga Configuration options Documentation](https://komga.org/installation/configuration.html)
* [Komga Social login Documentation](https://komga.org/installation/oauth2.html)

[Authelia]: https://www.authelia.com
[Komga]: https://www.komga.org
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
