---
title: "GitLab"
description: "Integrating GitLab with the Authelia OpenID Connect 1.0 Provider."
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
* [GitLab] CE
  * 16.9.0

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://gitlab.example.com/`
* __Authelia Root URL:__ `https://auth.example.com/`
* __Client ID:__ `gitlab`
* __Client Secret:__ `insecure_secret`

## Configuration

### Authelia

_**Important Note:** This configuration assumes you've configured the `client_auth_method` in [GitLab] as per below. If you
have not done this the default in [GitLab] will require the `token_endpoint_auth_method` changes to
`client_secret_post`._

The following YAML configuration is an example __Authelia__ [client configuration] for use with [GitLab] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'gitlab'
        client_name: 'GitLab'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        redirect_uris:
          - 'https://gitlab.example.com/users/auth/openid_connect/callback'
        scopes:
          - 'openid'
          - 'profile'
          - 'groups'
          - 'email'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

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
      strategy_class: "OmniAuth::Strategies::OpenIDConnect",
      issuer: "https://auth.example.com",
      discovery: true,
      scope: ["openid","profile","email","groups"],
      client_auth_method: "basic",
      response_type: "code",
      response_mode: "query",
      uid_field: "preferred_username",
      send_scope_to_token_endpoint: true,
      pkce: true,
      client_options: {
        identifier: "gitlab",
        secret: "insecure_secret",
        redirect_uri: "https://gitlab.example.com/users/auth/openid_connect/callback"
      }
    }
  }
]
```

#### Groups

[GitLab] offers group mapping options with OpenID Connect 1.0, shamefully it's only for paid plans. However see
[the guide](https://docs.gitlab.com/ee/administration/auth/oidc.html#configure-users-based-on-oidc-group-membership) on
how to configure it on their end.

Alternatively if GitLab is associated with LDAP you can use that as a group source, and you can configure a policy on
Authelia to restrict which resource owners are allowed access to the client for free via a custom `authorization_policy`
value.

## See Also

* [GitLab OpenID Connect OmniAuth Documentation](https://docs.gitlab.com/ee/administration/auth/oidc.html)

[Authelia]: https://www.authelia.com
[GitLab]: https://about.gitlab.com/
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
