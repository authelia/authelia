---
title: "YouTrack"
description: "Integrating YouTrack with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-04-19T09:33:58+10:00
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
  * [v4.39.1](https://github.com/authelia/authelia/releases/tag/v4.39.1)
* [YouTrack]
  * 2025.1

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://youtrack.{{< sitevar name="domain" nojs="example.com" >}}/`
  * Also assumes that the [YouTrack] server is utilizing the inbuilt hub.
* __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
* __Client ID:__ `youtrack`
* __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

{{% oidc-conformance-claims %}}

The following YAML configuration is an example __Authelia__ [client configuration] for use with [YouTrack] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'youtrack'
        client_name: 'YouTrack'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        redirect_uris:
          - 'https://youtrack.{{< sitevar name="domain" nojs="example.com" >}}/hub/api/rest/oauth2/auth'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
        userinfo_signed_response_alg: 'none'
```

### Application

To configure [YouTrack] to utilize Authelia as an [OpenID Connect 1.0] Provider:

1. Visit the [YouTrack] Hub At [https://youtrack.{{< sitevar name="domain" nojs="example.com" >}}/hub](https://youtrack.{{< sitevar name="domain" nojs="example.com" >}}/hub)
2. Login as an admin
3. Navigate to Administration > Auth Modules > New Auth Module
4. For the identity provider select `OAuth 2.0` from the list of authentication protocols at the bottom
5. Fill in the following fields:
   1. Name: `Authelia`
   2. Button image: feel free to upload one from the [branding](../../../reference/guides/branding.md) page
   3. Client ID: `youtrack`
   4. Client Secret: `insecure_secret`
   5. Authorization Service Endpoints:
      1. Authorization: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization`
      2. Token: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token`
      3. User Data: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo`
   6. Field Mapping:
      1. User ID: `sub`
      2. Username: `preferred_username`
      3. Full Name: `name`
      4. Email: `email`
      5. Email Verification State: `email_verified`
   7. Scope: `openid profile email`
6. Click the `Enable module` button
7. Click the `Test login` button

## See Also

- [YouTrack OAuth 2.0 Auth Module Documentation](https://www.jetbrains.com/help/youtrack/cloud/oauth2-authentication-module.html)

[YouTrack]: https://min.io/
[Authelia]: https://www.authelia.com
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
