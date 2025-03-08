---
title: "Zipline"
description: "Integrating Zipline with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2023-11-12T21:18:09+11:00
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

* [Authelia]
  * [v4.38.0](https://github.com/authelia/authelia/releases/tag/v4.38.0)
* [Zipline]
  * [4.0.0](https://github.com/diced/zipline/releases/tag/v4.0.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://zipline.{{< sitevar name="domain" nojs="example.com" >}}/`
* __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
* __Client ID:__ `zipline`
* __Client Secret:__ `insecure_secret`

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
        redirect_uris:
          - 'https://zipline.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/oauth/oidc'
        scopes:
          - 'openid'
          - 'email'
          - 'profile'
          - offline_access
        response_types: 'code'
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

To configure [Zipline] to utilize Authelia as an [OpenID Connect 1.0] Provider:

1. Go to Server Settings
2. Activate the **OAuth Registration** feature toggle
2. Configure:
   1. OIDC Client ID: `ziplinea`
   2. OIDC Client Secret: `insecure_secret`
   3. OIDC Authorize URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization`
   4. OIDC Token URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token`
   5. OIDC Userinfo URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo`
   6. OIDC Redirect URL can be left blank, but the default Zipline URL is HTTP, if you didn't activate **Return HTTPS URLs** in the Core settings, this impacts the OIDC Redirect URL
3. Click Save

{{< figure src="zipline.png" alt="Zipline configuration" width="300" >}}

Take a look at the [See Also](#see-also) section for the cheatsheets corresponding to the sections above for their
descriptions.

## See Also

- [Zipline]:
  - [OIDC documentation](https://zipline.diced.sh/docs/guides/oauth/oidc)

[Authelia]: https://www.authelia.com
[Zipline]: https://zipline.diced.sh/
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
