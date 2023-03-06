---
title: "OpenID Connect"
description: "An introduction into integrating the Authelia OpenID Connect Provider with an OpenID Connect relying party"
lead: "An introduction into integrating the Authelia OpenID Connect Provider with an OpenID Connect relying party."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  integration:
    parent: "openid-connect"
weight: 610
toc: true
aliases:
  - /docs/community/oidc-integrations.html
---

Authelia can act as an [OpenID Connect 1.0] Provider as part of an open beta. This section details implementation
specifics that can be used for integrating Authelia with an [OpenID Connect 1.0] Relying Party, as well as specific
documentation for some [OpenID Connect 1.0] Relying Party implementations.

See the [configuration documentation](../../configuration/identity-providers/open-id-connect.md) for information on how
to configure the Authelia [OpenID Connect 1.0] Provider.

## Scope Definitions

### openid

This is the default scope for [OpenID Connect 1.0]. This field is forced on every client by the configuration validation
that Authelia does.

*__Important Note:__ The subject identifiers or `sub` [Claim] has been changed to a [RFC4122] UUID V4 to identify the
individual user as per the [Subject Identifier Types] section of the [OpenID Connect 1.0] specification. Please use the
`preferred_username` [Claim] instead.*

|  [Claim]  |   JWT Type    | Authelia Attribute |                         Description                         |
|:---------:|:-------------:|:------------------:|:-----------------------------------------------------------:|
|    iss    |    string     |      hostname      |             The issuer name, determined by URL              |
|    jti    | string(uuid)  |       *N/A*        |     A [RFC4122] UUID V4 representing the JWT Identifier     |
|    rat    |    number     |       *N/A*        |            The time when the token was requested            |
|    exp    |    number     |       *N/A*        |                           Expires                           |
|    iat    |    number     |       *N/A*        |             The time when the token was issued              |
| auth_time |    number     |       *N/A*        |        The time the user authenticated with Authelia        |
|    sub    | string(uuid)  |     opaque id      |    A [RFC4122] UUID V4 linked to the user who logged in     |
|   scope   |    string     |       scopes       |              Granted scopes (space delimited)               |
|    scp    | array[string] |       scopes       |                       Granted scopes                        |
|    aud    | array[string] |       *N/A*        |                          Audience                           |
|    amr    | array[string] |       *N/A*        | An [RFC8176] list of authentication method reference values |
|    azp    |    string     |    id (client)     |                    The authorized party                     |
| client_id |    string     |    id (client)     |                        The client id                        |

### offline_access

This scope is a special scope designed to allow applications to obtain a [Refresh Token] which allows extended access to
an application on behalf of a user. A [Refresh Token] is a special [Access Token] that allows refreshing previously
issued token credentials, effectively it allows the relying party to obtain new tokens periodically.

Generally unless an application supports this and actively requests this scope they should not be granted this scope via
the client configuration.

### groups

This scope includes the groups the authentication backend reports the user is a member of in the [Claims] of the
[ID Token].

| [Claim] |   JWT Type    | Authelia Attribute |                                               Description                                               |
|:-------:|:-------------:|:------------------:|:-------------------------------------------------------------------------------------------------------:|
| groups  | array[string] |       groups       | List of user's groups discovered via [authentication](../../configuration/first-factor/introduction.md) |

### email

This scope includes the email information the authentication backend reports about the user in the [Claims] of the
[ID Token].

|     Claim      |   JWT Type    | Authelia Attribute |                        Description                        |
|:--------------:|:-------------:|:------------------:|:---------------------------------------------------------:|
|     email      |    string     |      email[0]      |       The first email address in the list of emails       |
| email_verified |     bool      |       *N/A*        | If the email is verified, assumed true for the time being |
|   alt_emails   | array[string] |     email[1:]      |  All email addresses that are not in the email JWT field  |

### profile

This scope includes the profile information the authentication backend reports about the user in the [Claims] of the
[ID Token].

|       Claim        | JWT Type | Authelia Attribute |               Description                |
|:------------------:|:--------:|:------------------:|:----------------------------------------:|
| preferred_username |  string  |      username      | The username the user used to login with |
|        name        |  string  |    display_name    |          The users display name          |

## Parameters

