---
layout: default
title: Proxmox
parent: Community-Tested OIDC Integrations
grand_parent: Community
nav_order: 2
---

# OIDC Integrations: Proxmox

{{ page.path }}

## Authelia config

**Note** these setting have been tested with authelia `v4.33.2` and Proxmox `7.1-10`

The specific client config for proxmox.

```yaml
identity_providers:
  oidc:
    clients:
      - id: some id you want to use on the client
        description: Some description you want to shown on the Authelia consent page
        secret: some secret string which should also be entered in the proxmox config
        public: false
        authorization_policy: two_factor
        audience: []
        scopes:
          - openid
        redirect_uris:
          - https://proxmox.example.com
        userinfo_signing_algorithm: none
```

## Proxmox config

Under Datacenter go to **Persmission > Realms** and add the an OpenID Connect Server

<p align="center">
  <a href="../../images/portainer.gif" target="_blank"><img src="../../images/proxmox.gif" width="736"></a>
</p>