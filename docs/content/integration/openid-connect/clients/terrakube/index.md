---
title: "Terrakube"
description: "Integrating Terrakube with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-01-25T10:04:53+11:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/terrakube/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Terrakube | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Terrakube with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.38.18](https://github.com/authelia/authelia/releases/tag/v4.38.18)
- [Terrakube]
  - [v2.24.1](https://github.com/AzBuilder/terrakube/releases/tag/2.24.1)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://terrakube.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
        `https://terrakube.{{< sitevar name="domain" nojs="example.com" >}}/login`.
        This means if you change this value, you need to update the redirect URI.
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `terrakube`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Terrakube] which will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'terrakube'
        client_name: 'Terrakube'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://terrakube.{{< sitevar name="domain" nojs="example.com" >}}/dex/callback'
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

To configure [Terrakube] there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

{{< callout context="tip" title="Did you know?" icon="outline/rocket" >}}
The configuration file below is the Dex configuration for Terrakube.
{{< /callout >}}

To configure [Terrakube] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```yaml
connectors:
  - type: oidc
    id: TerrakubeClient
    name: TerrakubeClient
    config:
      issuer: "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}"
      clientID: "terrakube"
      clientSecret: "insecure_secret"
      redirectURI: "https://terrakube.{{< sitevar name="domain" nojs="example.com" >}}/dex/callback"
      insecureEnableGroups: true
```

## See Also

- [Terrakube Documentation](https://docs.terrakube.io/)

[Authelia]: https://www.authelia.com
[Terrakube]: https://terrakube.io/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
