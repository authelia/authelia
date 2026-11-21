---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Omnissa Workspace ONE Access"
description: "A guide on integrating Omnissa Workspace ONE Access with the Authelia OpenID Connect 1.0 Provider for SSO, with configuration examples and available options."
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
  title: "Omnissa Workspace ONE Access | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Omnissa Workspace ONE Access with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Omnissa Workspace ONE Access]
  - Not Applicable

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `omnissa-workspace-one-access`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with
[Omnissa Workspace ONE Access] which will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'omnissa-workspace-one-access'
        client_name: 'Omnissa Workspace ONE Access'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - ''  # See step 4 below.
          - 'awgb://oauth2'  # Used by the Workspace ONE mobile applications.
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

To configure [Omnissa Workspace ONE Access] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [Omnissa Workspace ONE Access] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the
following instructions:

1. Login to the [Omnissa Workspace ONE Access] console as an administrator.
2. Navigate to `Integrations` > `Identity Providers`.
3. Click `Add` and select `OpenID Connect IDP`.
4. Copy the URL listed under `Integrate with Open ID Connect Provider using Redirect URI below` in the `Redirect URI`
   section and configure the Authelia `redirect_uris` option.
5. Configure the following options:
   - Identity Provider Name: `Authelia`
   - Configuration Type: `Automatic Discovery`
   - Configuration URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration`
   - Client ID: `omnissa-workspace-one-access`
   - Client Secret: `insecure_secret`
   - Open ID User Identifier Attribute: `preferred_username`
   - Omnissa Access User Identifier Attribute: `userName`
   - Users: Select the directories whose users are allowed to authenticate using Authelia
   - Network: Select the network ranges this identity provider can be used from
   - Authentication Method Name: `Authelia`
6. Click `Add`.
7. Navigate to `Resources` > `Policies`.
8. Edit the access policy which targets the applications you want to use Authelia for, and add the `Authelia`
   authentication method to the relevant policy rules.

## See Also

- [Omnissa Add and Configure an OpenID Connect Third-Party Identity Provider in Omnissa Access Documentation](https://docs.omnissa.com/bundle/workspace-one-access-managing-authentication-guideVSaaS/page/AddandConfigureanOpenIDConnectThird-PartyIdentityProviderinWorkspaceONEAccessCloudOnly.html)
- [Omnissa Managing the Default Access Policy in Omnissa Access Documentation](https://docs.omnissa.com/bundle/workspace-one-access-managing-authentication-guideVSaaS/page/ManagingtheDefaultAccessPolicyinWorkspaceONEAccess.html)

[Authelia]: https://www.authelia.com
[Omnissa Workspace ONE Access]: https://www.omnissa.com/products/omnissa-access/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
