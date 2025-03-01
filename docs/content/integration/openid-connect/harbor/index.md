---
title: "Harbor"
description: "Integrating Harbor with the Authelia OpenID Connect 1.0 Provider."
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

* [Authelia]
  * [v4.38.15](https://github.com/authelia/authelia/releases/tag/v4.38.15)
* [Harbor]
  * 2.11.2

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://harbor.{{< sitevar name="domain" nojs="example.com" >}}/`
* __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
* __Client ID:__ `harbor`
* __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Harbor] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'harbor'
        client_name: 'Harbor'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        redirect_uris:
          - 'https://harbor.{{< sitevar name="domain" nojs="example.com" >}}/c/oidc/callback'
        grant_types:
          - 'refresh_token'
          - 'authorization_code'
        scopes:
          - 'openid'
          - 'profile'
          - 'groups'
          - 'email'
          - 'offline_access'
        userinfo_signed_response_alg: 'none'
```

### Application

To configure [Harbor] to utilize Authelia as an [OpenID Connect 1.0] Provider:

1. Visit Administration
2. Visit Configuration
3. Visit Authentication
4. Select `OIDC` from the `Auth Mode` drop down
5. Set the following values:
   1. OIDC Provider Name: `Authelia`
   2. OIDC Provider Endpoint: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
   3. OIDC Client ID: `harbor`
   4. OIDC Client Secret: `insecure_secret`
   5. Group Claim Name: `groups`
   6. For OIDC Admin Group you can specify a group name that matches your authentication backend.
   7. OIDC Scope: `openid,profile,email,groups,offline_access`
   8. Ensure `Verify Certificate` is checked.
   9. Ensure `Automatic onboarding` is checked if you want users to be created by default.
   10. Username Claim: `preferred_username`
6. Click `Test OIDC Server`
7. Click `Save`

## See Also

* [Harbor OpenID Connect Provider Documentation](https://goharbor.io/docs/2.11.0/administration/configure-authentication/oidc-auth/)

[Authelia]: https://www.authelia.com
[Harbor]: https://goharbor.io/
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
