---
title: "Synology DSM"
description: "Integrating Synology DSM with the Authelia OpenID Connect Provider."
lead: ""
date: 2022-10-18T21:22:13+11:00
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
  * [v4.35.6](https://github.com/authelia/authelia/releases/tag/v4.35.6)
* [Synology DSM]
  * v7.1

## Before You Begin

### Common Notes

1. You are *__required__* to utilize a unique client id for every client.
2. The client id on this page is merely an example and you can theoretically use any alphanumeric string.
3. You *__should not__* use the client secret in this example, We *__strongly recommend__* reading the
   [Generating Client Secrets] guide instead.

### Specific Notes

*__Important Note:__ [Synology DSM] does not support automatically creating users via [OpenID Connect]. It is therefore
recommended that you ensure Authelia and [Synology DSM] share a LDAP server.*

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://dsm.example.com/`
* __Authelia Root URL:__ `https://auth.example.com`
* __Client ID:__ `synology-dsm`
* __Client Secret:__ `synology-dsm_client_secret`

## Configuration

### Application

To configure [Synology DSM] to utilize Authelia as an [OpenID Connect] Provider:

1. Go to DSM.
2. Go to `Control Panel`.
3. Go To `Domain/LDAP`.
4. Go to `SSO Client`.
5. Check the `Enable OpenID Connect SSO service` checkbox in the `OpenID Connect SSO Service` section.
6. Configure the following values:
    * Profile: `OIDC`
    * Name: `Authelia`
    * Well Known URL: `https://auth.example.com/.well-known/openid-configuration`
    * Application ID: `synology-dsm`
    * Application Key: `synology-dsm_client_secret`
    * Redirect URL: `https://dsm.example.com`
    * Authorisation Scope: `openid profile groups email`
    * Username Claim: `preferred_username`
7. Save the settings.

{{< figure src="client.png" alt="Synology" width="736" >}}

### Authelia

The following YAML configuration is an example __Authelia__
[client configuration](../../../configuration/identity-providers/open-id-connect.md#clients) for use with [Synology DSM]
which will operate with the above example:

```yaml
- id: synology-dsm
  description: Synology DSM
  secret: '$plaintext$synology-dsm_client_secret'
  public: false
  authorization_policy: two_factor
  redirect_uris:
    - https://dsm.example.com
  scopes:
    - openid
    - profile
    - groups
    - email
  userinfo_signing_algorithm: none
```

## See Also

* [Synology DSM SSO Client Documentation](https://kb.synology.com/en-af/DSM/help/DSM/AdminCenter/file_directory_service_sso?version=7)

[Authelia]: https://www.authelia.com
[Synology DSM]: https://www.synology.com/en-global/dsm
[OpenID Connect]: ../../openid-connect/introduction.md