The following section describes advanced parameters which can be used in various endpoints as well as their related
configuration options.

### Grant Types

The following describes the various [OAuth 2.0] and [OpenID Connect 1.0] grant types and their support level. The value
field is both the required value for the `grant_type` parameter in the authorization request and the `grant_types`
configuration option.

|                   Grant Type                    | Supported |                     Value                      |                                Notes                                |
|:-----------------------------------------------:|:---------:|:----------------------------------------------:|:-------------------------------------------------------------------:|
|         [OAuth 2.0 Authorization Code]          |    Yes    |              `authorization_code`              |                                                                     |
| [OAuth 2.0 Resource Owner Password Credentials] |    No     |                   `password`                   | This Grant Type has been deprecated and should not normally be used |
|         [OAuth 2.0 Client Credentials]          |    Yes    |              `client_credentials`              |                                                                     |
|              [OAuth 2.0 Implicit]               |    Yes    |                   `implicit`                   | This Grant Type has been deprecated and should not normally be used |
|            [OAuth 2.0 Refresh Token]            |    Yes    |                `refresh_token`                 |                                                                     |
|             [OAuth 2.0 Device Code]             |    No     | `urn:ietf:params:oauth:grant-type:device_code` |                                                                     |
|

[OAuth 2.0 Authorization Code]: https://datatracker.ietf.org/doc/html/rfc6749#section-1.3.1
[OAuth 2.0 Implicit]: https://datatracker.ietf.org/doc/html/rfc6749#section-1.3.2
[OAuth 2.0 Resource Owner Password Credentials]: https://datatracker.ietf.org/doc/html/rfc6749#section-1.3.3
[OAuth 2.0 Client Credentials]: https://datatracker.ietf.org/doc/html/rfc6749#section-1.3.4
[OAuth 2.0 Refresh Token]: https://datatracker.ietf.org/doc/html/rfc6749#section-1.5
[OAuth 2.0 Device Code]: https://datatracker.ietf.org/doc/html/rfc8628#section-3.4

### Response Types

The following describes the supported response types. See the [OAuth 2.0 Multiple Response Type Encoding Practices] for
more technical information.

|         Flow Type         |        Values         |
|:-------------------------:|:---------------------:|
| [Authorization Code Flow] |        `code`         |
|      [Implicit Flow]      |   `token id_token`    |
|      [Implicit Flow]      |      `id_token`       |
|      [Implicit Flow]      |        `token`        |
|       [Hybrid Flow]       |     `code token`      |
|       [Hybrid Flow]       |    `code id_token`    |
|       [Hybrid Flow]       | `code token id_token` |

[Authorization Code Flow]: https://openid.net/specs/openid-connect-core-1_0.html#CodeFlowAuth
[Implicit Flow]: https://openid.net/specs/openid-connect-core-1_0.html#ImplicitFlowAuth
[Hybrid Flow]: https://openid.net/specs/openid-connect-core-1_0.html#HybridFlowAuth

[OAuth 2.0 Multiple Response Type Encoding Practices]: https://openid.net/specs/oauth-v2-multiple-response-types-1_0.html

### Response Modes

The following describes the supported response modes. See the [OAuth 2.0 Multiple Response Type Encoding Practices] for
more technical information.

|         Name          |    Value    |
|:---------------------:|:-----------:|
|     Query String      |   `query`   |
|       Fragment        | `fragment`  |
| [OAuth 2.0 Form Post] | `form_post` |

[OAuth 2.0 Form Post]: https://openid.net/specs/oauth-v2-form-post-response-mode-1_0.html

## Authentication Method References

Authelia currently supports adding the `amr` [Claim] to the [ID Token] utilizing the [RFC8176] Authentication Method
Reference values.

The values this [Claim] has are not strictly defined by the [OpenID Connect 1.0] specification. As such, some backends may
expect a specification other than [RFC8176] for this purpose. If you have such an application and wish for us to support
it then you're encouraged to create a [feature request](https://www.authelia.com/l/fr).

Below is a list of the potential values we place in the [Claim] and their meaning:

