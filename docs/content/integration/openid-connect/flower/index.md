---
title: "Flower"
description: "Integrating Flower with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-08-20T21:53:14+10:00
draft: false
images: []
weight: 620
toc: true
support:
  level: community
  versions: true
  integration: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

* [Authelia]
  * [v4.38.8](https://github.com/authelia/authelia/releases/tag/v4.38.8)
* [Flower]
  * [2.0.1 3b9ed2f](https://github.com/m4rc3l-h3/flower)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://flower.{{< sitevar name="domain" nojs="example.com" >}}/`
  * This option determines the redirect URI in the format of
        `https://flower.{{< sitevar name="domain" nojs="example.com" >}}/login`.
        This means if you change this value, you need to update the redirect URI.
* __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
* __Client ID:__ `flower`
* __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Flower] which will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'flower'
        client_name: 'Flower'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        redirect_uris:
          - 'https://flower.{{< sitevar name="domain" nojs="example.com" >}}/login'
        scopes:
          - 'openid'
          - 'email'
          - 'profile'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_post'
```

### Application


To configure [Flower] to utilize Authelia as an [OpenID Connect 1.0] Provider you have to update the `flowerconfig.py` configuration file and configure the `FLOWER_OAUTH2_AUTHELIA_BASE_URL` environment variable.

#### Configuration File

Add the following values to [Flower] `flowerconfig.py`:
```python
auth = '.*@{{< sitevar name="domain" nojs="example.com" >}}'
auth_provider = 'flower.views.auth.AutheliaLoginHandler'
oauth2_key = 'flower'
oauth2_secret = 'insecure_secret'
oauth2_redirect_uri = 'https://flower.{{< sitevar name="domain" nojs="example.com" >}}/login'
```

#### Environment Variables

Add the `FLOWER_OAUTH2_AUTHELIA_BASE_URL` environment variable and set it to Authelia Root URL:
``` bash
export FLOWER_OAUTH2_AUTHELIA_BASE_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
```

Take a look at the [See Also](#see-also) section for the cheatsheets corresponding to the sections above for their descriptions.

## See Also

- [Flower]
  - [Authentication](https://github.com/m4rc3l-h3/flower/blob/master/docs/auth.rst#authentication)
  - [Configuration](https://github.com/m4rc3l-h3/flower/blob/master/docs/config.rst#configuration)

[Authelia]: https://www.authelia.com
[Flower]: https://github.com/m4rc3l-h3/flower/blob/master/docs/auth.rst#authelia-oauth
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
