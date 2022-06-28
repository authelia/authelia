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

Authelia supports [OpenID Connect] as part of an open beta. This section details implementation specifics that can be
used for integrating Authelia with relying parties, as well as specific documentation for some relying parties.

See the [configuration documentation](../../configuration/identity-providers/open-id-connect.md) for information on how
to configure [OpenID Connect].

## Scope Definitions

### openid

This is the default scope for [OpenID Connect]. This field is forced on every client by the configuration validation
that Authelia does.

*__Important Note:__ The subject identifiers or `sub` [Claim] has been changed to a [RFC4122] UUID V4 to identify the
individual user as per the [Subject Identifier Types] section of the [OpenID Connect] specification. Please use the
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

## Authentication Method References

Authelia currently supports adding the `amr` [Claim] to the [ID Token] utilizing the [RFC8176] Authentication Method
Reference values.

The values this [Claim] has are not strictly defined by the [OpenID Connect] specification. As such, some backends may
expect a specification other than [RFC8176] for this purpose. If you have such an application and wish for us to support
it then you're encouraged to create an issue.

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

|      Endpoint       |                      Path                       |  Discovery Attribute   |
|:-------------------:|:-----------------------------------------------:|:----------------------:|
| [JSON Web Key Sets] |       https://auth.example.com/jwks.json        |        jwks_uri        |
|   [Authorization]   | https://auth.example.com/api/oidc/authorization | authorization_endpoint |
|       [Token]       |     https://auth.example.com/api/oidc/token     |     token_endpoint     |
|     [Userinfo]      |   https://auth.example.com/api/oidc/userinfo    |   userinfo_endpoint    |
|   [Introspection]   | https://auth.example.com/api/oidc/introspection | introspection_endpoint |
|    [Revocation]     |  https://auth.example.com/api/oidc/revocation   |  revocation_endpoint   |

[ID Token]: https://openid.net/specs/openid-connect-core-1_0.html#IDToken
[Access Token]: https://datatracker.ietf.org/doc/html/rfc6749#section-1.4
[Refresh Token]: https://openid.net/specs/openid-connect-core-1_0.html#RefreshTokens

[Claims]: https://openid.net/specs/openid-connect-core-1_0.html#Claims
[Claim]: https://openid.net/specs/openid-connect-core-1_0.html#Claims

[OpenID Connect]: https://openid.net/connect/

[OpenID Connect Discovery]: https://openid.net/specs/openid-connect-discovery-1_0.html
[OAuth 2.0 Authorization Server Metadata]: https://www.rfc-editor.org/rfc/rfc8414.html

[JSON Web Key Sets]: https://www.rfc-editor.org/rfc/rfc7517.html#section-5

[Authorization]: https://openid.net/specs/openid-connect-core-1_0.html#AuthorizationEndpoint
[Token]: https://openid.net/specs/openid-connect-core-1_0.html#TokenEndpoint
[Userinfo]: https://openid.net/specs/openid-connect-core-1_0.html#UserInfo
[Introspection]: https://www.rfc-editor.org/rfc/rfc7662.html
[Revocation]: https://www.rfc-editor.org/rfc/rfc7009.html

[RFC8176]: https://www.rfc-editor.org/rfc/rfc8176.html
[RFC4122]: https://www.rfc-editor.org/rfc/rfc4122.html
[Subject Identifier Types]: https://openid.net/specs/openid-connect-core-1_0.html#SubjectIDTypes
