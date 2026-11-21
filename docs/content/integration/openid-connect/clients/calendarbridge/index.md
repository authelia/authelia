---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "CalendarBridge"
description: "A guide on integrating CalendarBridge with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
summary: ""
date: 2026-11-12T00:00:00+00:00
draft: false
images: []
weight: 620
toc: true
support:
  level: community
  versions: true
  integration: true
seo:
  title: "CalendarBridge | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring CalendarBridge with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [CalendarBridge]

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `calendarbridge`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [CalendarBridge]
which will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'calendarbridge'
        client_name: 'CalendarBridge'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://auth.calendarbridge.com/oauth2/idpresponse'  # See step 4 below.
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
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

To configure [CalendarBridge] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

Single sign-on requires a [CalendarBridge] Group (Organization) subscription, an organization administrator account,
and at least one verified domain on the organization. [CalendarBridge] matches users to the organization using the
`email` claim, so the email address of each user in Authelia must be on a verified domain of the organization.

To configure [CalendarBridge] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [CalendarBridge] as an organization administrator.
2. Navigate to `Manage your Organization`.
3. Navigate to the `Single Sign-On` tab and click `Configure SSO`.
4. Copy the displayed `Redirect URI` and ensure it matches the redirect URI configured in Authelia.
5. Configure the following options:
   - Protocol: `OIDC`
   - Domain: Select the verified domain your users login with
   - Identity Provider: Select the generic option
   - Discovery URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration`
   - Client ID: `calendarbridge`
   - Client Secret: `insecure_secret`
6. Click `Save Configuration`.

Once you have confirmed single sign-on works you can optionally enable `Enforce SSO` which makes sign-in through
Authelia mandatory for every user on the selected domain.

## See Also

- [CalendarBridge Configuring Single Sign-On Documentation](https://help.calendarbridge.com/group-admin-docs/configuring-sso/)
- [CalendarBridge Managing Authorized Domains Documentation](https://help.calendarbridge.com/group-admin-docs/managing-authorized-domains/)

[Authelia]: https://www.authelia.com
[CalendarBridge]: https://calendarbridge.com/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
