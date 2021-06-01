---
layout: default
title: OpenID Connect
parent: Identity Providers
grand_parent: Configuration
nav_order: 2
---

# OpenID Connect

**Authelia** currently supports the [OpenID Connect] OP role as a [beta](#beta) feature. The OP role is the 
[OpenID Connect] Provider role, not the Relaying Party or RP role. This means other applications that implement the 
[OpenID Connect] RP role can use Authelia as an authentication and authorization backend similar to how you may use 
social media or development platforms for login.

The Relaying Party role is the role which allows an application to use GitHub, Google, or other [OpenID Connect]
providers for authentication and authorization. We do not intend to support this functionality at this moment in time.

## Beta

We have decided to implement [OpenID Connect] as a beta feature, it's suggested you only utilize it for testing and
providing feedback, and should take caution in relying on it in production. [OpenID Connect] and it's related endpoints
are not enabled by default unless you specifically configure the [OpenID Connect] section.

The beta will be broken up into stages. Each stage will bring additional features. The following table is a *rough* plan
for which stage will have each feature, and may evolve over time:

<table>
    <thead>
      <tr>
        <th class="tbl-header">Stage</th>
        <th class="tbl-header">Feature Description</th>
      </tr>
    </thead>
    <tbody>
      <tr>
        <td rowspan="7" class="tbl-header tbl-beta-stage">beta1</td>
        <td><a href="https://openid.net/specs/openid-connect-core-1_0.html#Consent" target="_blank" rel="noopener noreferrer">User Consent</a></td>
      </tr>
      <tr>
        <td><a href="https://openid.net/specs/openid-connect-core-1_0.html#CodeFlowSteps" target="_blank" rel="noopener noreferrer">Authorization Code Flow</a></td>
      </tr>
      <tr>
        <td><a href="https://openid.net/specs/openid-connect-discovery-1_0.html" target="_blank" rel="noopener noreferrer">OpenID Connect Discovery</a></td>
      </tr>
      <tr>
        <td>RS256 Signature Strategy</td>
      </tr>
      <tr>
        <td>Per Client Scope/Grant Type/Response Type Restriction</td>
      </tr>
      <tr>
        <td>Per Client Authorization Policy (1FA/2FA)</td>
      </tr>
      <tr>
        <td class="tbl-beta-stage">Per Client List of Valid Redirection URI's</td>
      </tr>
      <tr>
        <td rowspan="2" class="tbl-header tbl-beta-stage">beta2 <sup>1</sup></td>
        <td>Token Storage</td>
      </tr>
      <tr>
        <td class="tbl-beta-stage">Audit Storage</td>
      </tr>
      <tr>
        <td rowspan="4" class="tbl-header tbl-beta-stage">beta3 <sup>1</sup></td>
        <td><a href="https://openid.net/specs/openid-connect-backchannel-1_0.html" target="_blank" rel="noopener noreferrer">Back-Channel Logout</a></td>
      </tr>
      <tr>
        <td>Deny Refresh on Session Expiration</td>
      </tr>
      <tr>
        <td><a href="https://openid.net/specs/openid-connect-messages-1_0-20.html#rotate.sig.keys" target="_blank" rel="noopener noreferrer">Signing Key Rotation Policy</a></td>
      </tr>
      <tr>
        <td class="tbl-beta-stage">Client Secrets Hashed in Configuration</td>
      </tr>
      <tr>
        <td class="tbl-header tbl-beta-stage">GA <sup>1</sup></td>
        <td class="tbl-beta-stage">General Availability after previous stages are vetted for bug fixes</td>
      </tr>
      <tr>
        <td rowspan="2" class="tbl-header">misc</td>
        <td>List of other features that may be implemented</td>
      </tr>
      <tr>
        <td class="tbl-beta-stage"><a href="https://openid.net/specs/openid-connect-frontchannel-1_0.html" target="_blank" rel="noopener noreferrer">Front-Channel Logout</a> <sup>2</sup></td>
      </tr>
    </tbody>
</table>

*<sup>1</sup> this stage has not been implemented as of yet*

*<sup>2</sup> this individual feature has not been implemented as of yet*

## Configuration

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
        authorization_policy: two_factor
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

The HMAC secret used to sign the [OpenID Connect] JWT's. The provided string is hashed to a SHA256 byte string for
the purpose of meeting the required format.

Can also be defined using a [secret](../secrets.md) which is the recommended for containerized deployments.

### issuer_private_key

The private key in DER base64 encoded PEM format used to encrypt the [OpenID Connect] JWT's.

Can also be defined using a [secret](../secrets.md) which is the recommended for containerized deployments.

### clients

A list of clients to configure. The options for each client are described below.

#### id

The Client ID for this client. Must be configured in the application consuming this client.

#### description

A friendly description for this client shown in the UI. This defaults to the same as the ID.

#### secret

The shared secret between Authelia and the application consuming this client. Currently this is stored in plain text.

#### authorization_policy

The authorization policy for this client. Either `one_factor` or `two_factor`.

#### redirect_uris

A list of valid callback URL's this client will redirect to. All other callbacks will be considered unsafe. The URL's
are case-sensitive.

#### scopes

A list of scopes to allow this client to consume. See [scope definitions](#scope-definitions) for more information.

#### grant_types

A list of grant types this client can return. It is recommended that this isn't configured at this time unless you know
what you're doing. 

#### response_types

A list of response types this client can return. It is recommended that this isn't configured at this time unless you 
know what you're doing.

## Scope Definitions

### openid

This is the default scope for openid. This field is forced on every client by the configuration
validation that Authelia does.

|JWT Field|JWT Type     |Authelia Attribute|Description                             |
|:-------:|:-----------:|:----------------:|:--------------------------------------:|
|sub      |string       |Username          |The username the user used to login with|
|scope    |string       |scopes            |Granted scopes (space delimited)        |
|scp      |array[string]|scopes            |Granted scopes                          |
|iss      |string       |hostname          |The issuer name, determined by URL      |
|at_hash  |string       |_N/A_             |Access Token Hash                       |
|auth_time|number       |_N/A_             |Authorize Time                          |
|aud      |array[string]|_N/A_             |Audience                                |
|exp      |number       |_N/A_             |Expires                                 |
|iat      |number       |_N/A_             |Issued At                               |
|rat      |number       |_N/A_             |Requested At                            |
|jti      |string(uuid) |_N/A_             |JWT Identifier                          |

### groups

This scope includes the groups the authentication backend reports the user is a member of in the token.

|JWT Field|JWT Type     |Authelia Attribute|Description           |
|:-------:|:-----------:|:----------------:|:--------------------:|
|groups   |array[string]|Groups            |The users display name|

### email

This scope includes the email information the authentication backend reports about the user in the token.

|JWT Field     |JWT Type|Authelia Attribute|Description                                              |
|:------------:|:------:|:----------------:|:-------------------------------------------------------:|
|email         |string  |email[0]          |The first email in the list of emails                    |
|email_verified|bool    |_N/A_             |If the email is verified, assumed true for the time being|

### profile

This scope includes the profile information the authentication backend reports about the user in the token.

|JWT Field|JWT Type|Authelia Attribute|Description           |
|:-------:|:------:|:----------------:|:--------------------:|
|name     |string  | display_name     |The users display name|


[OpenID Connect]: https://openid.net/connect/