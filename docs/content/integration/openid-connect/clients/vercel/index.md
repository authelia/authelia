---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Vercel"
description: "A guide on integrating Vercel with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Vercel | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Vercel with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Vercel]
  - Not Applicable

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
[Vercel] single sign-on requires the Enterprise plan or the Pro plan with the SAML SSO add-on, and must be configured
by a team owner. The Authelia instance must be reachable from the public internet.
{{< /callout >}}

{{% oidc-common bugs="claims-hydration" %}}

### Assumptions

This example makes the following assumptions:

- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `vercel`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Vercel] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'vercel'
        client_name: 'Vercel'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - ''  # See step 6 below.
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
        token_endpoint_auth_method: 'client_secret_basic'
```

#### Configuration Escape Hatch

{{% oidc-escape-hatch-claims-hydration client_id="vercel" claims="email,email_verified,name,given_name,family_name" %}}

### Application

To configure [Vercel] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [Vercel] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [Vercel] as a team owner and ensure your team is selected in the team switcher.
2. Navigate to `Settings`.
3. Navigate to `Security & Privacy`.
4. In the `Authentication and User Provisioning` section, click `Configure` in the `SAML` row.
5. Select `Custom OIDC` as the provider type.
6. Copy the `Login redirect URL` displayed by Vercel and configure the Authelia `redirect_uris` option. It will be in
   the format of `https://auth.vercel.com/sso/oidc/<identifier>/callback`.
7. Configure the following options:
   - Discovery endpoint: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration`
   - Client ID: `vercel`
   - Client Secret: `insecure_secret`
8. Save the configuration.

Team members can sign in using the `Continue with SAML SSO` button on the Vercel login page, or directly by going to
`https://vercel.com/login?saml=<team_slug>`.

## See Also

- [Vercel SAML Single Sign-On Documentation](https://vercel.com/docs/saml)

[Authelia]: https://www.authelia.com
[Vercel]: https://vercel.com/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
