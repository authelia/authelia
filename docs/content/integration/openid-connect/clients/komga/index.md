---
title: "Komga"
description: "Integrating Komga with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/komga/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Komga | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Komga with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.38.0](https://github.com/authelia/authelia/releases/tag/v4.38.0)
- [Komga]
  - [v0.157.1](https://github.com/gotson/komga/releases/tag/v0.157.1)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://komga.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `komga`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

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
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://komga.{{< sitevar name="domain" nojs="example.com" >}}/login/oauth2/code/authelia'
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

To configure [Komga] there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

{{< callout context="tip" title="Did you know?" icon="outline/rocket" >}}
Generally the configuration file is named `application.yml`.
{{< /callout >}}

To configure [Komga] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```yaml {title="application.yml"}
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
            issuer-uri: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
            user-name-attribute: 'preferred_username'
````

## See Also

- [Komga Configuration options Documentation](https://komga.org/docs/installation/configuration.html)
- [Komga Social login Documentation](https://komga.org/docs/installation/oauth2/)

[Authelia]: https://www.authelia.com
[Komga]: https://www.komga.org
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
