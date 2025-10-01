---
title: "Proxmox"
description: "Integrating Proxmox with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/proxmox/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Proxmox | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Proxmox with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.11](https://github.com/authelia/authelia/releases/tag/v4.39.11)
- [Proxmox Virtual Environment]
  - [v8.4.1](https://pve.proxmox.com/wiki/Roadmap#Proxmox_VE_8.4)
- [Proxmox Backup Server]
  - [v3.4.2](https://pbs.proxmox.com/wiki/index.php/Roadmap#Proxmox_Backup_Server_3.4)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://proxmox.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `proxmox`
- __Client Secret:__ `insecure_secret`
- __Realm__ `authelia`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Proxmox Virtual Environment] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'proxmox'
        client_name: 'Proxmox'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://proxmox.{{< sitevar name="domain" nojs="example.com" >}}'
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

To configure [Proxmox Virtual Environment] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [Proxmox Virtual Environment] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Visit Datacenter.
2. Visit Permission.
3. Visit Realms.
4. Add an OpenID Connect Server.
5. Configure the following options:
   - Issuer URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
   - Realm: `authelia`
   - Client ID: `proxmox`
   - Client Key: `insecure_secret`
   - Username Claim: `Default (subject)`
   - Scopes: `openid email profile groups`
   - Autocreate Users: Enable if you want users to automatically be created in [Proxmox].
   - Autocreate Groups: Enable if you want groups to automatically be created in [Proxmox].
   - Groups Claim: Set to `groups` to add users to existing proxmox groups.

{{< figure src="proxmox.png" alt="Proxmox" process="resize 600x" >}}

## See Also

- [Proxmox User Management Documentation](https://pve.proxmox.com/wiki/User_Management)

[Authelia]: https://www.authelia.com
[Proxmox Virtual Environment]: https://pve.proxmox.com
[Proxmox Backup Server]: https://pbs.proxmox.com
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
