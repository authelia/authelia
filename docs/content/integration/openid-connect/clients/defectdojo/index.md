---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "DefectDojo"
description: "A guide on integrating DefectDojo with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "DefectDojo | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring DefectDojo with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [DefectDojo] Pro
  - [v3.5.101](https://docs.defectdojo.com/releases/pro/changelog/)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://defectdojo.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://defectdojo.{{< sitevar name="domain" nojs="example.com" >}}/complete/oidc/`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `defectdojo`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

{{< callout context="note" title="DefectDojo Pro" icon="outline/info-circle" >}}
As of DefectDojo 3.0 single sign-on is only available in DefectDojo Pro. The open source edition of DefectDojo only
supports local username and password login.
{{< /callout >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [DefectDojo] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'defectdojo'
        client_name: 'DefectDojo'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://defectdojo.{{< sitevar name="domain" nojs="example.com" >}}/complete/oidc/'
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

To configure [DefectDojo] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [DefectDojo] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [DefectDojo] as an administrator.
2. Navigate to `Connect` > `Authorization` > `Providers`.
3. Select `OpenID Connect`.
4. Configure the following options:
   - Endpoint: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
   - Client ID: `defectdojo`
   - Client Secret: `insecure_secret`
   - Enable OIDC: Enabled
5. Optionally click `Validate Config` and ensure the redirect URI shown is
   `https://defectdojo.{{< sitevar name="domain" nojs="example.com" >}}/complete/oidc/`.
6. Click `Submit`.

[DefectDojo] only requests the `openid`, `profile`, and `email` scopes, and Authelia only releases the `groups` claim
when the `groups` scope is requested. As such the [DefectDojo] Group Mapping feature can not be used with Authelia.

## See Also

- [DefectDojo Pro OIDC Documentation](https://docs.defectdojo.com/admin/sso/pro__oidc/)

[Authelia]: https://www.authelia.com
[DefectDojo]: https://defectdojo.com/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
