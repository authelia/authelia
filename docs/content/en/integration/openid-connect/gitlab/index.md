---
title: "GitLab"
description: "Integrating GitLab with the Authelia OpenID Connect Provider."
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

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://gitlab.example.com`
* __Authelia Root URL:__ `https://auth.example.com`
* __Client ID:__ `gitlab`
* __Client Secret:__ `insecure_secret`

## Configuration

### Application

To configure [GitLab] to utilize Authelia as an [OpenID Connect 1.0] Provider:

1. Add the Omnibus [OpenID Connect 1.0] OmniAuth configuration to `gitlab.rb`:

```ruby
gitlab_rails['omniauth_providers'] = [
  {
    name: "openid_connect",
    label: "Authelia",
    icon: "https://www.authelia.com/images/branding/logo-cropped.png",
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
        secret: "insecure_secret",
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
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
    - id: 'gitlab'
      description: 'GitLab'
      secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
      public: false
      authorization_policy: 'two_factor'
      redirect_uris:
        - 'https://gitlab.example.com/users/auth/openid_connect/callback'
      scopes:
        - 'openid'
        - 'profile'
        - 'groups'
        - 'email'
      userinfo_signing_algorithm: 'none'
```

## See Also

* [GitLab OpenID Connect OmniAuth Documentation](https://docs.gitlab.com/ee/administration/auth/oidc.html)

[Authelia]: https://www.authelia.com
[GitLab]: https://about.gitlab.com/
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
