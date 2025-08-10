---
title: "Windmill"
description: "Integrating Windmill with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/windmill/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Windmill | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Windmill with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.38.0](https://github.com/authelia/authelia/releases/tag/v4.38.0)
- [Windmill]
  - [v1.224.0](https://github.com/windmill-labs/windmill/releases/tag/v1.224.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://windmill.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `windmill`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Windmill]
which will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'windmill'
        client_name: 'Windmill'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://windmill.{{< sitevar name="domain" nojs="example.com" >}}/user/login_callback/authelia'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
          - 'groups'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

## Application

To configure [Windmill] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [Windmill] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Navigate to Superadmin settings.
2. Navigate to Core.
3. Configure the following options:
   - Base Url: `https://windmill.{{< sitevar name="domain" nojs="example.com" >}}`
4. Click Save.
5. Navigate to Superadmin settings.
6. Navigate to SSO/OAuth.
7. Configure the following options:
   - Config URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
   - Client ID: `windmill`
   - Client Secret: `insecure_secret`

## See Also

- [Windmill OpenID Connect Documentation](https://www.windmill.dev/docs/misc/setup_oauth)

[Authelia]: https://www.authelia.com
[Windmill]: https://www.windmill.dev
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
