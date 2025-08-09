---
title: "Warpgate"
description: "Integrating Warpgate with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-03-23T23:08:06+11:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/warpgate/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Warpgate | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Warpgate with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.38.0](https://github.com/authelia/authelia/releases/tag/v4.38.0)
- [Warpgate]
  - [v0.9.1](https://github.com/warp-tech/warpgate/releases/tag/v0.9.1)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://warpgate.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `warpgate`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Warpgate]
which will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'warpgate'
        client_name: 'Warpgate'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://warpgate.{{< sitevar name="domain" nojs="example.com" >}}/@warpgate/api/sso/return'
        scopes:
          - 'openid'
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

To configure [Warpgate] there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

{{< callout context="tip" title="Did you know?" icon="outline/rocket" >}}
Generally the configuration file is named `warpgate.yaml`.
{{< /callout >}}

To configure [Warpgate] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```yaml {title="warpgate.yaml"}
external_host: warpgate.{{< sitevar name="domain" nojs="example.com" >}}
sso_providers:
- name: authelia
  label: Authelia
  provider:
    type: custom
    client_id: warpgate
    client_secret: insecure_secret
    issuer_url: https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
    scopes: ["openid", "email"]
```

## See Also

- [Warpgate OpenID Connect Documentation](https://github.com/warp-tech/warpgate/wiki/SSO-Authentication)

[Authelia]: https://www.authelia.com
[Warpgate]: https://github.com/warp-tech/warpgate
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
