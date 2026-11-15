---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "osTicket"
description: "A guide on integrating osTicket with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "osTicket | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring osTicket with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [osTicket]
  - [v1.18.4](https://github.com/osTicket/osTicket/releases/tag/v1.18.4)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://osticket.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://osticket.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/oauth2`.
    This means if you change this value, you need to update the redirect URI. If [osTicket] is installed in a
    subdirectory, the subdirectory must be included before `/api/auth/oauth2`.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `osticket`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [osTicket] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'osticket'
        client_name: 'osTicket'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://osticket.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/oauth2'
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

To configure [osTicket] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
[osTicket] can automatically create End User accounts during sign in, however Agent accounts must already exist with an
email address which matches the email address of the Authelia user.
{{< /callout >}}

To configure [osTicket] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Download the `Authentication :: Oauth2` plugin from the [osTicket Download Page](https://osticket.com/download)
   and place it in the `include/plugins` directory of your [osTicket] installation.
2. Login to [osTicket] as an administrator.
3. Navigate to `Admin Panel`.
4. Navigate to `Manage`.
5. Navigate to `Plugins`.
6. Click `Add New Plugin` and follow the instructions to install the `Oauth2` plugin.
7. Click the `Oauth2` plugin, set `Status` to `Active`, and click `Save Changes`.
8. Navigate to the `Instances` tab.
9. Click `Add New Instance`.
10. Select `OAuth2 - Other`.
11. Configure the following options on the `Instance` tab:
    - Name: `Authelia`
    - Status: `Enabled`
12. Configure the following options on the `Config` tab:
    - Name: `Authelia`
    - Authentication Target: `Agents and End Users`
    - Authentication Label: `Authelia`
    - Redirect URI: `https://osticket.{{< sitevar name="domain" nojs="example.com" >}}/api/auth/oauth2`
    - Client Id: `osticket`
    - Client Secret: `insecure_secret`
    - Authorization Endpoint: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization`
    - Token Endpoint: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token`
    - Resource Details Endpoint: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo`
    - Scopes: `openid email profile`
    - User Identifier: `email`
    - Given Name: `name`
    - Surname: leave this field empty
    - Email Address: `email`
13. Click `Save Changes`.
14. If Agent authentication was enabled, navigate to `Admin Panel`, then `Agents`, and ensure each Agent which
    should use Authelia has the `Authelia` authentication backend or `Use any available backend` selected.

## See Also

- [osTicket OAuth2 Guide](https://docs.osticket.com/en/latest/Guides/OAuth2%20Guide.html)

[Authelia]: https://www.authelia.com
[osTicket]: https://osticket.com/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
