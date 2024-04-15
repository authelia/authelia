---
title: "Argo CD"
description: "Integrating Argo CD with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2022-07-13T04:27:30+10:00
draft: false
images: []
weight: 620
toc: true
community: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

* [Authelia]
  * [v4.38.0](https://github.com/authelia/authelia/releases/tag/v4.38.0)
* [Argo CD]
  * v2.4.5

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://argocd.example.com/`
* __Authelia Root URL:__ `https://auth.example.com/`
* __Client ID:__ `argocd`
* __Client Secret:__ `insecure_secret`
* __CLI Client ID:__ `argocd-cli`

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
        redirect_uris:
          - 'https://argocd.example.com/auth/callback'
        scopes:
          - 'openid'
          - 'groups'
          - 'email'
          - 'profile'
        userinfo_signed_response_alg: 'none'
      - client_id: 'argocd-cli'
        client_name: 'Argo CD (CLI)'
        public: true
        authorization_policy: 'two_factor'
        redirect_uris:
          - 'http://localhost:8085/auth/callback'
        scopes:
          - 'openid'
          - 'groups'
          - 'email'
          - 'profile'
          - 'offline_access'
        userinfo_signed_response_alg: 'none'
```

### Application

To configure [Argo CD] to utilize Authelia as an [OpenID Connect 1.0] Provider use the following configuration:

```yaml
name: 'Authelia'
issuer: 'https://auth.example.com'
clientID: 'argocd'
clientSecret: 'insecure_secret'
cliClientID: 'argocd-cli'
requestedScopes:
  - 'openid'
  - 'profile'
  - 'email'
  - 'groups'
```

## See Also

* [Argo CD OpenID Connect Documentation](https://argo-cd.readthedocs.io/en/stable/operator-manual/user-management/#existing-oidc-provider)

[Authelia]: https://www.authelia.com
[Argo CD]: https://argo-cd.readthedocs.io/en/stable/
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
