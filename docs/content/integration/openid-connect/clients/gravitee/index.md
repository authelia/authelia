---
title: "Gravitee"
description: "Integrating Gravitee with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-04-26T11:03:16+00:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/gravitee/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Gravitee | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Gravitee with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.6](https://github.com/authelia/authelia/releases/tag/v4.39.6)
- [Gravitee]
  - [v4.7](https://documentation.gravitee.io/apim/release-information/release-notes/apim-4.7)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://gravitee.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `gravitee`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Gravitee] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'gravitee'
        client_name: 'Gravitee'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://gravitee.{{< sitevar name="domain" nojs="example.com" >}}/'
        scopes:
          - 'openid'
          - 'email'
          - 'profile'
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

To configure [Gravitee] there is one method, using the [Web GUI](#web-gui).

#### Configuration File

{{< callout context="tip" title="Did you know?" icon="outline/rocket" >}}
Generally the configuration file is named `gravitee.yaml`.
{{< /callout >}}

To configure [Gravitee] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```yaml {title="gravitee.yaml"}
security:
  providers:
    - type: 'oidc'
      clientId: 'gravitee'
      clientSecret: 'insecure_secret'
      tokenIntrospectionEndpoint: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/introspection'
      tokenEndpoint: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token'
      authorizeEndpoint: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization'
      userInfoEndpoint: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo'
      syncMappings: true
      scopes:
        - 'openid'
        - 'email'
        - 'profile'
        - 'groups'
      userMapping:
        id: 'sub'
        email: 'email'
        lastname: 'family_name'
        firstname: 'given_name'
        picture: 'photo'
      roleMapping:
        - condition: "{(#jsonPath(#profile, '$.groups') matches 'gravitee-admin' )}"
          roles:
            - "ORGANIZATION:ADMIN"
            - "ENVIRONMENT:ADMIN"
```

#### Web GUI

To configure [Gravitee] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [Gravitee].
2. Navigate to Settings.
3. Navigate to OIDC.
4. Click `+ New Auth Provider`.
5. Configure the following options:
   - Name: `Authelia`
   - Client ID: `gravitee`
   - Client Secret: `insecure_secret`
   - Issuer: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
   - Authorization Endpoint: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization`
   - Token Endpoint: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token`
   - Userinfo Endpoint: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo`
   - JWKS Endpoint: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/jwks.json`
6. Press `Submit` at the bottom.

## See Also

- [Gravitee OpenID Connect Documentation](https://documentation.gravitee.io/apim/administration/authentication/openid-connect)
- [Gravitee Roles and Groups Mapping](https://documentation.gravitee.io/apim/administration/authentication/roles-and-groups-mapping)

[Authelia]: https://www.authelia.com
[Gravitee]: https://www.gravitee.io/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
