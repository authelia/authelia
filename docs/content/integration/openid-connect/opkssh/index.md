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
  - [v4.39.4](https://github.com/authelia/authelia/releases/tag/v4.39.4)
- [opkssh]
  - [v0.7.0](https://github.com/openpubkey/opkssh/releases/tag/v0.7.0)

{{% oidc-common bugs="claims-hydration" %}}

### Assumptions

This example makes the following assumptions:

- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `opkssh`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
At the time of this writing this third party client has a bug and does not support [OpenID Connect 1.0](https://openid.net/specs/openid-connect-core-1_0.html). This
configuration will likely require configuration of an escape hatch to work around the bug on their end. See
[Configuration Escape Hatch](#configuration-escape-hatch) for details.
{{< /callout >}}

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
          - 'offline_access'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
          - 'refresh_token'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'none'
```

#### Configuration Escape Hatch

{{% oidc-escape-hatch-claims-hydration client_id="opkssh" claims="email" %}}

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

##### Configuration File

{{< callout context="tip" title="Did you know?" icon="outline/rocket" >}}
Generally the configuration file is named `~/.opk/config.yml` on Linux and `C:\Users\{USER}\.opk\config.yml` on Windows.
{{< /callout >}}

To create a persistent configuration, generate a new configuration file by running the following command:

```shell
opkssh login --create-config
```

Then add Authelia to the existing providers:

```yaml{title="~/.opk/config.yml"}
providers:
  - alias: authelia
    issuer: https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
    client_id: opkssh
    scopes: openid offline_access profile email
    access_type: offline
    prompt: consent
    redirect_uris:
      - http://localhost:3000/login-callback
      - http://localhost:10001/login-callback
      - http://localhost:11110/login-callback
```

You can now run `opkssh login` to login.

## See Also

- [opkssh Custom OpenID Providers](https://github.com/openpubkey/opkssh?tab=readme-ov-file#custom-openid-providers-authentik-authelia-keycloak-zitadel)

[Authelia]: https://www.authelia.com
[opkssh]: https://github.com/openpubkey/opkssh
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
