---
title: "Synapse"
description: "Integrating Synapse with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/synapse/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Synapse | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Synapse with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.15](https://github.com/authelia/authelia/releases/tag/v4.39.15)
- [Synapse]
  - [v1.127.1](https://github.com/element-hq/synapse/releases/tag/v1.127.1)

{{% oidc-common bugs="claims-hydration" %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://synapse.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `synapse`
- __Client Secret:__ `insecure_secret`
- __Groups:__ the `synapse-users` group exists and only members of this group are expected to be able to use Synapse.

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Synapse] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'synapse'
        client_name: 'Synapse'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://synapse.{{< sitevar name="domain" nojs="example.com" >}}/_synapse/client/oidc/callback'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
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

To configure [Synapse] there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

{{< callout context="tip" title="Did you know?" icon="outline/rocket" >}}
Generally the configuration file is named `homeserver.yaml`.
{{< /callout >}}

To configure [Synapse] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```yaml {title="homeserver.yaml"}
oidc_providers:
  - idp_id: authelia
    idp_name: 'Authelia'
    idp_icon: 'mxc://authelia.com/cKlrTPsGvlpKxAYeHWJsdVHI'
    discover: true
    issuer: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
    client_id: 'synapse'
    client_secret: 'insecure_secret'
    scopes:
     - 'openid'
     - 'profile'
     - 'email'
     - 'groups'
    allow_existing_users: true
    user_mapping_provider:
      config:
        subject_claim: 'sub'
        localpart_template: '{{ user.preferred_username }}'
        display_name_template: '{{ user.name }}'
        email_template: '{{ user.email }}'
    attribute_requirements:
     - attribute: 'groups'
       value: 'synapse-users'
```

#### Configuration Escape Hatch

{{% oidc-escape-hatch-claims-hydration example="disable" %}}

```yaml
oidc_providers:
  - user_profile_method: 'userinfo_endpoint'
```

## See Also

- [Synapse OpenID Connect Authentication Documentation](https://matrix-org.github.io/synapse/latest/openid.html)

[Authelia]: https://www.authelia.com
[Synapse]: https://github.com/matrix-org/synapse
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
