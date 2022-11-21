---
title: "Grafana"
description: "Integrating Grafana with the Authelia OpenID Connect Provider."
lead: ""
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  integration:
    parent: "openid-connect"
weight: 620
toc: true
community: true
---

## Tested Versions

* [Authelia]
  * [v4.35.5](https://github.com/authelia/authelia/releases/tag/v4.35.5)
* [Grafana]
  * 8.0.0

## Before You Begin

### Common Notes

1. You are *__required__* to utilize a unique client id for every client.
2. The client id on this page is merely an example and you can theoretically use any alphanumeric string.
3. You *__should not__* use the client secret in this example, We *__strongly recommend__* reading the
   [Generating Client Secrets] guide instead.

[Generating Client Secrets]: ../specific-information.md#generating-client-secrets

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://grafana.example.com`
* __Authelia Root URL:__ `https://auth.example.com`
* __Client ID:__ `grafana`
* __Client Secret:__ `grafana_client_secret`

## Configuration

### Application

To configure [Grafana] to utilize Authelia as an [OpenID Connect] Provider you have two effective options:

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
client_secret = grafana_client_secret
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
|     GF_AUTH_GENERIC_OAUTH_CLIENT_SECRET     |              grafana_client_secret              |
|        GF_AUTH_GENERIC_OAUTH_SCOPES         |           openid profile email groups           |
|     GF_AUTH_GENERIC_OAUTH_EMPTY_SCOPES      |                      false                      |
|       GF_AUTH_GENERIC_OAUTH_AUTH_URL        | https://auth.example.com/api/oidc/authorization |
|       GF_AUTH_GENERIC_OAUTH_TOKEN_URL       |     https://auth.example.com/api/oidc/token     |
|        GF_AUTH_GENERIC_OAUTH_API_URL        |   https://auth.example.com/api/oidc/userinfo    |
| GF_AUTH_GENERIC_OAUTH_LOGIN_ATTRIBUTE_PATH  |               preferred_username                |
| GF_AUTH_GENERIC_OAUTH_GROUPS_ATTRIBUTE_PATH |                     groups                      |
|  GF_AUTH_GENERIC_OAUTH_NAME_ATTRIBUTE_PATH  |                      name                       |
|       GF_AUTH_GENERIC_OAUTH_USE_PKCE        |                      true                       |

### Authelia

The following YAML configuration is an example __Authelia__
[client configuration](../../../configuration/identity-providers/open-id-connect.md#clients) for use with [Grafana]
which will operate with the above example:

```yaml
- id: grafana
  description: Grafana
  secret: '$plaintext$grafana_client_secret'
  public: false
  authorization_policy: two_factor
  redirect_uris:
    - https://grafana.example.com/login/generic_oauth
  scopes:
    - openid
    - profile
    - groups
    - email
  userinfo_signing_algorithm: none
```

## See Also

* [Grafana OAuth Documentation](https://grafana.com/docs/grafana/latest/auth/generic-oauth/)

[Authelia]: https://www.authelia.com
[Grafana]: https://grafana.com/
[OpenID Connect]: ../../openid-connect/introduction.md
