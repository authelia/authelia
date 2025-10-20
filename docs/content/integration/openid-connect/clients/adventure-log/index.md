---
title: "AdventureLog"
description: "Integrating AdventureLog with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-04-26T18:35:57+10:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/adventure-log/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "AdventureLog | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring AdventureLog with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.13](https://github.com/authelia/authelia/releases/tag/v4.39.13)
- [AdventureLog]
  - [v0.9.0](https://github.com/seanmorley15/AdventureLog/releases/tag/v0.9.0)

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
A bug with Adventure Log requires manual adjustments to this guide and those adjustments are noted. See
[#544](https://github.com/seanmorley15/AdventureLog/issues/544) for more detail.
{{< /callout >}}

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://adventurelog.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `adventurelog`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [AdventureLog] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'adventurelog'
        client_name: 'Adventure Log'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://adventurelog.{{< sitevar name="domain" nojs="example.com" >}}/accounts/oidc/authelia/login/callback/'
          - 'https://adventurelog.{{< sitevar name="domain" nojs="example.com" >}}/accounts/oidc/adventurelog/login/callback/'  # Note: this is the workaround redirect_uri.
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

To configure [AdventureLog] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [AdventureLog] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [AdventureLog].
2. Navigate to the Admin Panel:
   1. Clicking your profile picture.
   2. Select Settings.
   3. Click Launch Admin Panel.
3. Scroll down to Social Accounts.
4. Under Social Applications, click Add.
5. Configure the following options:
   - Provider: `OpenID Connect`
   - Provider ID: `authelia`  (_**Note**: this will need to be `adventurelog` until the bug is fixed_)
   - Name: `Authelia`
   - Client ID: `adventurelog`
   - Secret Key: `insecure_secret`
   - Settings:
     `{"server_url": "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}"}`
   - Sites: Select the sites you want to enable OpenID Connect for.
      (By default, you should add the pre-created `example.com` site.)
6. Press `Save` at the bottom.


## Linking Existing Accounts

Users can manually link their accounts by:

1. Click their profile picture.
2. Select Settings.
3. Select Launch Account Connections.
4. Select Authelia.

## See Also

- [AdventureLog OIDC Social Authentication Documentation](https://adventurelog.app/docs/configuration/social_auth/oidc.html)

[Authelia]: https://www.authelia.com
[AdventureLog]: https://adventurelog.app/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
