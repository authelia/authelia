---
title: "AdventureLog"
description: "Integrating AdventureLog with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
weight: 620
toc: true
support:
  level: community
  versions: true
  integration: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.3](https://github.com/authelia/authelia/releases/tag/v4.39.3)
- [AdventureLog]
  - [v0.9.0](https://github.com/seanmorley15/AdventureLog/releases/tag/v0.9.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://adventurelog.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `adventurelog-authelia`
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
      - client_id: 'adventurelog-authelia'
        client_name: 'Adventure Log'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        redirect_uris:
          - 'https://adventurelog.{{< sitevar name="domain" nojs="example.com" >}}/accounts/oidc/adventurelog-authelia/login/callback/'
        scopes:
          - 'openid'
          - 'email'
          - 'profile'
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
4. Click Add.
5. Configure the following options:
   - Provider: `OpenID Connect`
   - Provider ID: `adventurelog-authelia`
   - Name: `Authelia`
   - Client ID: `adventurelog-authelia`
   - Secret Key: `insecure_secret`
   - Settings:
     `{"server_url": "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}"}`
   - Sites: Select the sites you want to enable OpenID Connect for.
6. Press `Save` at the bottom.

Note: the `Provider ID` and `Client ID` configured in step 5 must be identical.
This is a known bug in the Adventurelog frontend, see [Issue 544](https://github.com/seanmorley15/AdventureLog/issues/544) and [PR 556](https://github.com/seanmorley15/AdventureLog/pull/556).


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
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
