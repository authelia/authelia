---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Oracle Cloud"
description: "A guide on integrating Oracle Cloud with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Oracle Cloud | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Oracle Cloud with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Oracle Cloud Identity Domain URL:** `https://idcs-example.identity.oraclecloud.com/`
  - This option determines the redirect URI in the format of
    `https://idcs-example.identity.oraclecloud.com/oauth2/v1/social/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `oracle-cloud`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Oracle Cloud] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'oracle-cloud'
        client_name: 'Oracle Cloud'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://idcs-example.identity.oraclecloud.com/oauth2/v1/social/callback'
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

### Application

To configure [Oracle Cloud] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [Oracle Cloud] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to the [Oracle Cloud] console as an administrator.
2. Open the navigation menu and navigate to `Identity & Security`, `Domains`.
3. Select the identity domain, select the `Federation` tab, and under `Actions` select `Add Social IdP`.
4. Configure the following options:
   - Type: `OpenID Connect`
   - Name: `Authelia`
   - Client ID: `oracle-cloud`
   - Client secret: `insecure_secret`
   - Discovery service URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration`
   - Enable Just-In-Time (JIT) provisioning: Enabled
5. Click `Add`.
6. Click the `Actions` menu for the new identity provider, select `Activate IdP`, and confirm.
7. Navigate to the `Identity provider policies` for the identity domain and edit the rule which should offer Authelia.
8. Under `Assign identity providers` add `Authelia` and save the rule.

## See Also

- [Oracle Cloud Adding a Social Identity Provider Documentation](https://docs.oracle.com/en-us/iaas/Content/Identity/identityproviders/add-social-identity-provider.htm)
- [Oracle Cloud Identity Provider Policies Documentation](https://docs.oracle.com/en-us/iaas/Content/Identity/idppolicies/list-idp-policies.htm)

[Authelia]: https://www.authelia.com
[Oracle Cloud]: https://www.oracle.com/cloud/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
