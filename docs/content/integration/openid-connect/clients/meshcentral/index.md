---
title: "MeshCentral"
description: "Integrating MeshCentral with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-04-26T18:35:57+10:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/meshcentral/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "MeshCentral | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring MeshCentral with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.9](https://github.com/authelia/authelia/releases/tag/v4.39.9)
- [MeshCentral]
  - [v1.1.44](https://github.com/Ylianst/MeshCentral/releases/tag/1.1.44)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://meshcentral.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `meshcentral`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [MeshCentral] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'meshcentral'
        client_name: 'MeshCentral'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        redirect_uris:
          - 'https://meshcentral.{{< sitevar name="domain" nojs="example.com" >}}/auth-oidc-callback'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [MeshCentral] there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

```json {title="config.json"}
{
  "domains": {
    "": {
      "title": "Example",
      "title2": "{{< sitevar name="domain" nojs="example.com" >}}",
      "authStrategies": {
        "oidc": {
          "issuer": "https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}",
          "clientid": "meshcentral",
          "clientsecret": "insecure_secret",
          "newAccounts": true
        }
      }
    }
  }
}
```

## See Also

- [MeshCentral Generic OpenID Connect Setup Guide](https://ylianst.github.io/MeshCentral/meshcentral/#generic-openid-connect-setup)

[MeshCentral]: https://meshcentral.com/
[Authelia]: https://www.authelia.com
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
