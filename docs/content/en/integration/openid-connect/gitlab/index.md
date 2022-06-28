---
title: "GitLab"
description: "Integrating GitLab with Authelia via OpenID Connect."
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
* [GitLab] CE
  * 14.0.1

## Before You Begin

You are required to utilize a unique client id and a unique and random client secret for all [OpenID Connect] relying
parties. You should not use the client secret in this example, you should randomly generate one yourself. You may also
choose to utilize a different client id, it's completely up to you.

This example makes the following assumptions:

* __Application Root URL:__ `https://gitlab.example.com`
* __Authelia Root URL:__ `https://auth.example.com`
* __Client ID:__ `gitlab`
* __Client Secret:__ `gitlab_client_secret`

## Configuration

### Application

To configure [GitLab] to utilize Authelia as an [OpenID Connect] Provider:

1. Add the Omnibus [OpenID Connect] OmniAuth configuration to `gitlab.rb`:

```ruby
gitlab_rails['omniauth_providers'] = [
  {
    name: "openid_connect",
    label: "Authelia",
    icon: "https://avatars.githubusercontent.com/u/59122411?s=280&v=4",
    args: {
      name: "openid_connect",
      scope: ["openid","profile","email","groups"],
      response_type: "code",
      issuer: "https://auth.example.com",
      discovery: true,
      client_auth_method: "query",
      uid_field: "preferred_username",
      send_scope_to_token_endpoint: "false",
      client_options: {
        identifier: "gitlab",
        secret: "gitlab_client_secret",
        redirect_uri: "https://gitlab.example.com/users/auth/openid_connect/callback"
      }
    }
  }
]
```

### Authelia

The following YAML configuration is an example __Authelia__
[client configuration](../../../configuration/identity-providers/open-id-connect.md#clients) for use with [GitLab]
which will operate with the above example:

```yaml
- id: gitlab
  secret: gitlab_client_secret
  public: false
  authorization_policy: two_factor
  scopes:
    - openid
    - profile
    - groups
    - email
  redirect_uris:
    - https://gitlab.example.com/users/auth/openid_connect/callback
  userinfo_signing_algorithm: none
```

## See Also

* [GitLab OpenID Connect OmniAuth Documentation](https://docs.gitlab.com/ee/administration/auth/oidc.html)

[Authelia]: https://www.authelia.com
[GitLab]: https://about.gitlab.com/
[OpenID Connect]: ../../openid-connect/introduction.md
