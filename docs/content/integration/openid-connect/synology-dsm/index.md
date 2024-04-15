---
title: "Synology DSM"
description: "Integrating Synology DSM with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2022-10-18T21:22:13+11:00
draft: false
images: []
weight: 620
toc: true
community: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

* [Authelia]
  * [v4.38.0](https://github.com/authelia/authelia/releases/tag/v4.38.0)
* [Synology DSM]
  * v7.1
  * v7.2

{{% oidc-common %}}

### Specific Notes

*__Important Note:__ [Synology DSM] does not support automatically creating users via [OpenID Connect 1.0]. It is therefore
recommended that you ensure Authelia and [Synology DSM] share an LDAP server (for DSM v7.1).
With DSM v7.2+ you have the possibility to also use local DSM accounts (see `Account type` below) and do not need to set up a shared LDAP.*

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://dsm.example.com/`
* __Authelia Root URL:__ `https://auth.example.com/`
* __Client ID:__ `synology-dsm`
* __Client Secret:__ `insecure_secret`

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Synology DSM] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'synology-dsm'
        client_name: 'Synology DSM'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        redirect_uris:
          - 'https://dsm.example.com'
        scopes:
          - 'openid'
          - 'profile'
          - 'groups'
          - 'email'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

To configure [Synology DSM] to utilize Authelia as an [OpenID Connect 1.0] Provider:

1. Go to DSM.
2. Go to `Control Panel`.
3. Go To `Domain/LDAP`.
4. Go to `SSO Client`.
5. Check the `Enable OpenID Connect SSO service` checkbox in the `OpenID Connect SSO Service` section.
6. Configure the following values:
  * Profile: `OIDC`
  * Account type: `Domain/LDAP/local` (Note: Account type is supported DSM v7.2+)
  * Name: `Authelia`
  * Well Known URL: `https://auth.example.com/.well-known/openid-configuration`
  * Application ID: `synology-dsm`
  * Application Key: `insecure_secret`
  * Redirect URL: `https://dsm.example.com`
  * Authorisation Scope: `openid profile groups email`
  * Username Claim: `preferred_username`
7. Save the settings.

{{< figure src="client.png" alt="Synology" width="736" >}}

## See Also

* [Synology DSM SSO Client Documentation](https://kb.synology.com/en-af/DSM/help/DSM/AdminCenter/file_directory_service_sso?version=7)

[Authelia]: https://www.authelia.com
[Synology DSM]: https://www.synology.com/en-global/dsm
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
