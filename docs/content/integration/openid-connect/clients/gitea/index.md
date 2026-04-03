---
title: "Gitea"
description: "Integrating Gitea with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/gitea/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Gitea | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Gitea with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.13](https://github.com/authelia/authelia/releases/tag/v4.39.13)
- [Gitea]
  - [v1.25.1](https://github.com/go-gitea/gitea/releases/tag/v1.25.1)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://gitea.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `gitea`
- __Client Secret:__ `insecure_secret`
- __Authentication Name (Gitea):__ `authelia`:
    - This option determines the redirect URI in the format of
      `https://gitea.{{< sitevar name="domain" nojs="example.com" >}}/user/oauth2/<Authentication Name>/callback`.
      This means if you change this value you need to update the redirect URI.

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Gitea] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'gitea'
        client_name: 'Gitea'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://gitea.{{< sitevar name="domain" nojs="example.com" >}}/user/oauth2/authelia/callback'
        scopes:
          - 'openid'
          - 'email'
          - 'profile'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [Gitea] there are two methods, using the [Web GUI](#web-gui), or using the [CLI](#cli).

#### Web GUI

To configure [Gitea] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Expand User Options
2. Visit Site Administration
3. Visit Authentication Sources
4. Visit Add Authentication Source
5. Configure the following options:
   - Authentication Name: `authelia`
   - OAuth2 Provider: `OpenID Connect`
   - Client ID (Key): `gitea`
   - Client Secret: `insecure_secret`
   - OpenID Connect Auto Discovery URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration`

{{< figure src="gitea.png" alt="Gitea" width="300" >}}

#### CLI

_**Important Note:** Please refer to the [Gitea CLI Guide](https://docs.gitea.com/administration/command-line) regarding the correct usage of the CLI._

To configure [Gitea] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Run `gitea migrate`.
2. Run `gitea admin auth add-oauth --provider=openidConnect --name=authelia --key=gitea --secret=insecure_secret --auto-discover-url=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration --scopes='openid email profile'`


### Automatic User Creation

To configure [Gitea] to perform automatic user creation for the `{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}` domain via [OpenID Connect 1.0]:

1. Edit the following values in the [Gitea] `app.ini`:
```ini
[openid]
ENABLE_OPENID_SIGNIN = false
ENABLE_OPENID_SIGNUP = true
WHITELISTED_URIS     = {{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}

[service]
DISABLE_REGISTRATION                          = false
ALLOW_ONLY_EXTERNAL_REGISTRATION              = true
SHOW_REGISTRATION_BUTTON                      = false
```

## See Also

- [Gitea]
  - [Config Cheat Sheet](https://docs.gitea.io/en-us/config-cheat-sheet)
    - [OpenID](https://docs.gitea.io/en-us/config-cheat-sheet/#openid-openid)
    - [Service](https://docs.gitea.io/en-us/config-cheat-sheet/#service-service)

[Authelia]: https://www.authelia.com
[Gitea]: https://gitea.io/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
