---
title: "Komga"
description: "Integrating Komga with the Authelia OpenID Connect Provider."
lead: ""
date: 2022-08-26T11:39:00+10:00
draft: false
images: []
menu:
  integration:
    parent: "openid-connect"
weight: 620
toc: true
community: true
---

## Tested Versions

* [Authelia]
  * [v4.36.4](https://github.com/authelia/authelia/releases/tag/v4.36.4)
* [Komga]
  * [v0.157.1](https://github.com/gotson/komga/releases/tag/v0.157.1)

## Before You Begin

### Common Notes

1. You are *__required__* to utilize a unique client id for every client.
2. The client id on this page is merely an example and you can theoretically use any alphanumeric string.
3. You *__should not__* use the client secret in this example, We *__strongly recommend__* reading the
   [Generating Client Secrets] guide instead.

[Generating Client Secrets]: ../specific-information.md#generating-client-secrets

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://komga.example.com`
* __Authelia Root URL:__ `https://auth.example.com`
* __Client ID:__ `komga`
* __Client Secret:__ `komga_client_secret`

## Configuration

### Application

To configure [Komga] to utilize Authelia as an [OpenID Connect] Provider:

1. Configure the security section of the [Komga] configuration:
```yaml
komga:
  ## Comment if you don't want automatic account creation.
  oauth2-account-creation: true
spring:
  security:
    oauth2:
      client:
        registration:
          authelia:
            client-id: `komga`
            client-secret: `komga_client_secret`
            client-name: Authelia
            scope: openid,profile,email
            authorization-grant-type: authorization_code
            redirect-uri: "{baseScheme}://{baseHost}{basePort}{basePath}/login/oauth2/code/authelia"
        provider:
          authelia:
            issuer-uri: https://auth.example.com
            user-name-attribute: preferred_username
````

### Authelia

The following YAML configuration is an example __Authelia__
[client configuration](../../../configuration/identity-providers/open-id-connect.md#clients) for use with [Komga]
which will operate with the above example:

```yaml
- id: komga
  description: Komga
  secret: '$plaintext$komga_client_secret'
  public: false
  authorization_policy: two_factor
  redirect_uris:
    - https://komga.example.com/login/oauth2/code/authelia
  scopes:
    - openid
    - profile
    - email
  grant_types:
    - authorization_code
  userinfo_signing_algorithm: none
```

## See Also

* [Komga Configuration options Documentation](https://komga.org/installation/configuration.html)
* [Komga Social login Documentation](https://komga.org/installation/oauth2.html)

[Authelia]: https://www.authelia.com
[Komga]: https://www.komga.org
[OpenID Connect]: ../../openid-connect/introduction.md
