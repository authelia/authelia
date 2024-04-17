---
title: "Grafana"
description: "Integrating Grafana with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2022-06-15T17:51:47+10:00
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
* [Grafana]
  * 8.0.0

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://grafana.example.com/`
* __Authelia Root URL:__ `https://auth.example.com/`
* __Client ID:__ `grafana`
* __Client Secret:__ `insecure_secret`

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
          - 'https://grafana.example.com/login/generic_oauth'
        scopes:
          - 'openid'
          - 'profile'
          - 'groups'
          - 'email'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [Grafana] to utilize Authelia as an [OpenID Connect 1.0] Provider, you have two effective options:

#### Configuration File

Add the following Generic OAuth configuration to the [Grafana] configuration:

```ini
[server]
root_url = https://grafana.example.com
[auth.generic_oauth]
enabled = true
name = Authelia
icon = signin
client_id = grafana
client_secret = insecure_secret
scopes = openid profile email groups
empty_scopes = false
auth_url = https://auth.example.com/api/oidc/authorization
token_url = https://auth.example.com/api/oidc/token
api_url = https://auth.example.com/api/oidc/userinfo
login_attribute_path = preferred_username
groups_attribute_path = groups
name_attribute_path = name
use_pkce = true
```

#### Environment Variables

Configure the following environment variables:

|                  Variable                   |                      Value                      |
|:-------------------------------------------:|:-----------------------------------------------:|
|             GF_SERVER_ROOT_URL              |           https://grafana.example.com           |
|        GF_AUTH_GENERIC_OAUTH_ENABLED        |                      true                       |
|         GF_AUTH_GENERIC_OAUTH_NAME          |                    Authelia                     |
|       GF_AUTH_GENERIC_OAUTH_CLIENT_ID       |                     grafana                     |
|     GF_AUTH_GENERIC_OAUTH_CLIENT_SECRET     |                 insecure_secret                 |
|        GF_AUTH_GENERIC_OAUTH_SCOPES         |           openid profile email groups           |
|     GF_AUTH_GENERIC_OAUTH_EMPTY_SCOPES      |                      false                      |
|       GF_AUTH_GENERIC_OAUTH_AUTH_URL        | https://auth.example.com/api/oidc/authorization |
|       GF_AUTH_GENERIC_OAUTH_TOKEN_URL       |     https://auth.example.com/api/oidc/token     |
|        GF_AUTH_GENERIC_OAUTH_API_URL        |   https://auth.example.com/api/oidc/userinfo    |
| GF_AUTH_GENERIC_OAUTH_LOGIN_ATTRIBUTE_PATH  |               preferred_username                |
| GF_AUTH_GENERIC_OAUTH_GROUPS_ATTRIBUTE_PATH |                     groups                      |
|  GF_AUTH_GENERIC_OAUTH_NAME_ATTRIBUTE_PATH  |                      name                       |
|       GF_AUTH_GENERIC_OAUTH_USE_PKCE        |                      true                       |
|  GF_AUTH_GENERIC_OAUTH_ROLE_ATTRIBUTE_PATH  |            See [Role Attribute Path]            |

[Role Attribute Path]: #role-attribute-path

#### Role Attribute Path

The role attribute path configuration is optional but allows mapping Authelia group membership with Grafana roles. If
you do not wish to automatically do this you can just omit the environment variable.

The ways you can configure this rule value is vast as an examle if you wanted a default role of `Viewer`, but also
wanted everyone in the `admin` Authelia group to be in the `Admin` role, and everyone in the `editor` Authelia group to
be in the `Editor` role, a rule similar to
`contains(groups, 'admin') && 'Admin' || contains(groups, 'editor') && 'Editor' || 'Viewer'` would be needed.

See [Grafana Generic OAuth2 Documentation: Configure role mapping] for more information.

## See Also

* [Grafana OAuth Documentation](https://grafana.com/docs/grafana/latest/auth/generic-oauth/)
* [Grafana Generic OAuth2 Documentation: Configure role mapping]

[Authelia]: https://www.authelia.com
[Grafana]: https://grafana.com/
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[Grafana Generic OAuth2 Documentation: Configure role mapping]: https://grafana.com/docs/grafana/latest/setup-grafana/configure-security/configure-authentication/generic-oauth/#configure-role-mapping
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
