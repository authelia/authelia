---
title: "Vikunja"
description: "Integrating Vikunja with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-04-13T13:46:05+10:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/vikunja/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Vikunja | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Vikunja with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.38.0](https://github.com/authelia/authelia/releases/tag/v4.38.0)
- [Vikunja]
  - [v0.23.0](https://kolaente.dev/vikunja/vikunja/releases/tag/v0.23.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://vikunja.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `vikunja`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Vikunja] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'vikunja'
        client_name: 'Vikunja'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://vikunja.{{< sitevar name="domain" nojs="example.com" >}}/auth/openid/authelia'
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

To configure [Vikunja] there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

{{< callout context="tip" title="Did you know?" icon="outline/rocket" >}}
Generally the configuration file is named `config.yml`.
{{< /callout >}}

To configure [Vikunja] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```yaml {title="config.yml"}
auth:
  openid:
    enabled: true
    providers:
      authelia:
        name: 'Authelia'
        authurl: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
        clientid: 'vikunja'
        clientsecret: 'insecure_secret'
        scope: 'openid profile email'
```

## See Also

- [Vikunja OpenID Documentation](https://vikunja.io/docs/openid/)

[Vikunja]: https://vikunja.io/
[Authelia]: https://www.authelia.com
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
