---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Headlamp"
description: "A guide on integrating Headlamp with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
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
  title: "Headlamp | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Headlamp with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Headlamp]
  - [v0.46.0](https://github.com/kubernetes-sigs/headlamp/releases/tag/v0.46.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://headlamp.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://headlamp.{{< sitevar name="domain" nojs="example.com" >}}/oidc-callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `headlamp`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Headlamp] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    claims_policies:
      headlamp:
        id_token: ['email', 'email_verified', 'groups', 'preferred_username', 'name']
    clients:
      - client_id: 'headlamp'
        client_name: 'Headlamp'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        claims_policy: 'headlamp'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://headlamp.{{< sitevar name="domain" nojs="example.com" >}}/oidc-callback'
        scopes:
          - 'openid'
          - 'offline_access'
          - 'email'
          - 'profile'
          - 'groups'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
          - 'refresh_token'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

{{< callout context="note" title="Token Authentication" icon="outline/info-circle" >}}
[Headlamp] forwards the ID Token to the Kubernetes API server, which is why the `email`, `groups`, and other claims
are included in the ID Token using the `headlamp` claims policy. The Kubernetes API server must be configured to trust
Authelia as an OpenID Connect 1.0 issuer with the `headlamp` client ID as the audience. See the
[Kube Login](../kubelogin/index.md#kubernetes-api-server-configuration) guide for an example of this configuration.
{{< /callout >}}

### Application

To configure [Headlamp] there are two methods, using the [Helm Chart](#helm-chart), or using
[Environment Variables](#environment-variables).

#### Helm Chart

To configure [Headlamp] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following Helm values:

```yaml {title="values.yaml"}
config:
  oidc:
    clientID: 'headlamp'
    clientSecret: 'insecure_secret'
    issuerURL: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
    scopes: 'offline_access,email,profile,groups'
    callbackURL: 'https://headlamp.{{< sitevar name="domain" nojs="example.com" >}}/oidc-callback'
    usePKCE: true
```

#### Environment Variables

To configure [Headlamp] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following environment
variables:

##### Standard

```shell {title=".env"}
HEADLAMP_CONFIG_OIDC_CLIENT_ID=headlamp
HEADLAMP_CONFIG_OIDC_CLIENT_SECRET=insecure_secret
HEADLAMP_CONFIG_OIDC_IDP_ISSUER_URL=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
HEADLAMP_CONFIG_OIDC_SCOPES=offline_access,email,profile,groups
HEADLAMP_CONFIG_OIDC_CALLBACK_URL=https://headlamp.{{< sitevar name="domain" nojs="example.com" >}}/oidc-callback
HEADLAMP_CONFIG_OIDC_USE_PKCE=true
```

## See Also

- [Headlamp Accessing using OpenID Connect Documentation](https://headlamp.dev/docs/latest/installation/in-cluster/oidc)
- [Kubernetes OpenID Connect Tokens Documentation](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#openid-connect-tokens)

[Authelia]: https://www.authelia.com
[Headlamp]: https://headlamp.dev/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
