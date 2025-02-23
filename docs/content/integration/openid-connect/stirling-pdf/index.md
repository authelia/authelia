---
title: "Stirling-PDF"
description: "Integrating Stirling-PDF with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-02-23T04:38:52+00:00
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
  * [v4.38.19](https://github.com/authelia/authelia/releases/tag/v4.38.19)
* [Stirling-PDF]
  * [0.42.0](https://github.com/Stirling-Tools/Stirling-PDF/releases/tag/v0.42.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://stirlingpdf.{{< sitevar name="domain" nojs="example.com" >}}/`
* __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
* __Client ID:__ `stirlingpdf`
* __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Stirling-PDF] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'stirlingpdf'
        client_name: 'Stirling-PDF'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        redirect_uris:
          - 'https://stirlingpdf.{{< sitevar name="domain" nojs="example.com" >}}/login/oauth2/code/oidc'
        scopes:
          - 'openid'
          - 'profile'
          - 'groups'
          - 'email'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
Stirling-PDF OIDC Login requires you to login with a user which isn't already registered with Stirling-PDF. You can rename your current ('web') user via [https://stirlingpdf.{{< sitevar name="domain" nojs="example.com" >}}/account](https://stirlingpdf.{{< sitevar name="domain" nojs="example.com" >}}/account)
{{< /callout >}}

To configure [Stirling-PDF] to utilize Authelia as an [OpenID Connect 1.0] Provider:

1. Include the [Stirling-PDF] environment variables for [OpenID Connect 1.0] configuration:

```env
DOCKER_ENABLE_SECURITY=true
SECURITY_ENABLE_LOGIN=true
SECURITY_OAUTH2_ENABLED=true
SECURITY_OAUTH2_AUTOCREATEUSER=true
SECURITY_OAUTH2_ISSUER=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
SECURITY_OAUTH2_CLIENTID=stirlingpdf
SECURITY_OAUTH2_CLIENTSECRET=insecure_secret
SECURITY_OAUTH2_BLOCKREGISTRATION=false
SECURITY_OAUTH2_SCOPES="openid, profile, email"
SECURITY_OAUTH2_USEASUSERNAME=preferred_username
SECURITY_OAUTH2_PROVIDER=Authelia
```

2. Once you successfully enabled Single Sign-on, you might want to disable the form login (and eventually block registration/auto user creation). This can be done by changing the following [Stirling-PDF] environment variables:

```env
SECURITY_LOGINMETHOD=oauth2 # Change this to all if you need to login with the web user which might be your Stirling-PDF administrator.
SECURITY_OAUTH2_AUTOCREATEUSER=false
SECURITY_OAUTH2_BLOCKREGISTRATION=true
```

## See Also

* [Stirling-PDF SSO Documentation](https://docs.stirlingpdf.com/Advanced%20Configuration/Single%20Sign-On%20Configuration)

[Authelia]: https://www.authelia.com
[Stirling-PDF]: https://www.stirlingpdf.com
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
