---
title: "OpenID Connect 1.0 Claims"
description: "An introduction into utilizing the Authelia OpenID Connect 1.0 Claims functionality"
summary: "An introduction into utilizing the Authelia OpenID Connect 1.0 Claims functionality."
date: 2024-03-05T19:11:16+10:00
draft: false
images: []
weight: 611
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Claims

The [OAuth 2.0] and [OpenID Connect 1.0] effectively names the individual content of a token as a [Claim]. Each [Claim]
can either be granted individually via:

1. The requested and granted [Scope] which generally makes the associated [Claim] values available at the [UserInfo]
   endpoint, **_not_** within the [ID Token].
2. The [Claims Parameter] which can request the authorization server explicitly
   include a [Claim] in the [ID Token] and/or [UserInfo] endpoint.

Authelia supports several claims related features:

1. The availability of the [Standard Claims] using the appropriate [scopes](#scope-definitions):
   * These claims come from the standard attributes available in the authentication backends, usually with the attribute
   of the same name.
2. The availability of creating your own [custom claims](#custom-claims) and [custom scopes](#custom-scopes).
3. The ability to create [Custom Attributes] to bolster the [custom claims](#custom-claims) functionality.
4. The ability to request individual claims by clients with the [Claims Parameter] assuming the client is allowed to
   request the specified claim.

Because Authelia supports the [Claims Parameter] the [ID Token] only returns claims in a way that is privacy and
security focused, resulting in a minimal [ID Token] which solely proves authorization occurred. This allows various
flows which require the relying party to share the [ID Token] with another party to prove they have been authorized, and
relies on the [Access Token] which should be kept completely private to request the additional granted claims from the
[UserInfo] endpoint.

The [Scope Definitions] indicate the default locations for a specific claim depending on the granted [Scope] when the
[Claims Parameter] is not used and the default behaviour is not overridden by the registered client configuration.

[Scope Definitions]: #scope-definitions
[Scope]: https://openid.net/specs/openid-connect-core-1_0.html#ScopeClaims
[Claims Parameter]: https://openid.net/specs/openid-connect-core-1_0.html#ClaimsParameter

### Authentication Method References

Authelia currently supports adding the `amr` [Claim] to the [ID Token] utilizing the [RFC8176] Authentication Method
Reference values.

The values this [Claim] has, are not strictly defined by the [OpenID Connect 1.0] specification. As such, some backends
may
expect a specification other than [RFC8176] for this purpose. If you have such an application and wish for us to support
it then you're encouraged to create a [feature request](https://www.authelia.com/l/fr).

Below is a list of the potential values we place in the [Claim] and their meaning:

| Value |                            Description                            | Factor | Channel  |
|:-----:|:-----------------------------------------------------------------:|:------:|:--------:|
|  mfa  |      User used multiple factors to login (see factor column)      |  N/A   |   N/A    |
|  mca  |     User used multiple channels to login (see channel column)     |  N/A   |   N/A    |
| user  |  User confirmed they were present when using their hardware key   |  N/A   |   N/A    |
|  pin  | User confirmed they are the owner of the hardware key with a pin  |  N/A   |   N/A    |
|  pwd  |            User used a username and password to login             |  Know  | Browser  |
|  otp  |                      User used TOTP to login                      |  Have  | Browser  |
|  pop  | User used a software or hardware proof-of-possession key to login |  Have  | Browser  |
|  hwk  |       User used a hardware proof-of-possession key to login       |  Have  | Browser  |
|  swk  |       User used a software proof-of-possession key to login       |  Have  | Browser  |
|  sms  |                      User used Duo to login                       |  Have  | External |

### Custom Claims

Authelia supports methods to define your own claims. These can either come from [Standard Attributes] or
[Custom Attributes]. These claims are delivered to provided via a claims policy which describes which claims are
available to specific scopes.

In the example below we configure a claims policy named `custom_claims_policy` which provides a claim named `claim_name`
which comes from the attribute named `attribute_name`. The claim `claim_name` is then made available via the scope named
`scope_name`.

This policy and scope is then configured within the client with client id `client_example_id`. The client can then
either request the claim via the `claims` parameter or via the `scope` parameter.

```yaml
identity_providers:
  oidc:
    claims_policies:
      custom_claims_policy:
        custom_claims:
          claim_name:
            attribute: 'attribute_name'
    scopes:
      scope_name:
        claims:
          - 'claim_name'
    clients:
      - client_id: 'client_example_id'
        claims_policy: 'custom_claims_policy'
        scopes:
          - 'scope_name'
```

### Restore Functionality Prior to Claims Parameter

The introduction of the claims parameter has removed most claims from the [ID Token]. This may not work for some clients
which do not make requests to the user information endpoint which contains all of these claims, and they may not support
the claims parameter.

The following example is a claims policy which restores that behaviour for those clients. Users may choose to expand
on this on their own as they desire. This example also shows how to apply this policy to a client using the
`claims_policy` option.

We strongly recommend implementers use the standard process for obtaining the extra claims not generally intended to be
included in the [ID Token] by using the [Access Token] to obtain them from the [UserInfo Endpoint]. This process is
considered significantly more stable and forms the basis for the future guarantee. This option only exists as a
break-glass measure and is only offered on a best-effort basis.

```yaml
identity_providers:
  oidc:
    claims_policies:
      default:
        id_token: ['groups', 'email', 'email_verified', 'alt_emails', 'preferred_username', 'name']
    clients:
      - client_id: 'client_example_id'
        claims_policy: 'default'
```

## Scope Definitions

The following scope definitions describe each scope supported and the associated effects including the individual claims
returned by granting this scope. By default, we do not issue any claims which reveal the users identity which allows
administrators semi-granular control over which claims the client is entitled to.

### openid

This is the default scope for [OpenID Connect 1.0]. This field is forced on every client by the configuration validation
that Authelia does.

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
The combination of the issuer (i.e. `iss`) [Claim](https://openid.net/specs/openid-connect-core-1_0.html#Claims) and
subject (i.e. `sub`) [Claim](https://openid.net/specs/openid-connect-core-1_0.html#Claims) are utilized to uniquely
identify a
user and per the specification the only reliable way to do so as they are guaranteed to be a unique combination. As such
this is the supported method for linking an account to Authelia. The `preferred_username` and `email` claims from the
`profile` and `email` scopes respectively should only be utilized for provisioning a new account.

In addition, the `sub` [Claim](https://openid.net/specs/openid-connect-core-1_0.html#Claims) utilizes
a [RFC4122](https://datatracker.ietf.org/doc/html/rfc4122) UUID V4 to identify the individual user as per the
[Subject Identifier Types](https://openid.net/specs/openid-connect-core-1_0.html#SubjectIDTypes) section of
the [OpenID Connect 1.0](https://openid.net/connect/) specification.
{{< /callout >}}

|  [Claim]  |   JWT Type    | Authelia Attribute | Default Location |                         Description                         |
|:---------:|:-------------:|:------------------:|:----------------:|:-----------------------------------------------------------:|
|    iss    |    string     |      hostname      |    [ID Token]    |             The issuer name, determined by URL              |
|    jti    | string(uuid)  |       *N/A*        |    [ID Token]    |     A [RFC4122] UUID V4 representing the JWT Identifier     |
|    sub    | string(uuid)  |     opaque id      |    [ID Token]    |    A [RFC4122] UUID V4 linked to the user who logged in     |
|    aud    | array[string] |       *N/A*        |    [ID Token]    |                          Audience                           |
|    exp    |    number     |       *N/A*        |    [ID Token]    |                           Expires                           |
|    iat    |    number     |       *N/A*        |    [ID Token]    |             The time when the token was issued              |
| auth_time |    number     |       *N/A*        |    [ID Token]    |        The time the user authenticated with Authelia        |
|    rat    |    number     |       *N/A*        |    [ID Token]    |            The time when the token was requested            |
|   nonce   |    string     |       *N/A*        |    [ID Token]    |        The time the user authenticated with Authelia        |
|    amr    | array[string] |       *N/A*        |    [ID Token]    | An [RFC8176] list of authentication method reference values |
|    azp    |    string     |    id (client)     |    [ID Token]    |                    The authorized party                     |
|   scope   |    string     |       scopes       |    [UserInfo]    |              Granted scopes (space delimited)               |
|    scp    | array[string] |       scopes       |    [UserInfo]    |                       Granted scopes                        |
| client_id |    string     |    id (client)     |    [UserInfo]    |                        The client id                        |

### offline_access

This scope is a special scope designed to allow applications to obtain a [Refresh Token] which allows extended access to
an application on behalf of a user. A [Refresh Token] is a special [Access Token] that allows refreshing previously
issued token credentials, effectively it allows the Relying Party to obtain new tokens periodically.

As per [OpenID Connect 1.0] Section 11 [Offline Access] can only be granted during the [Authorization Code Flow] or a
[Hybrid Flow]. The [Refresh Token] will only ever be returned at the [Token Endpoint] when the client is exchanging
their [OAuth 2.0 Authorization Code].

Generally unless an application supports this and actively requests this scope they should not be granted this scope via
the client configuration.

It is also important to note that we treat a [Refresh Token] as single use and reissue a new [Refresh Token] during the
refresh flow.

### profile

This scope allows the client to access the profile information the authentication backend reports about the user.

|      [Claim]       | JWT Type | Authelia Attribute | Default Location |               Description                |
|:------------------:|:--------:|:------------------:|:----------------:|:----------------------------------------:|
|        name        |  string  |    display_name    |    [UserInfo]    |          The users display name          |
|    family_name     |  string  |    family_name     |    [UserInfo]    |          The users family name           |
|     given_name     |  string  |     given_name     |    [UserInfo]    |           The users given name           |
|    	middle_name    |  string  |    	middle_name    |    [UserInfo]    |          The users middle name           |
|      nickname      |  string  |      nickname      |    [UserInfo]    |            The users nickname            |
| preferred_username |  string  |      username      |    [UserInfo]    | The username the user used to login with |
|      profile       |  string  |      profile       |    [UserInfo]    |          The users profile URL           |
|      picture       |  string  |      picture       |    [UserInfo]    |          The users picture URL           |
|      website       |  string  |      website       |    [UserInfo]    |          The users website URL           |
|       gender       |  string  |       gender       |    [UserInfo]    |             The users gender             |
|     birthdate      |  string  |     birthdate      |    [UserInfo]    |           The users birthdate            |
|      zoneinfo      |  string  |      zoneinfo      |    [UserInfo]    |            The users zoneinfo            |
|       locale       |  string  |       locale       |    [UserInfo]    |             The users locale             |

### email

This scope allows the client to access the email information the authentication backend reports about the user.

|    [Claim]     |   JWT Type    | Authelia Attribute | Default Location |                        Description                        |
|:--------------:|:-------------:|:------------------:|:----------------:|:---------------------------------------------------------:|
|     email      |    string     |      email[0]      |    [UserInfo]    |       The first email address in the list of emails       |
| email_verified |     bool      |       *N/A*        |    [UserInfo]    | If the email is verified, assumed true for the time being |
|   alt_emails   | array[string] |     email[1:]      |    [UserInfo]    |  All email addresses that are not in the email JWT field  |

### address

This scope allows the client to access the address information the authentication backend reports about the user. See
the [Address Claim](https://openid.net/specs/openid-connect-core-1_0.html#AddressClaim) definition for information on
the format of this claim.

| [Claim] | JWT Type | Authelia Attribute | Default Location |          Description          |
|:-------:|:--------:|:------------------:|:----------------:|:-----------------------------:|
| address |  object  |      various       |    [UserInfo]    | The users address information |

The following table indicates the various sub-claims within the address claim.

|    [Claim]     | JWT Type | Authelia Attribute |                       Description                       |
|:--------------:|:--------:|:------------------:|:-------------------------------------------------------:|
| street_address |  string  |   street_address   |                The users street address                 |
|    locality    |  string  |      locality      |             The users locality such as city             |
|     region     |  string  |       region       | The users region such as state, province, or prefecture |
|  postal_code   |  string  |    postal_code     |                   The users postcode                    |
|    country     |  string  |      country       |                    The users country                    |

### phone

This scope allows the client to access the address information the authentication backend reports about the user.

|        [Claim]        | JWT Type |       Authelia Attribute       | Default Location |                                              Description                                              |
|:---------------------:|:--------:|:------------------------------:|:----------------:|:-----------------------------------------------------------------------------------------------------:|
|     phone_number      |  string  | phone_number + phone_extension |    [UserInfo]    | The combination of the users phone number and extension in the format specified in OpenID Connect 1.0 |
| phone_number_verified | boolean  |              N/A               |    [UserInfo]    |                        Currently returns true if the phone number has a value.                        |

### groups

This scope includes the groups the authentication backend reports the user is a member of in the [Claims] of the
[ID Token].

| [Claim] |   JWT Type    | Authelia Attribute | Default Location |                                               Description                                               |
|:-------:|:-------------:|:------------------:|:----------------:|:-------------------------------------------------------------------------------------------------------:|
| groups  | array[string] |       groups       |    [UserInfo]    | List of user's groups discovered via [authentication](../../configuration/first-factor/introduction.md) |

### Special Scopes

The following scopes represent special permissions granted to a specific token.

#### authelia.bearer.authz

This scope allows the granted access token to be utilized with the bearer authorization scheme on endpoints protected
via Authelia.

The specifics about this scope are discussed in the
[OAuth 2.0 Bearer Token Usage for Authorization Endpoints](oauth-2.0-bearer-token-usage.md#authorization-endpoints)
guide.

[OAuth 2.0]: https://oauth.net/2/
[OpenID Connect 1.0]: https://openid.net/connect/

[ID Token]: https://openid.net/specs/openid-connect-core-1_0.html#IDToken
[Access Token]: https://datatracker.ietf.org/doc/html/rfc6749#section-1.4
[Refresh Token]: https://openid.net/specs/openid-connect-core-1_0.html#RefreshTokens

[Claims]: https://openid.net/specs/openid-connect-core-1_0.html#Claims
[Standard Claims]: https://openid.net/specs/openid-connect-core-1_0.html#StandardClaims
[Claim]: https://openid.net/specs/openid-connect-core-1_0.html#Claims
[Offline Access]: https://openid.net/specs/openid-connect-core-1_0.html#OfflineAccess
[UserInfo Endpoint]: https://openid.net/specs/openid-connect-core-1_0.html#UserInfo

[Standard Attributes]: ../../reference/guides/attributes.md#standard-attributes
[Custom Attributes]: ../../reference/guides/attributes.md#custom-attributes

[Authorization Code Flow]: https://openid.net/specs/openid-connect-core-1_0.html#CodeFlowAuth
[Hybrid Flow]: https://openid.net/specs/openid-connect-core-1_0.html#HybridFlowAuth

[Token Endpoint]: https://openid.net/specs/openid-connect-core-1_0.html#TokenEndpoint
[UserInfo]: https://openid.net/specs/openid-connect-core-1_0.html#UserInfo

[RFC4122]: https://datatracker.ietf.org/doc/html/rfc4122
[RFC8176]: https://datatracker.ietf.org/doc/html/rfc8176

[OAuth 2.0 Authorization Code]: https://datatracker.ietf.org/doc/html/rfc6749#section-1.3.1
