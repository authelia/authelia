---
title: "freshrss"
description: "Integrating Freshrss with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-03-05T21:58:32+11:00
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
* [FreshRSS]
  * [1.23.1](https://github.com/FreshRSS/FreshRSS/releases/tag/1.23.1)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://freshrss.example.com/`
* __Authelia Root URL:__ `https://auth.example.com/`
* __Client ID:__ `freshrss`
* __Client Secret:__ `insecure_secret`
* __Port:__ '443'
  * This is the port [FreshRSS] is served over (usually 80 for http and 443 for https) NOT the port of the container.

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
        client_name: 'freshrss'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        redirect_uris:
          - 'https://freshrss.example.com:443/i/oidc/'
        scopes:
          - 'openid'
          - 'groups'
          - 'email'
          - 'profile'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

1. To configure [FreshRSS] to utilize Authelia as an [OpenID Connect 1.0](https://www.authelia.com/integration/openid-connect/introduction/) Provider, specify the below environment
   variables.
2. Open the newly created [FreshRSS] instance.
3. During initial config, select "HTTP" during the user creation

```yaml
environment:
  OIDC_ENABLED: 1
  OIDC_PROVIDER_METADATA_URL: https://auth.example.com/.well-known/openid-configuration
  OIDC_CLIENT_ID: freshrss
  OIDC_CLIENT_SECRET: insecure_secret
  OIDC_CLIENT_CRYPTO_KEY: XXXXXXXXXX
  OIDC_REMOTE_USER_CLAIM: preferred_username
  OIDC_SCOPES: openid groups email profile
  OIDC_X_FORWARDED_HEADERS: X-Forwarded-Host X-Forwarded-Port X-Forwarded-Proto
```

## See Also

- [freshrss OIDC documentation](https://freshrss.github.io/FreshRSS/en/admins/16_OpenID-Connect.html)

[Authelia]: https://www.authelia.com
[FreshRSS]: https://freshrss.github.io/FreshRSS/
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
