---
title: "Proxmox"
description: "Integrating Proxmox with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
weight: 620
toc: true
support:
  level: community
  versions: true
  integration: true
aliases:
  - /docs/community/oidc-integrations/proxmox.html
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

* [Authelia]
  * [v4.38.17](https://github.com/authelia/authelia/releases/tag/v4.38.17)
* [Proxmox]
  * [8.3.0](https://pve.proxmox.com/wiki/Roadmap#Proxmox_VE_8.3)

{{% oidc-common %}}

### Specific Notes

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
[Proxmox](https://www.proxmox.com/) requires you create the Realm before adding the provider. This is not covered in this
guide.
{{< /callout >}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://proxmox.{{< sitevar name="domain" nojs="example.com" >}}/`
* __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
* __Client ID:__ `proxmox`
* __Client Secret:__ `insecure_secret`
* __Realm__ `authelia`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Proxmox] which will
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
        userinfo_signed_response_alg: 'none'
```

### Application

To configure [Proxmox] to utilize Authelia as an [OpenID Connect 1.0] Provider:

1. Visit Datacenter
2. Visit Permission
3. Visit Realms
4. Add an OpenID Connect Server
5. Set the following values:
   1. Issuer URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
   2. Realm: `authelia`
   3. Client ID: `proxmox`
   4. Client Key: `insecure_secret`
   5. Username Claim `preferred_username`
   6. Scopes: `openid profile email`
   7. Enable *Autocreate Users* if you want users to automatically be created in [Proxmox].

{{< figure src="proxmox.png" alt="Proxmox" width="736" style="padding-right: 10px" >}}

## See Also

* [Proxmox User Management Documentation](https://pve.proxmox.com/wiki/User_Management)

[Authelia]: https://www.authelia.com
[Proxmox]: https://www.proxmox.com/
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
