---
title: "opkssh"
description: "Integrating OpenPubkey SSH with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-04-04T10:36:34+00:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/opkssh/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "opkssh | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring opkssh with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.13](https://github.com/authelia/authelia/releases/tag/v4.39.13)
- [opkssh]
  - [v0.10.0](https://github.com/openpubkey/opkssh/releases/tag/v0.10.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `opkssh`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

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

### Application

To configure [opkssh] to utilize Authelia as an [OpenID Connect 1.0] Provider:

#### Client

To log in using Authelia run:

```shell
opkssh login --provider=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}},opkssh
```

You will now see your unique user identifier `sub` in the CLI, copy it to set up the access control on the server.

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
    scopes: openid offline_access
    access_type: offline
    prompt: consent
    redirect_uris:
      - http://localhost:3000/login-callback
      - http://localhost:10001/login-callback
      - http://localhost:11110/login-callback
```

You can now run `opkssh login` to login.

#### Server

To configure [opkssh] there is one method, using the [Configuration File](#configuration-file).

##### Configuration File

{{< callout context="tip" title="Did you know?" icon="outline/rocket" >}}
Generally the configuration file is named `/etc/opk/providers`.
{{< /callout >}}

To configure [opkssh] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```txt {title="/etc/opk/providers"}
https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}} opkssh 24h
```

In addition to above, the CLI will need to be used to map users manually.

For example allow the user `john` with the user identifier of `f0919359-9d15-4e15-bcba-83b41620a073` to login as `root` :

```shell
opkssh add root f0919359-9d15-4e15-bcba-83b41620a073 https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
```

To set up access control not just for yourself but other users as well, use the [authelia storage user identifiers export](https://www.authelia.com/reference/cli/authelia/authelia_storage_user_identifiers_export/)
command to get all user identifiers.

## See Also

- [opkssh Custom OpenID Providers](https://github.com/openpubkey/opkssh?tab=readme-ov-file#custom-openid-providers-authentik-authelia-keycloak-zitadel)

[Authelia]: https://www.authelia.com
[opkssh]: https://github.com/openpubkey/opkssh
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
