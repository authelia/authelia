---
title: "Argo CD"
description: "Integrating Argo CD with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/argocd/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Argo CD | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Argo CD with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.38.0](https://github.com/authelia/authelia/releases/tag/v4.38.0)
- [Argo CD]
  - [v2.4.5](https://github.com/argoproj/argo-cd/releases/tag/v2.4.5)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://argocd.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `argocd`
- __Client Secret:__ `insecure_secret`
- __CLI Client ID:__ `argocd-cli`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Argo CD] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'argocd'
        client_name: 'Argo CD'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://argocd.{{< sitevar name="domain" nojs="example.com" >}}/auth/callback'
        scopes:
          - 'openid'
          - 'groups'
          - 'email'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
      - client_id: 'argocd-cli'
        client_name: 'Argo CD (CLI)'
        public: true
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'http://localhost:8085/auth/callback'
        scopes:
          - 'openid'
          - 'offline_access'
          - 'groups'
          - 'email'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
          - 'refresh_token'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'none'
```

### Application

To configure [Argo CD] there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

{{< callout context="tip" title="Did you know?" icon="outline/rocket" >}}
Generally the configuration file is named `argocd-cm.yaml`.
{{< /callout >}}

To configure [Argo CD] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```yaml {title="argocd-cm.yaml"}
oidc.config: |
  name: 'Authelia'
  issuer: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
  clientID: 'argocd'
  clientSecret: 'insecure_secret'
  cliClientID: 'argocd-cli'
  requestedScopes:
    - 'openid'
    - 'email'
    - 'groups'
  enableUserInfoGroups: true
  userInfoPath: '/api/oidc/userinfo'
```

##### Group Mapping

You can use the following example to map the `argocd-admins` Authelia group to the `admin` role.

```csv {title="policy.csv"}
g, argocd-admins, role:admin
```

## See Also

- [Argo CD OpenID Connect Documentation](https://argo-cd.readthedocs.io/en/stable/operator-manual/user-management/#existing-oidc-provider)

[Authelia]: https://www.authelia.com
[Argo CD]: https://argo-cd.readthedocs.io/en/stable/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
