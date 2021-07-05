---
layout: default
title: OpenID Connect
parent: Identity Providers
grand_parent: Configuration
nav_order: 2
---

# OpenID Connect

**Authelia** currently supports the [OpenID Connect] OP role as a [**beta**](#beta) feature. The OP role is the 
[OpenID Connect] Provider role, not the Relaying Party or RP role. This means other applications that implement the 
[OpenID Connect] RP role can use Authelia as an authentication and authorization backend similar to how you may use 
social media or development platforms for login.

The Relaying Party role is the role which allows an application to use GitHub, Google, or other [OpenID Connect]
providers for authentication and authorization. We do not intend to support this functionality at this moment in time.

## Roadmap

We have decided to implement [OpenID Connect] as a beta feature, it's suggested you only utilize it for testing and
providing feedback, and should take caution in relying on it in production as of now. [OpenID Connect] and it's related endpoints
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

¹ _This stage has not been implemented as of yet_.

² _This individual feature has not been implemented as of yet_.

## Configuration

The following snippet provides a sample-configuration for the OIDC identity provider explaining each field in detail.

```yaml
identity_providers:
  oidc:
    hmac_secret: this_is_a_secret_abc123abc123abc              # [1]
    issuer_private_key: |                                      # [2]
      --- KEY START
      --- KEY END
    
    clients:
      - id: myapp                                              # [3]
        description: My Application                            # [4]
        secret: this_is_a_secret                               # [5]
        authorization_policy: two_factor                       # [6]
        redirect_uris:                                         # [7]
          - https://oidc.example.com:8080/oauth2/callback
        scopes:                                                # [8]
          - openid
          - groups
          - email
          - profile
        grant_types:                                           # [9]
          - refresh_token
          - authorization_code
        response_types:                                        # [10]
          - code
```

### Options

#### [1] hmac_secret

The HMAC secret used to sign the [OpenID Connect] JWT's. The provided string is hashed to a SHA256 byte string for the purpose of meeting the required format. It can also be defined using a [secret](../secrets.md) which is the recommended for containerized deployments. You must [generate this option yourself](#generating-options-yourself).

#### [2] issuer_private_key

The private key in DER base64 encoded PEM format used to encrypt the [OpenID Connect] JWT's. Can also be defined using a [secret](../secrets.md) which is the recommended for containerized deployments. The reason for using only the private key here is that one is able to calculate the public key easily from the private key in this format (`openssl rsa -in rsa.key -pubout > rsa.pem`).

You must [generate this option yourself](#generating-options-yourself). To create this option, use `docker run -u "$(id -u):$(id -g)" -v "$(pwd)":/keys docker.io/authelia/authelia:latest authelia rsa generate --dir /keys` to generate both the private and public key in the current directory. You can then paste the private key into your configuration.

#### clients

A list of clients to configure. The options for each client are described below.

##### [3] id

The Client ID for this client. It must exactly match the Client ID configured in the application consuming this client.

##### [4] description

A friendly description for this client shown in the UI. This defaults to the same as the ID.

##### [5] secret

The shared secret between Authelia and the application consuming this client. This secret must match the secret configured in the application. Currently this is stored in plain text. You must [generate this option yourself](#generating-options-yourself).

##### [6] authorization_policy

The authorization policy for this client: either `one_factor` or `two_factor`.

##### [7] redirect_uris

A list of valid callback URL´s this client will redirect to. All other callbacks will be considered unsafe. The URL's are case-sensitive and they differ from application to application - we have provided [a list of URL´s for common applications](../../community/oidc-integrations.md).

##### [8] scopes

A list of scopes to allow this client to consume. See [scope definitions](#scope-definitions) for more information. The documentation for the application you want to use with Authelia will most-likely provide you with the scopes to grant.

##### [9] grant_types

A list of grant types this client can return. _It is recommended that this isn't configured at this time unless you know what you're doing_.

##### [10] response_types

A list of response types this client can return. _It is recommended that this isn't configured at this time unless you  know what you're doing._

### Generating Options Yourself

If you must generate an option yourself, you can use a random string of sufficient length. The command `openssl rand -base64 32` provides such a random string with base64-conform characters. For Kubernetes, see [this section too](../secrets.md#Kubernetes).

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

[//]: # (Links)

[OpenID Connect]: https://openid.net/connect/
