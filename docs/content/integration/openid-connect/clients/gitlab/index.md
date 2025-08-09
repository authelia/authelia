---
title: "GitLab"
description: "Integrating GitLab with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/gitlab/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "GitLab | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring GitLab with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.38.0](https://github.com/authelia/authelia/releases/tag/v4.38.0)
- [GitLab] CE
  - [v16.9.0](https://about.gitlab.com/releases/2024/02/15/gitlab-16-9-released/)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://gitlab.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `gitlab`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
This configuration assumes you've configured the `client_auth_method` in [GitLab](https://about.gitlab.com/) as per below. If you
have not done this, the default in [GitLab](https://about.gitlab.com/) will require the `token_endpoint_auth_method` changes to
`client_secret_post`.
{{< /callout >}}

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
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://gitlab.{{< sitevar name="domain" nojs="example.com" >}}/users/auth/openid_connect/callback'
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

### Application

To configure [GitLab] there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

{{< callout context="tip" title="Did you know?" icon="outline/rocket" >}}
Generally the configuration file is named `gitlab.rb`.
{{< /callout >}}

To configure [GitLab] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```ruby {title="gitlab.rb"}
gitlab_rails['omniauth_providers'] = [
  {
    name: "openid_connect",
    label: "Authelia",
    icon: "https://www.authelia.com/images/branding/logo-cropped.png",
    args: {
      name: "openid_connect",
      strategy_class: "OmniAuth::Strategies::OpenIDConnect",
      issuer: "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}",
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
        redirect_uri: "https://gitlab.{{< sitevar name="domain" nojs="example.com" >}}/users/auth/openid_connect/callback"
      }
    }
  }
]
```

### Groups

[GitLab] offers group mapping options with OpenID Connect 1.0, shamefully it's only for paid plans. However see
[the guide](https://docs.gitlab.com/ee/administration/auth/oidc.html#configure-users-based-on-oidc-group-membership) on
how to configure it on their end.

Alternatively if GitLab is associated with LDAP you can use that as a group source, and you can configure a policy on
Authelia to restrict which resource owners are allowed access to the client for free via a custom `authorization_policy`
value.

## See Also

- [GitLab OpenID Connect OmniAuth Documentation](https://docs.gitlab.com/ee/administration/auth/oidc.html)

[Authelia]: https://www.authelia.com
[GitLab]: https://about.gitlab.com/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