| Value |                           Description                            | Factor | Channel  |
|:-----:|:----------------------------------------------------------------:|:------:|:--------:|
|  mfa  |     User used multiple factors to login (see factor column)      |  N/A   |   N/A    |
|  mca  |    User used multiple channels to login (see channel column)     |  N/A   |   N/A    |
| user  |  User confirmed they were present when using their hardware key  |  N/A   |   N/A    |
|  pin  | User confirmed they are the owner of the hardware key with a pin |  N/A   |   N/A    |
|  pwd  |            User used a username and password to login            |  Know  | Browser  |
|  otp  |                     User used TOTP to login                      |  Have  | Browser  |
|  hwk  |                User used a hardware key to login                 |  Have  | Browser  |
|  sms  |                      User used Duo to login                      |  Have  | External |

## User Information Signing Algorithm

The following table describes the response from the [UserInfo] endpoint depending on the
[userinfo_signing_algorithm](../../configuration/identity-providers/open-id-connect.md#userinfosigningalgorithm).

| Signing Algorithm |   Encoding   |            Content Type             |
|:-----------------:|:------------:|:-----------------------------------:|
|      `none`       |     JSON     | `application/json; charset="UTF-8"` |
|      `RS256`      | JWT (Signed) | `application/jwt; charset="UTF-8"`  |

## Endpoint Implementations

The following section documents the endpoints we implement and their respective paths. This information can
traditionally be discovered by relying parties that utilize [OpenID Connect Discovery], however this information may be
useful for clients which do not implement this.

The endpoints can be discovered easily by visiting the Discovery and Metadata endpoints. It is recommended regardless
of your version of Authelia that you utilize this version as it will always produce the correct endpoint URLs. The paths
for the Discovery/Metadata endpoints are part of IANA's well known registration but are also documented in a table
below.

These tables document the endpoints we currently support and their paths in the most recent version of Authelia. The
paths are appended to the end of the primary URL used to access Authelia. The tables use the url
https://auth.example.com as an example of the Authelia root URL which is also the OpenID Connect issuer.

### Well Known Discovery Endpoints

These endpoints can be utilized to discover other endpoints and metadata about the Authelia OP.

|                 Endpoint                  |                              Path                               |
|:-----------------------------------------:|:---------------------------------------------------------------:|
|        [OpenID Connect Discovery]         |    https://auth.example.com/.well-known/openid-configuration    |
| [OAuth 2.0 Authorization Server Metadata] | https://auth.example.com/.well-known/oauth-authorization-server |

### Discoverable Endpoints

These endpoints implement OpenID Connect elements.

|            Endpoint             |                         Path                         |          Discovery Attribute          |
|:-------------------------------:|:----------------------------------------------------:|:-------------------------------------:|
|       [JSON Web Key Set]        |          https://auth.example.com/jwks.json          |               jwks_uri                |
|         [Authorization]         |   https://auth.example.com/api/oidc/authorization    |        authorization_endpoint         |
| [Pushed Authorization Requests] | https://auth.example.com/api/oidc/authorization/push | pushed_authorization_request_endpoint |
|             [Token]             |       https://auth.example.com/api/oidc/token        |            token_endpoint             |
|           [UserInfo]            |      https://auth.example.com/api/oidc/userinfo      |           userinfo_endpoint           |
|         [Introspection]         |   https://auth.example.com/api/oidc/introspection    |        introspection_endpoint         |
|          [Revocation]           |     https://auth.example.com/api/oidc/revocation     |          revocation_endpoint          |

## Security

The following information covers some security topics some users may wish to be familiar with.

#### Pushed Authorization Requests Endpoint

The [Pushed Authorization Requests] endpoint is discussed in depth in [RFC9126] as well as in the
[OAuth 2.0 Pushed Authorization Requests](https://oauth.net/2/pushed-authorization-requests/) documentation.

Essentially it's a special endpoint that takes the same parameters as the [Authorization] endpoint (including
[Proof Key Code Exchange](#proof-key-code-exchange)) with a few caveats:

1. The same [Client Authentication] mechanism required by the [Token] endpoint **MUST** be used.
2. The request **MUST** use the [HTTP POST method].
3. The request **MUST** use the `application/x-www-form-urlencoded` content type (i.e. the parameters **MUST** be in the
   body, not the URI).
4. The request **MUST** occur over the back-channel.

The response of this endpoint is a JSON Object with two key-value pairs:
- `request_uri`
- `expires_in`

The `expires_in` indicates how long the `request_uri` is valid for. The `request_uri` is used as a parameter to the
[Authorization] endpoint instead of the standard parameters (as the `request_uri` parameter).

The advantages of this approach are as follows:

1. [Pushed Authorization Requests] cannot be created or influenced by any party other than the Relying Party (client).
2. Since you can force all [Authorization] requests to be initiated via [Pushed Authorization Requests] you drastically
   improve the authorization flows resistance to phishing attacks (this can be done globally or on a per-client basis).
3. Since the [Pushed Authorization Requests] endpoint requires all of the same [Client Authentication] mechanisms as the
   [Token] endpoint:
   1. Clients using the confidential [Client Type] can't have [Pushed Authorization Requests] generated by parties who do not
      have the credentials.
   2. Clients using the public [Client Type] and utilizing [Proof Key Code Exchange](#proof-key-code-exchange) never
      transmit the verifier over any front-channel making even the `plain` challenge method relatively secure.

#### Proof Key Code Exchange

The [Proof Key Code Exchange] mechanism is discussed in depth in [RFC7636] as well as in the
[OAuth 2.0 Proof Key Code Exchange](https://oauth.net/2/pkce/) documentation.

Essentially a random opaque value is generated by the Relying Party and optionally (but recommended) passed through a
SHA256 hash. The original value is saved by the Relying Party, and the hashed value is sent in the [Authorization]
request in the `code_verifier` parameter with the `code_challenge_method` set to `S256` (or `plain` using a bad practice
of not hashing the opaque value).

When the Relying Party requests the token from the [Token] endpoint, they must include the `code_verifier` parameter
again (in the body), but this time they send the value without it being hashed.

The advantages of this approach are as follows:

1. Provided the value was hashed it's certain that the Relying Party which generated the authorization request is the
   same party as the one requesting the token or is permitted by the Relying Party to make this request.
2. Even when using the public [Client Type] there is a form of authentication on the  [Token] endpoint.

[ID Token]: https://openid.net/specs/openid-connect-core-1_0.html#IDToken
[Access Token]: https://datatracker.ietf.org/doc/html/rfc6749#section-1.4
[Refresh Token]: https://openid.net/specs/openid-connect-core-1_0.html#RefreshTokens

[Claims]: https://openid.net/specs/openid-connect-core-1_0.html#Claims
[Claim]: https://openid.net/specs/openid-connect-core-1_0.html#Claims

[OpenID Connect 1.0]: https://openid.net/connect/

[OpenID Connect Discovery]: https://openid.net/specs/openid-connect-discovery-1_0.html
[OAuth 2.0 Authorization Server Metadata]: https://datatracker.ietf.org/doc/html/rfc8414

[JSON Web Key Set]: https://datatracker.ietf.org/doc/html/rfc7517#section-5

[Authorization]: https://openid.net/specs/openid-connect-core-1_0.html#AuthorizationEndpoint
[Pushed Authorization Requests]: https://datatracker.ietf.org/doc/html/rfc9126
[Token]: https://openid.net/specs/openid-connect-core-1_0.html#TokenEndpoint
[UserInfo]: https://openid.net/specs/openid-connect-core-1_0.html#UserInfo
[Introspection]: https://datatracker.ietf.org/doc/html/rfc7662
[Revocation]: https://datatracker.ietf.org/doc/html/rfc7009
[Proof Key Code Exchange]: https://www.rfc-editor.org/rfc/rfc7636.html

[Subject Identifier Types]: https://openid.net/specs/openid-connect-core-1_0.html#SubjectIDTypes
[Client Authentication]: https://datatracker.ietf.org/doc/html/rfc6749#section-2.3
[Client Type]: https://oauth.net/2/client-types/
[HTTP POST method]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/POST
[Proof Key Code Exchange]: #proof-key-code-exchange

[RFC4122]: https://datatracker.ietf.org/doc/html/rfc4122
[RFC7636]: https://datatracker.ietf.org/doc/html/rfc7636
[RFC8176]: https://datatracker.ietf.org/doc/html/rfc8176
[RFC9126]: https://datatracker.ietf.org/doc/html/rfc9126
