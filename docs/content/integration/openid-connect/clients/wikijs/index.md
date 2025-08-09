---
title: "Wiki.js"
description: "Integrating Wiki.js with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-04-18T15:25:09+10:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/wikijs/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Wiki.js | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Wiki.js with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.38.8](https://github.com/authelia/authelia/releases/tag/v4.38.8)
- [Wiki.js]
  - [v2.5.301](https://github.com/requarks/wiki/releases/tag/v2.5.301)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://wiki.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `Wiki.js`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Wiki.js] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'wikijs'
        client_name: 'Wiki'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://wikijs.{{< sitevar name="domain" nojs="example.com" >}}/login/<UUID>/callback'  # Note this must be copied during step 7 of the Application configuration.
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

To configure [Wiki.js] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [Wiki.js] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [Wiki.js] as an Administrator.
2. Visit Administration.
3. Select `Modules` > `Authentication`.
4. Select `+ Add Strategy`.
5. Select `Generic OpenID Connect / OAuth2`.
6. Configure the following options:
   - Display Name: `Authelia`
   - Client ID: `wikijs`
   - Client Secret: `insecure_secret`
   - Authorization Endpoint URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization`
   - Token Endpoint URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token`
   - User Info Endpoint URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo`
   - Issuer URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
   - Email Claim: `email`
   - Display Name Claim: `name`
   - Map Groups: Disabled
   - Groups Claim: `groups`
   - Allow self-registration: Enabled
7. Copy the `Callback URL / Redirect URI` for the Authelia configuration.
8. Click Apply.

{{< figure src="wikijs.png" alt="Wiki.js" width="736" style="padding-right: 10px" >}}

## See Also

- [Wiki.js Authentication Guide](https://docs.requarks.io/auth)

[Wiki.js]: https://js.wiki/
[Authelia]: https://www.authelia.com
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
