---
title: "Apache Guacamole"
description: "Integrating Apache Guacamole with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/apache-guacamole/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Apache Guacamole | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Apache Guacamole with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.10](https://github.com/authelia/authelia/releases/tag/v4.39.10)
- [Apache Guacamole]
  - [v1.5.5](https://guacamole.apache.org/releases/1.5.5/)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://guacamole.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `guacamole`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Apache Guacamole] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'guacamole'
        client_name: 'Apache Guacamole'
        public: true
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://guacamole.{{< sitevar name="domain" nojs="example.com" >}}'
        scopes:
          - 'openid'
          - 'profile'
          - 'groups'
          - 'email'
        response_types:
          - 'id_token'
        grant_types:
          - 'implicit'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

Before configuring or using [OpenID Connect 1.0] with [Apache Guacamole] you must ensure the
[openid extension](https://guacamole.apache.org/doc/gug/openid-auth.html#installing-support-for-openid-connect) is
installed.

To configure [Apache Guacamole]  there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

To configure [Apache Guacamole] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```yaml
openid-client-id: guacamole
openid-scope: openid profile groups email
openid-issuer: https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
openid-jwks-endpoint: https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/jwks.json
openid-authorization-endpoint: https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization?state=1234abcedfdhf
openid-redirect-uri: https://guacamole.{{< sitevar name="domain" nojs="example.com" >}}
openid-username-claim-type: preferred_username
openid-groups-claim-type: groups
```

## See Also

- [Apache Guacamole OpenID Connect Authentication Documentation](https://guacamole.apache.org/doc/gug/openid-auth.html)

[Authelia]: https://www.authelia.com
[Apache Guacamole]: https://guacamole.apache.org/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
