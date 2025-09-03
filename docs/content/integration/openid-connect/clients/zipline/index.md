---
title: "Zipline"
description: "Integrating Zipline with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-03-04T23:12:34+11:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/zipline/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Zipline | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Zipline with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.8](https://github.com/authelia/authelia/releases/tag/v4.39.8)
- [Zipline]
  - [v4.2.3](https://github.com/diced/zipline/releases/tag/v4.2.3)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://zipline.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `zipline`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Zipline] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'zipline'
        client_name: 'Zipline'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://zipline.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/oauth/oidc'
        scopes:
          - 'openid'
          - 'offline_access'
          - 'email'
          - 'profile'
        response_types:
          - 'code'
        grant_types:
          - 'refresh_token'
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

To configure [Zipline] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [Zipline] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Go to Server Settings.
2. Activate the **OAuth Registration** feature toggle.
3. Configure the following options:
   - OIDC Client ID: `zipline`
   - OIDC Client Secret: `insecure_secret`
   - OIDC Authorize URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization`
   - OIDC Token URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token`
   - OIDC Userinfo URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo`
   - OIDC Redirect URL can be left blank, but the default Zipline URL is HTTP, if you didn't activate **Return HTTPS URLs** in the Core settings, this impacts the OIDC Redirect URL
4. Click Save.

{{< figure src="zipline.png" alt="Zipline configuration" width="300" >}}

## See Also

- [Zipline]:
  - [OIDC documentation](https://zipline.diced.sh/docs/guides/oauth/oidc)

[Authelia]: https://www.authelia.com
[Zipline]: https://zipline.diced.sh/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
