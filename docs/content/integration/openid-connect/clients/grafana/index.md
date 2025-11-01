---
title: "Grafana"
description: "Integrating Grafana with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/grafana/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Grafana | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Grafana with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.14](https://github.com/authelia/authelia/releases/tag/v4.39.14)
- [Grafana]
  - [v12.0.2](https://github.com/grafana/grafana/releases/tag/v12.0.2)

{{% oidc-common bugs="claims-hydration" %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://grafana.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `grafana`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Grafana] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'grafana'
        client_name: 'Grafana'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://grafana.{{< sitevar name="domain" nojs="example.com" >}}/login/generic_oauth'
        scopes:
          - 'openid'
          - 'profile'
          - 'groups'
          - 'email'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

#### Configuration Escape Hatch

{{% oidc-escape-hatch-claims-hydration client_id="grafana" claims="email,name,groups,preferred_username" %}}

### Application

To configure [Grafana] there are two methods, using the [Configuration File](#configuration-file), or using
[Environment Variables](#environment-variables).

#### Configuration File

{{< callout context="tip" title="Did you know?" icon="outline/rocket" >}}
Generally the configuration file is named `grafana.ini`.
{{< /callout >}}

To configure [Grafana] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```ini {title="grafana.ini"}
[server]
root_url = https://grafana.{{< sitevar name="domain" nojs="example.com" >}}

[auth.generic_oauth]
enabled = true
name = Authelia
icon = signin
client_id = grafana
client_secret = insecure_secret
scopes = openid profile email groups
empty_scopes = false
auth_url = https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization
token_url = https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token
api_url = https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo
login_attribute_path = preferred_username
groups_attribute_path = groups
name_attribute_path = name
use_pkce = true
role_attribute_path =
```

#### Environment Variables

To configure [Grafana] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

##### Standard

```shell {title=".env"}
GF_SERVER_ROOT_URL=https://grafana.{{< sitevar name="domain" nojs="example.com" >}}
GF_AUTH_GENERIC_OAUTH_ENABLED=true
GF_AUTH_GENERIC_OAUTH_NAME=Authelia
GF_AUTH_GENERIC_OAUTH_ICON=signin
GF_AUTH_GENERIC_OAUTH_CLIENT_ID=grafana
GF_AUTH_GENERIC_OAUTH_CLIENT_SECRET=insecure_secret
GF_AUTH_GENERIC_OAUTH_SCOPES=openid profile email groups
GF_AUTH_GENERIC_OAUTH_EMPTY_SCOPES=false
GF_AUTH_GENERIC_OAUTH_AUTH_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization
GF_AUTH_GENERIC_OAUTH_TOKEN_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token
GF_AUTH_GENERIC_OAUTH_API_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo
GF_AUTH_GENERIC_OAUTH_LOGIN_ATTRIBUTE_PATH=preferred_username
GF_AUTH_GENERIC_OAUTH_GROUPS_ATTRIBUTE_PATH=groups
GF_AUTH_GENERIC_OAUTH_NAME_ATTRIBUTE_PATH=name
GF_AUTH_GENERIC_OAUTH_USE_PKCE=true
GF_AUTH_GENERIC_OAUTH_ROLE_ATTRIBUTE_PATH=
```

##### Docker Compose

```yaml {title="compose.yml"}
services:
  grafana:
    environment:
      GF_SERVER_ROOT_URL: 'https://grafana.{{< sitevar name="domain" nojs="example.com" >}}'
      GF_AUTH_GENERIC_OAUTH_ENABLED: 'true'
      GF_AUTH_GENERIC_OAUTH_NAME: 'Authelia'
      GF_AUTH_GENERIC_OAUTH_ICON: 'signin'
      GF_AUTH_GENERIC_OAUTH_CLIENT_ID: 'grafana'
      GF_AUTH_GENERIC_OAUTH_CLIENT_SECRET: 'insecure_secret'
      GF_AUTH_GENERIC_OAUTH_SCOPES: 'openid profile email groups'
      GF_AUTH_GENERIC_OAUTH_EMPTY_SCOPES: 'false'
      GF_AUTH_GENERIC_OAUTH_AUTH_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization'
      GF_AUTH_GENERIC_OAUTH_TOKEN_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token'
      GF_AUTH_GENERIC_OAUTH_API_URL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo'
      GF_AUTH_GENERIC_OAUTH_LOGIN_ATTRIBUTE_PATH: 'preferred_username'
      GF_AUTH_GENERIC_OAUTH_GROUPS_ATTRIBUTE_PATH: 'groups'
      GF_AUTH_GENERIC_OAUTH_NAME_ATTRIBUTE_PATH: 'name'
      GF_AUTH_GENERIC_OAUTH_USE_PKCE: 'true'
      GF_AUTH_GENERIC_OAUTH_ROLE_ATTRIBUTE_PATH: ''
```

### Role Attribute Path

The role attribute path configuration is optional but allows mapping Authelia group membership with Grafana roles. If
you do not wish to automatically do this you can just omit the `role_attribute_path` configuration option or
`GF_AUTH_GENERIC_OAUTH_ROLE_ATTRIBUTE_PATH` environment variable.

The ways you can configure this rule value is vast, here is a simple example:
- User's with the authelia group `admin` should be a member of the Grafana group `Admin`
- User's with the authelia group `editor` should be a member of the Grafana group `Editor`
- Everyone else should be a member of the Grafana group 'Viewer'

To achieve the above structure you would use the following `role_attribute_path`:
`contains(groups[*], 'admin') && 'Admin' || contains(groups[*], 'editor') && 'Editor' || 'Viewer'`

See [Grafana Generic OAuth2 Documentation: Configure role mapping] for more information.

## See Also

- [Grafana OAuth Documentation](https://grafana.com/docs/grafana/latest/auth/generic-oauth/)
- [Grafana Generic OAuth2 Documentation: Configure role mapping]

[Authelia]: https://www.authelia.com
[Grafana]: https://grafana.com/
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[Grafana Generic OAuth2 Documentation: Configure role mapping]: https://grafana.com/docs/grafana/latest/setup-grafana/configure-security/configure-authentication/generic-oauth/#configure-role-mapping
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
