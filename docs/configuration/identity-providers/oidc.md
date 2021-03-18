---
layout: default
title: OpenID Connect
parent: Identity Providers
grand_parent: Configuration
nav_order: 2
---

# OpenID Connect

**Authelia** currently supports [OpenID Connect] but is preview status. This means it's suggested you only implement
it with caution. The main purpose of it being available is for us to allow users to try it and provide feedback. The 
reason we do it this way is [OpenID Connect] is a complicated technology to implement well, and we are more likely to
get good feedback if we allow people to test it. By default [OpenID Connect] is disabled unless you configure it.

```yaml
identity_providers:
  oidc:
    hmac_secret: this_is_a_secret_abc123abc123abc
    issuer_private_key: |
      --- KEY START
      --- KEY END
    clients:
      - id: myapp
        description: My Application
        secret: this_is_a_secret
        policy: two_factor
        redirect_uris:
          - https://oidc.example.com:8080/oauth2/callback
        scopes:
          - openid
          - groups
          - email
          - profile
        grant_types:
          - refresh_token
          - authorization_code
        response_types:
          - code
```

## Options

### hmac_secret

The HMAC secret used to sign the [OpenID Connect] JWT's. 
Can also be defined using a [secret](../secrets.md) which is the recommended for containerized deployments.

### issuer_private_key

The private key in DER base64 encoded PEM format used to encrypt the [OpenID Connect] JWT's. 
Can also be defined using a [secret](../secrets.md) which is the recommended for containerized deployments.

### clients

A list of clients to setup. The options for each client are described below.

#### id

The Client ID for this client. Must be configured in the application consuming this client.

#### description

A friendly description for this client shown in the UI. This defaults to the same as the ID.

#### secret

The shared secret between Authelia and the application consuming this client.

#### policy

The authentication policy for this client. Either `one_factor` or `two_factor`.

#### redirect_uris

A list of valid callback URL's this client will redirect to. All other callbacks will be considered unsafe.

#### scopes

A list of scopes to allow this client to consume. See [scope definitions](#scope-definitions) for more information.

#### grant_types

A list of grant types this client can return.

#### response_types

A list of response types this client can return.

## Scope Definitions

### openid

This is the default scope for openid and is required. Force this on.

|JWT Field|JWT Type     |Authelia Attribute|Description           |
|:-------:|:-----------:|:----------------:|:--------------------:|
|         |             |                  |                      |

### groups

This scope includes the groups the authentication backend reports the user is a member of in the JWT.

|JWT Field|JWT Type     |Authelia Attribute|Description           |
|:-------:|:-----------:|:----------------:|:--------------------:|
|groups   |array[string]| groups           |The users display name|

### email

This scope includes the email information the authentication backend reports about the user in the JWT.

|JWT Field|JWT Type|Authelia Attribute|Description                          |
|:-------:|:------:|:----------------:|:-----------------------------------:|
|email    |string  | email[0]         |The first email in the list of emails|

### profile

This scope includes the profile information the authentication backend reports about the user in the JWT.

|JWT Field|JWT Type|Authelia Attribute|Description           |
|:-------:|:------:|:----------------:|:--------------------:|
|name     |string  | display_name     |The users display name|


[OpenID Connect]: https://openid.net/connect/