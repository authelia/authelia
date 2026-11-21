---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "OneDev"
description: "A guide on integrating OneDev with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "OneDev | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring OneDev with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [OneDev]
  - [v17.2.0](https://github.com/theonedev/onedev/releases/tag/v17.2.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://onedev.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://onedev.{{< sitevar name="domain" nojs="example.com" >}}/~sso/callback/Authelia`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `onedev`
- **Client Secret:** `insecure_secret`
- **Provider Name:** `Authelia`
  - This option determines the last segment of the redirect URI and is case-sensitive.

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [OneDev] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'onedev'
        client_name: 'OneDev'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://onedev.{{< sitevar name="domain" nojs="example.com" >}}/~sso/callback/Authelia'
        scopes:
          - 'openid'
          - 'email'
          - 'profile'
          - 'groups'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [OneDev] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [OneDev] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [OneDev] as an administrator.
2. Navigate to `Administration` > `System Setting` and ensure the `Server URL` is set to
   `https://onedev.{{< sitevar name="domain" nojs="example.com" >}}`.
3. Navigate to `Administration` > `External Authentication` > `Single Sign On`.
4. Add a new provider.
5. Configure the following options:
   - Name: `Authelia`
   - Type: `OpenID`
   - Configuration Discovery Url: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration`
   - Client Id: `onedev`
   - Client Secret: `insecure_secret`
6. Expand `More Settings` and configure the following options:
   - Request Scopes: `openid email profile groups`
   - Groups Claim: `groups`
7. Click `Save`.

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
Groups returned in the `groups` claim are matched against the names of groups created under `Administration` >
`Group Management` in [OneDev].
{{< /callout >}}

## See Also

- [OneDev SSO with Okta Documentation](https://docs.onedev.io/tutorials/security/sso-with-okta)

[Authelia]: https://www.authelia.com
[OneDev]: https://onedev.io/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
