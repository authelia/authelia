---
title: "YouTrack"
description: "Integrating YouTrack with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-04-18T23:36:08+00:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/youtrack/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "YouTrack | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring YouTrack with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.12](https://github.com/authelia/authelia/releases/tag/v4.39.12)
- [YouTrack]
  - 2025.1

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://youtrack.{{< sitevar name="domain" nojs="example.com" >}}/`
  - Also assumes that the [YouTrack] server is utilizing the inbuilt hub.
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `youtrack`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

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
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://youtrack.{{< sitevar name="domain" nojs="example.com" >}}/hub/api/rest/oauth2/auth'
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
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [YouTrack] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [YouTrack] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Visit the [YouTrack] Hub at `https://youtrack.{{< sitevar name="domain" nojs="example.com" >}}/hub`
2. Login as an admin
3. Navigate to Administration > Auth Modules > New Auth Module
4. For the identity provider select `OAuth 2.0` from the list of authentication protocols at the bottom
5. Configure the following options:
   1. General Settings:
      - Name: `Authelia`
      - Button image: feel free to upload one from the [branding](../../../reference/guides/branding.md) page
      - Client ID: `youtrack`
      - Client Secret: `insecure_secret`
      - Scope: `openid profile email`
   2. Authorization Service Endpoints:
      - Authorization: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization`
      - Token: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token`
      - User Data: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo`
   3. Field Mapping:
      - User ID: `sub`
      - Username: `preferred_username`
      - Full Name: `name`
      - Email: `email`
      - Email Verification State: `email_verified`
6. Verify the Redirect URI displayed on this page matches the one you configured in Authelia otherwise update Authelia's
   configuration
7. Click the `Enable module` button
8. Click the `Test login` button

{{< figure src="youtrack_overview.png" process="resize 800x" >}}

{{< figure src="youtrack_authz_endpoints.png" process="resize 800x" >}}

{{< figure src="youtrack_field_mapping.png" process="resize 800x" >}}

{{< figure src="youtrack_additional_settings.png" process="resize 800x" >}}

## See Also

- [YouTrack OAuth 2.0 Auth Module Documentation](https://www.jetbrains.com/help/youtrack/cloud/oauth2-authentication-module.html)

[YouTrack]: https://www.jetbrains.com/youtrack/
[Authelia]: https://www.authelia.com
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
