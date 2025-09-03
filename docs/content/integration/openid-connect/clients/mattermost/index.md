---
title: "Mattermost"
description: "Integrating Mattermost with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 620
toc: true
aliases: []
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Mattermost | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Mattermost with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.8](https://github.com/authelia/authelia/releases/tag/v4.39.8)
- [Mattermost]

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://mattermost.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `mattermost`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Mattermost] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    claims_policies:
      mattermost:
        custom_claims:
          username:
            name: 'username'
            attribute: 'username'
    scopes:
      mattermost:
        claims:
          - 'username'
    clients:
      - client_id: 'mattermost'
        client_name: 'Mattermost'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        claims_policy: 'mattermost'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://mattermost.{{< sitevar name="domain" nojs="example.com" >}}/signup/gitlab/complete'
        scopes:
          - 'openid'
          - 'mattermost'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

Before configuring or using [OpenID Connect 1.0] with [Mattermost] you must ensure the
[openid extension](https://mattermost.apache.org/doc/gug/openid-auth.html#installing-support-for-openid-connect) is
installed.

To configure [Mattermost]  there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

To configure [Mattermost] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```json {title="config.json"}
{
  "GitLabSettings": {
    "Enable": true,
    "Id": "mattermost",
    "Secret": "insecure_secret",
    "Scope": "openid mattermost",
    "DiscoveryEndpoint": "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration",
    "ButtonText": "Log in with Authelia",
    "ButtonColor": "#000000"
  }
}
```

## See Also

- [Mattermost GitLab SSO Documentation](https://docs.mattermost.com/administration-guide/onboard/sso-gitlab.html)
- [Mattermost OpenID Connect SSO Documentation](https://docs.mattermost.com/administration-guide/onboard/sso-openidconnect.html)

[Authelia]: https://www.authelia.com
[Mattermost]: https://mattermost.com/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
