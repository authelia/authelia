---
title: "Forgejo"
description: "Integrating Forgejo with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-07-10T08:55:15+00:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/forgejo/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Forgejo | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Forgejo with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.13](https://github.com/authelia/authelia/releases/tag/v4.39.13)
- [Forgejo]
  - [v13.0.2](https://codeberg.org/forgejo/forgejo/releases/tag/v13.0.2)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://forgejo.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `forgejo`
- __Client Secret:__ `insecure_secret`
- __Authentication Name (Forgejo):__ `authelia`:
    - This option determines the redirect URI in the format of
      `https://forgejo.{{< sitevar name="domain" nojs="example.com" >}}/user/oauth2/<Authentication Name>/callback`.
      This means if you change this value you need to update the redirect URI.

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Forgejo] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'forgejo'
        client_name: 'Forgejo'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://forgejo.{{< sitevar name="domain" nojs="example.com" >}}/user/oauth2/authelia/callback'
        scopes:
          - 'openid'
          - 'email'
          - 'profile'
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

To configure [Forgejo] there are two methods, using the [Web GUI](#web-gui), or using the [CLI](#cli).

#### Web GUI

To configure [Forgejo] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Expand User Options
2. Visit Site Administration
3. Visit Authentication Sources
4. Visit Add Authentication Source
5. Configure the following options:
   - Authentication Name: `authelia`
   - OAuth2 Provider: `OpenID Connect`
   - Client ID (Key): `forgejo`
   - Client Secret: `insecure_secret`
   - OpenID Connect Auto Discovery URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration`
   - Additional scopes: `email profile groups`
6. Optionally update the following setting to specify which oidc groups have admin access to Forgejo
    - Claim name providing group names for this source. (Optional)


{{< figure src="forgejo.png" alt="Forgejo" width="300" >}}

#### CLI

_**Important Note:** Please refer to the [Forgejo CLI Guide](https://forgejo.org/docs/latest/admin/command-line) regarding the correct usage of the CLI._

To configure [Forgejo] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Run `forgejo migrate`.
2. Run `forgejo admin auth add-oauth --provider=openidConnect --name=authelia --key=forgejo --secret=insecure_secret --auto-discover-url=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration --scopes='openid email profile groups'`


### Automatic User Creation

To configure [Forgejo] to perform automatic user creation for the `{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}` domain via [OpenID Connect 1.0]:

1. Edit the following values in the [Forgejo] `app.ini`:
```ini
[openid]
ENABLE_OPENID_SIGNIN = false
ENABLE_OPENID_SIGNUP = true
WHITELISTED_URIS     = {{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}

[service]
DISABLE_REGISTRATION                          = false
ALLOW_ONLY_EXTERNAL_REGISTRATION              = true
SHOW_REGISTRATION_BUTTON                      = false
```
## Optional Configuration

### Authelia
To configure Forgejo to sync ssh public keys from Authelia you can define `sshpubkey` as a multi-valued [Extra Attribute](../../../../reference/guides/attributes.md#extra-attributes) and __combine__ the following configuration with the configuration.yml above.
``` yaml {title="configuration.yml"}
identity_providers:
  oidc:
    claims_policies:
      forgejo:
        custom_claims:
          sshpubkey: {}
    scopes:
      forgejo:
        claims:
          - sshpubkey
    clients:
      - client_id: 'forgejo'
        claims_policy: 'forgejo'
        scopes:
          - 'forgejo'
```
### Application

Forgejo configuration largely follows the instructions from [Web GUI](#web-gui) and [CLI](#cli)

#### Web GUI
Follow the instructions in [Web GUI](#web-gui) with the following additions

5. Configure the following options:
   - Additional scopes: `email profile groups forgejo`
   - Public SSH key attribute: `sshpubkey`


{{< figure src="forgejo-sshpubkey.png" alt="Forgejo" width="300" >}}

#### CLI
Follow the instructions from [CLI](#cli), and change the following command:

2. Run `forgejo admin auth add-oauth --auto-discover-url=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration --name=authelia --provider=openidConnect  --key=forgejo --secret=insecure_secret  --scopes='openid email profile groups forgejo' --attribute-ssh-public-key=sshpubkey`

## See Also

- [Forgejo]
  - [Config Cheat Sheet](https://forgejo.org/docs/latest/admin/config-cheat-sheet/)
    - [OpenID](https://forgejo.org/docs/latest/admin/config-cheat-sheet/#openid-openid)
    - [Service](https://forgejo.org/docs/latest/admin/config-cheat-sheet/#service-service)

[Authelia]: https://www.authelia.com
[Forgejo]: https://forgejo.org/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
