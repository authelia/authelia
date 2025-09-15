---
title: "Synology DSM"
description: "Integrating Synology DSM with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/synology-dsm/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Synology DSM | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Synology DSM with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.38.0](https://github.com/authelia/authelia/releases/tag/v4.38.0)
- [Synology DSM]
  - [v7.2](https://www.synology.com/en-global/releaseNote/DSM?os=DSM&version=7.2)

{{% oidc-common %}}

### Specific Notes

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
[Synology DSM](https://www.synology.com/en-global/dsm) does not support automatically creating users via [OpenID Connect 1.0](../../openid-connect/introduction.md). It is therefore
recommended that you ensure Authelia and [Synology DSM](https://www.synology.com/en-global/dsm) share an LDAP server (for DSM v7.1).
With DSM v7.2+ you have the possibility to also use local DSM accounts (see `Account type` below) and do not need to set
up a shared LDAP.
{{< /callout >}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://dsm.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `synology-dsm`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

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
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://dsm.{{< sitevar name="domain" nojs="example.com" >}}'
        scopes:
          - 'openid'
          - 'profile'
          - 'groups'
          - 'email'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

To configure [Synology DSM] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [Synology DSM] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Go to DSM.
2. Go to `Control Panel`.
3. Go To `Domain/LDAP`.
4. Go to `SSO Client`.
5. Check the `Enable OpenID Connect SSO service` checkbox in the `OpenID Connect SSO Service` section.
6. Configure the following options:
  - Profile: `OIDC`
  - Account type: `Domain/LDAP/local`
  - Name: `Authelia`
  - Well Known URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration`
  - Application ID: `synology-dsm`
  - Application Key: `insecure_secret`
  - Redirect URL: `https://dsm.{{< sitevar name="domain" nojs="example.com" >}}`
  - Authorisation Scope: `openid profile groups email`
  - Username Claim: `preferred_username`
7. Save the settings.

{{< figure src="client.png" alt="Synology" width="736" >}}

## See Also

- [Synology DSM SSO Client Documentation](https://kb.synology.com/en-af/DSM/help/DSM/AdminCenter/file_directory_service_sso?version=7)

[Authelia]: https://www.authelia.com
[Synology DSM]: https://www.synology.com/en-global/dsm
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
