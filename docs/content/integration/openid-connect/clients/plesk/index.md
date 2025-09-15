---
title: "Plesk"
description: "Integrating Plesk with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-04-26T18:35:57+10:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/plesk/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Plesk | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Plesk with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.10](https://github.com/authelia/authelia/releases/tag/v4.39.10)
- [Plesk]
  - [v18.0.69](https://docs.plesk.com/release-notes/obsidian/change-log/#plesk-18069)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://plesk.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `plesk`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

The following example uses the [OAuth login Extension] which is assumed to be installed when following
this section of the guide.

To install the [OAuth login Extension] for [Plesk] via the Web GUI:

1. Login to [Plesk].
2. Navigate to `Extensions`.
3. Navigate to `Extensions Catalog`.
4. Search for `OAuth login`.
5. Click Install.

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Plesk] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'plesk'
        client_name: 'Plesk'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://plesk.{{< sitevar name="domain" nojs="example.com" >}}/modules/oauth/public/login.php'
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

To configure [Plesk] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [Plesk] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [Plesk].
2. Navigate to Extensions.
3. Navigate to OAuth login.
4. Toggle the switch into the on position.
5. Configure the following options:
   - Type: `OpenID Connect`
   - Client ID: `plesk`
   - Client Secret: `insecure_secret`
   - Callback Host: `https://plesk.{{< sitevar name="domain" nojs="example.com" >}}`
   - Authorize URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization`
   - Token URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token`
   - Userinfo URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo`
   - Scopes: `openid,email,profile`
   - Login Button Text: `Login with Authelia`
6. Press `Save` at the bottom.

## See Also

- [Plesk OIDC documentation](https://ljpc.solutions/contact)

[Authelia]: https://www.authelia.com
[Plesk]: https://www.plesk.com
[OAuth login Extension]: https://www.plesk.com/extensions/oauth/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
