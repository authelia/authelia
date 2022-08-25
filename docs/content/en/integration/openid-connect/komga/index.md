---
title: "Komga"
description: "Integrating Komga with the Authelia OpenID Connect Provider."
lead: ""
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  integration:
    parent: "openid-connect"
weight: 620
toc: true
community: true
aliases:
  - /docs/community/oidc-integrations/komga.html
---

## Tested Versions

* [Authelia]
  * [v4.36.4](https://github.com/authelia/authelia/releases/tag/v4.36.4)
* [Komga] 
  * [v0.157.1](https://github.com/gotson/komga/releases/tag/v0.157.1)

## Before You Begin

You are required to utilize a unique client id and a unique and random client secret for all [OpenID Connect] relying
parties. You should not use the client secret in this example, you should randomly generate one yourself. You may also
choose to utilize a different client id, it's completely up to you.

This example makes the following assumptions:

* __Application Root URL:__ `https://komga.example.com`
* __Authelia Root URL:__ `https://auth.example.com`
* __Client ID:__ `komga-auth`
* __Client Secret:__ `komga_client_secret`

## Configuration

### Application

To configure [Komga] to utilize Authelia as an [OpenID Connect] Provider:

1. Create an `Application.yml` according to the [configuration options](https://komga.org/installation/configuration.html)
2. Add a section that describes the spring boot security configuration


```spring:
  security:
    oauth2:
      client:
        registration:
          authelia:
            client-id: `komga-auth`
            client-secret: `komga_client_secret`
            client-name: Authelia
            scope: openid, email
            authorization-grant-type: authorization_code
            redirect-uri: "{baseScheme}://{baseHost}{basePort}{basePath}/login/oauth2/code/authelia"
        provider:
          authelia:
            issuer-uri: `https:\\auth.example.com`
            user-name-attribute: email
````

### Optional configuration

You can enable some useful additional debug logging to `application.yml` by adding the `logging.level.org.springframework.security attribute`:

```
logging:
  file.name: /config/logs/komga.log
  level:
    org:
      springframework:
        security: info   #when changed to 'TRACE' adds additional spring security logging on top of komga logging.
      gotson:
        komga: info
```

Automatic creation of accounts (in Komga) by logging in with Authelia can be enabled with:

```
komga:
  oauth2-account-creation: true
```

In certain cases it might be necessary to add:

```
server:
  use-forward-headers: true
```


### Authelia

The following YAML configuration is an example __Authelia__
[client configuration](../../../configuration/identity-providers/open-id-connect.md#clients) for use with [Portainer]
which will operate with the above example:

```yaml
      -
        id: komga-auth
        description: Komga Comics OpenID
        secret: `komga_client_secret`
        public: false
        authorization_policy: two_factor
        audience: []
        scopes:
          - openid
          - email
        redirect_uris:
          - https://komga.example.com/login/oauth2/code/authelia

        grant_types:
          - authorization_code

        userinfo_signing_algorithm: none
```

Note: make sure that the `userinfo_signing_algorithm` is set to `none`, or Komga will throw an `application\jwt` error.


## See Also

* [Portainer OAuth Documentation](https://docs.portainer.io/admin/settings/authentication/oauth)

[Authelia]: https://www.authelia.com
[Komga]: https://www.komga.org
[OpenID Connect]: ../../openid-connect/introduction.md
