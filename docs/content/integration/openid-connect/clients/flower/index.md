---
title: "Flower"
description: "Integrating Flower with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-08-20T21:53:14+10:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/flower/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Flower | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Flower with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.38.8](https://github.com/authelia/authelia/releases/tag/v4.38.8)
- [Flower]

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://flower.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
        `https://flower.{{< sitevar name="domain" nojs="example.com" >}}/login`.
        This means if you change this value, you need to update the redirect URI.
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `flower`
- __Client Secret:__ `insecure_secret`

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
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://flower.{{< sitevar name="domain" nojs="example.com" >}}/login'
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

To configure [Flower] there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

{{< callout context="tip" title="Did you know?" icon="outline/rocket" >}}
Generally the configuration file is named `flowerconfig.py`.
{{< /callout >}}

To configure [Flower] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```python {title="flowerconfig.py"}
auth = '.*@{{< sitevar name="domain" nojs="example.com" >}}'
auth_provider = 'flower.views.auth.AutheliaLoginHandler'
oauth2_key = 'flower'
oauth2_secret = 'insecure_secret'
oauth2_redirect_uri = 'https://flower.{{< sitevar name="domain" nojs="example.com" >}}/login'
```

In addition to the configuration change you must also set the following environment variables:

##### Standard

```shell {title=".env"}
FLOWER_OAUTH2_AUTHELIA_BASE_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  expressjs-example:
    environment:
      FLOWER_OAUTH2_AUTHELIA_BASE_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
```

## See Also

- [Flower]
- [Authentication](https://github.com/m4rc3l-h3/flower/blob/master/docs/auth.rst#authentication)
- [Configuration](https://github.com/m4rc3l-h3/flower/blob/master/docs/config.rst#configuration)

[Authelia]: https://www.authelia.com
[Flower]: https://github.com/m4rc3l-h3/flower/blob/master/docs/auth.rst#authelia-oauth
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
