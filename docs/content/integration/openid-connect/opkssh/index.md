---
title: "opkssh"
description: "Integrating OpenPubkey SSH with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-04-04T10:36:34+00:00
draft: false
images: []
weight: 620
toc: true
support:
  level: community
  versions: true
  integration: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.1](https://github.com/authelia/authelia/releases/tag/v4.39.1)
- [opkssh]
  - [v0.4.0](https://github.com/openpubkey/opkssh/releases/tag/v0.4.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `opkssh`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

{{% oidc-conformance-claims claims="email" %}}

The following YAML configuration is an example __Authelia__ [client configuration] for use with [opkssh] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'opkssh'
        client_name: 'opkssh'
        public: true
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'http://localhost:3000/login-callback'
          - 'http://localhost:10001/login-callback'
          - 'http://localhost:11110/login-callback'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [opkssh] to utilize Authelia as an [OpenID Connect 1.0] Provider:

#### Server

To configure [opkssh] there is one method, using the [Configuration File](#configuration-file).

##### Configuration File

{{< callout context="tip" title="Did you know?" icon="outline/rocket" >}}
Generally the configuration file is named `/etc/opk/providers`.
{{< /callout >}}

To configure [opkssh] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```txt {title="/etc/opk/providers"}
https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/ opkssh 24h
```

In addition to above, the CLI will need to be used to map users manually.

For example allow the user `john@example.com` to login as `root` :

```shell
opkssh add root john@example.com https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/
```

#### Client

To log in using Authelia run:

```shell
opkssh login --provider=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/,opkssh
```

## See Also

- [opkssh Custom OpenID Providers](https://github.com/openpubkey/opkssh?tab=readme-ov-file#custom-openid-providers-authentik-authelia-keycloak-zitadel)

[Authelia]: https://www.authelia.com
[opkssh]: https://github.com/openpubkey/opkssh
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
