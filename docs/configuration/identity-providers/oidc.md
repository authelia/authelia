---
layout: default
title: OpenID Connect
parent: Identity Providers
grand_parent: Configuration
nav_order: 2
---

# OpenID Connect

**Authelia** currently supports the [OpenID Connect] OP role as a [**beta**](../../roadmap/oidc.md) feature. The OP role
is the [OpenID Connect] Provider role, not the Relying Party or RP role. This means other applications that implement the
[OpenID Connect] RP role can use Authelia as an authentication and authorization backend similar to how you may use
social media or development platforms for login.

The Relying Party role is the role which allows an application to use GitHub, Google, or other [OpenID Connect]
providers for authentication and authorization. We do not intend to support this functionality at this moment in time.

More information about the beta can be found in the [roadmap](../../roadmap/oidc.md).

## Configuration

The following snippet provides a sample-configuration for the OIDC identity provider explaining each field in detail.

```yaml
identity_providers:
  oidc:
    hmac_secret: this_is_a_secret_abc123abc123abc
    issuer_private_key: |
      --- KEY START
      --- KEY END
    access_token_lifespan: 1h
    authorize_code_lifespan: 1m
    id_token_lifespan: 1h
    refresh_token_lifespan: 90m
    enable_client_debug_messages: false
    enforce_pkce: public_clients_only
    cors:
      endpoints:
        - authorization
        - token
        - revocation
        - introspection
      allowed_origins:
        - https://example.com
      allowed_origins_from_client_redirect_uris: false
    clients:
      - id: myapp
        description: My Application
        secret: this_is_a_secret
        sector_identifier: ''
        public: false
        authorization_policy: two_factor
        audience: []
        scopes:
          - openid
          - groups
          - email
          - profile
        redirect_uris:
          - https://oidc.example.com:8080/oauth2/callback
        grant_types:
          - refresh_token
          - authorization_code
        response_types:
          - code
        response_modes:
          - form_post
          - query
          - fragment
        userinfo_signing_algorithm: none
```

## Options

### hmac_secret
<div markdown="1">
type: string
{: .label .label-config .label-purple } 
required: yes
{: .label .label-config .label-red }
</div>

The HMAC secret used to sign the [OpenID Connect] JWT's. The provided string is hashed to a SHA256
byte string for the purpose of meeting the required format. You must [generate this option yourself](#generating-a-random-secret).

Should be defined using a [secret](../secrets.md) which is the recommended for containerized deployments.

### issuer_private_key
<div markdown="1">
type: string
{: .label .label-config .label-purple }
required: yes
{: .label .label-config .label-red }
</div>

The private key in DER base64 encoded PEM format used to encrypt the [OpenID Connect] JWT's.[¹](../../faq.md#why-only-use-a-private-issuer-key-and-no-public-key-with-oidc)
You must [generate this option yourself](#generating-a-random-secret). To create this option, use
`docker run -u "$(id -u):$(id -g)" -v "$(pwd)":/keys authelia/authelia:latest authelia rsa generate --dir /keys`
to generate both the private and public key in the current directory. You can then paste the
private key into your configuration.

Should be defined using a [secret](../secrets.md) which is the recommended for containerized deployments.

### access_token_lifespan
<div markdown="1">
type: duration
{: .label .label-config .label-purple }
default: 1h
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

The maximum lifetime of an access token. It's generally recommended keeping this short similar to the default.
For more information read these docs about [token lifespan].

### authorize_code_lifespan
<div markdown="1">
type: duration
{: .label .label-config .label-purple }
default: 1m
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

The maximum lifetime of an authorize code. This can be rather short, as the authorize code should only be needed to
obtain the other token types. For more information read these docs about [token lifespan].

### id_token_lifespan
<div markdown="1">
type: duration
{: .label .label-config .label-purple }
default: 1h
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

The maximum lifetime of an ID token. For more information read these docs about [token lifespan].

### refresh_token_lifespan
<div markdown="1">
type: string
{: .label .label-config .label-purple }
default: 90m
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

The maximum lifetime of a refresh token. The
refresh token can be used to obtain new refresh tokens as well as access tokens or id tokens with an
up-to-date expiration. For more information read these docs about [token lifespan].

A good starting point is 50% more or 30 minutes more (which ever is less) time than the highest lifespan out of the
[access token lifespan](#access_token_lifespan), the [authorize code lifespan](#authorize_code_lifespan), and the
[id token lifespan](#id_token_lifespan). For instance the default for all of these is 60 minutes, so the default refresh
token lifespan is 90 minutes.

### enable_client_debug_messages
<div markdown="1">
type: boolean
{: .label .label-config .label-purple }
default: false
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Allows additional debug messages to be sent to the clients.

### minimum_parameter_entropy
<div markdown="1">
type: integer
{: .label .label-config .label-purple }
default: 8
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

This controls the minimum length of the `nonce` and `state` parameters.

***Security Notice:*** Changing this value is generally discouraged, reducing it from the default can theoretically
make certain scenarios less secure. It is highly encouraged that if your OpenID Connect RP does not send these parameters
or sends parameters with a lower length than the default that they implement a change rather than changing this value.

### enforce_pkce
<div markdown="1">
type: string
{: .label .label-config .label-purple } 
default: public_clients_only
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

[Proof Key for Code Exchange](https://datatracker.ietf.org/doc/html/rfc7636) enforcement policy: if specified, must be either `never`, `public_clients_only` or `always`.

If set to `public_clients_only` (default), PKCE will be required for public clients using the Authorization Code flow.

When set to `always`, PKCE will be required for all clients using the Authorization Code flow.

***Security Notice:*** Changing this value to `never` is generally discouraged, reducing it from the default can theoretically
make certain client-side applications (mobile applications, SPA) vulnerable to CSRF and authorization code interception attacks.

### enable_pkce_plain_challenge
<div markdown="1">
type: boolean
{: .label .label-config .label-purple }
default: false
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Allows PKCE `plain` challenges when set to `true`.

***Security Notice:*** Changing this value is generally discouraged. Applications should use the `S256` PKCE challenge method instead.

### cors

Some OpenID Connect Endpoints need to allow cross-origin resource sharing, however some are optional. This section allows
you to configure the optional parts. We reply with CORS headers when the request includes the Origin header.

#### endpoints
<div markdown="1">
type: list(string)
{: .label .label-config .label-purple }
default: empty
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

A list of endpoints to configure with cross-origin resource sharing headers. It is recommended that the `userinfo`
option is at least in this list. The potential endpoints which this can be enabled on are as follows:

* authorization
* token
* revocation
* introspection
* userinfo

#### allowed_origins
<div markdown="1">
type: list(string)
{: .label .label-config .label-purple }
default: empty
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

A list of permitted origins.

Any origin with https is permitted unless this option is configured or the allowed_origins_from_client_redirect_uris
option is enabled. This means you must configure this option manually if you want http endpoints to be permitted to
make cross-origin requests to the OpenID Connect endpoints, however this is not recommended.

Origins must only have the scheme, hostname and port, they may not have a trailing slash or path.

In addition to an Origin URI, you may specify the wildcard origin in the allowed_origins. It MUST be specified by itself
and the allowed_origins_from_client_redirect_uris MUST NOT be enabled. The wildcard origin is denoted as `*`. Examples:

```yaml
identity_providers:
  oidc:
    cors:
      allowed_origins: "*"
```

```yaml
identity_providers:
  oidc:
    cors:
      allowed_origins: 
        - "*"
```

#### allowed_origins_from_client_redirect_uris
<div markdown="1">
type: boolean
{: .label .label-config .label-purple }
default: false
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Automatically adds the origin portion of all redirect URI's on all clients to the list of allowed_origins, provided they
have the scheme http or https and do not have the hostname of localhost.

### clients

A list of clients to configure. The options for each client are described below.

#### id
<div markdown="1">
type: string
{: .label .label-config .label-purple } 
required: yes
{: .label .label-config .label-red }
</div>

The Client ID for this client. It must exactly match the Client ID configured in the application
consuming this client.

#### description
<div markdown="1">
type: string
{: .label .label-config .label-purple } 
default: *same as id*
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

A friendly description for this client shown in the UI. This defaults to the same as the ID.

#### secret
<div markdown="1">
type: string
{: .label .label-config .label-purple }
required: situational
{: .label .label-config .label-yellow }
</div>

The shared secret between Authelia and the application consuming this client. This secret must
match the secret configured in the application. Currently this is stored in plain text.
You must [generate this option yourself](#generating-a-random-secret).

This must be provided when the client is a confidential client type, and must be blank when using the public client
type. To set the client type to public see the [public](#public) configuration option.

#### sector_identifier
<div markdown="1">
type: string
{: .label .label-config .label-purple }
default: ''
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-red }
</div>

_**Important Note:** because adjusting this option will inevitably change the `sub` claim of all tokens generated for
the specified client, changing this should cause the relying party to detect all future authorizations as completely new
users._

Must be an empty string or the host component of a URL. This is commonly just the domain name, but may also include a
port.

Authelia utilizes UUID version 4 subject identifiers. By default the public subject identifier type is utilized for all
clients. This means the subject identifiers will be the same for all clients. This configuration option enables pairwise
for this client, and configures the sector identifier utilized for both the storage and the lookup of the subject
identifier.

1. All clients who do not have this configured will generate the same subject identifier for a particular user regardless
   of which client obtains the ID token.
2. All clients which have the same sector identifier will:
   1. have the same subject identifier for a particular user when compared to clients with the same sector identifier.
   2. have a completely different subject identifier for a particular user whe compared to:
      1. any client with the public subject identifier type.
      2. any client with a differing sector identifier.

In specific but limited scenarios this option is beneficial for privacy reasons. In particular this is useful when the
party utilizing the _Authelia_ [OpenID Connect] Authorization Server is foreign and not controlled by the user. It would
prevent the third party utilizing the subject identifier with another third party in order to track the user.

Keep in mind depending on the other claims they may still be able to perform this tracking and it is not a silver bullet.
There are very few benefits when utilizing this in a homelab or business where no third party is utilizing
the server.

#### public
<div markdown="1">
type: bool
{: .label .label-config .label-purple }
default: false
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

This enables the public client type for this client. This is for clients that are not capable of maintaining
confidentiality of credentials, you can read more about client types in [RFC6749](https://datatracker.ietf.org/doc/html/rfc6749#section-2.1).
This is particularly useful for SPA's and CLI tools. This option requires setting the [client secret](#secret) to a
blank string.

In addition to the standard rules for redirect URIs, public clients can use the `urn:ietf:wg:oauth:2.0:oob` redirect URI.

#### authorization_policy
<div markdown="1">
type: string
{: .label .label-config .label-purple } 
default: two_factor
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

The authorization policy for this client: either `one_factor` or `two_factor`.

#### audience
<div markdown="1">
type: list(string)
{: .label .label-config .label-purple } 
required: no
{: .label .label-config .label-green }
</div>

A list of audiences this client is allowed to request.

#### scopes
<div markdown="1">
type: list(string)
{: .label .label-config .label-purple }
default: openid, groups, profile, email
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

A list of scopes to allow this client to consume. See [scope definitions](#scope-definitions) for more
information. The documentation for the application you want to use with Authelia will most-likely provide
you with the scopes to allow.

#### redirect_uris
<div markdown="1">
type: list(string)
{: .label .label-config .label-purple }
required: yes
{: .label .label-config .label-red }
</div>

A list of valid callback URIs this client will redirect to. All other callbacks will be considered
unsafe. The URIs are case-sensitive and they differ from application to application - the community has
provided [a list of URL´s for common applications](../../community/oidc-integrations.md).

Some restrictions that have been placed on clients and
their redirect URIs are as follows:

1. If a client attempts to authorize with Authelia and its redirect URI is not listed in the client configuration the
   attempt to authorize wil fail and an error will be generated.
2. The redirect URIs are case-sensitive.
3. The URI must include a scheme and that scheme must be one of `http` or `https`.
4. The client can ignore rule 3 and use `urn:ietf:wg:oauth:2.0:oob` if it is a [public](#public) client type.

#### grant_types
<div markdown="1">
type: list(string)
{: .label .label-config .label-purple } 
default: refresh_token, authorization_code
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

A list of grant types this client can return. _It is recommended that this isn't configured at this time unless you
know what you're doing_. Valid options are: `implicit`, `refresh_token`, `authorization_code`, `password`,
`client_credentials`.

#### response_types
<div markdown="1">
type: list(string)
{: .label .label-config .label-purple } 
default: code
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

A list of response types this client can return. _It is recommended that this isn't configured at this time unless you
know what you're doing_. Valid options are: `code`, `code id_token`, `id_token`, `token id_token`, `token`,
`token id_token code`.

#### response_modes
<div markdown="1">
type: list(string)
{: .label .label-config .label-purple } 
default: form_post, query, fragment
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

A list of response modes this client can return. It is recommended that this isn't configured at this time unless you
know what you're doing. Potential values are `form_post`, `query`, and `fragment`.

#### userinfo_signing_algorithm
<div markdown="1">
type: string
{: .label .label-config .label-purple } 
default: none
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

The algorithm used to sign the userinfo endpoint responses. This can either be `none` or `RS256`.

## Generating a random secret

If you must provide a random secret in configuration, you can generate a random string of sufficient length. The command

```sh
LENGTH=64
tr -cd '[:alnum:]' < /dev/urandom | fold -w "${LENGTH}" | head -n 1 | tr -d '\n' ; echo
```

prints such a string with a length in characters of `${LENGTH}` on `stdout`. The string will only contain alphanumeric
characters. For Kubernetes, see [this section too](../secrets.md#Kubernetes).

## Scope Definitions

### openid

This is the default scope for openid. This field is forced on every client by the configuration validation that Authelia
does.

_**Important Note:** The subject identifiers or `sub` claim has been changed to a [RFC4122] UUID V4 to identify the 
individual user as per the [Subject Identifier Types] specification. Please use the claim `preferred_username` instead._

|   Claim   |   JWT Type    | Authelia Attribute |                         Description                         |
|:---------:|:-------------:|:------------------:|:-----------------------------------------------------------:|
|    sub    |    string     |      username      |    A [RFC4122] UUID V4 linked to the user who logged in     |
|   scope   |    string     |       scopes       |              Granted scopes (space delimited)               |
|    scp    | array[string] |       scopes       |                       Granted scopes                        |
|    iss    |    string     |      hostname      |             The issuer name, determined by URL              |
|  at_hash  |    string     |       _N/A_        |                      Access Token Hash                      |
|    aud    | array[string] |       _N/A_        |                          Audience                           |
|    exp    |    number     |       _N/A_        |                           Expires                           |
| auth_time |    number     |       _N/A_        |        The time the user authenticated with Authelia        |
|    rat    |    number     |       _N/A_        |            The time when the token was requested            |
|    iat    |    number     |       _N/A_        |             The time when the token was issued              |
|    jti    | string(uuid)  |       _N/A_        |     A JWT Identifier in the form of a [RFC4122] UUID V4     |
|    amr    | array[string] |       _N/A_        | An [RFC8176] list of authentication method reference values |

### groups

This scope includes the groups the authentication backend reports the user is a member of in the token.

| Claim  |   JWT Type    | Authelia Attribute |                                                    Description                                                     |
|:------:|:-------------:|:------------------:|:------------------------------------------------------------------------------------------------------------------:|
| groups | array[string] |       groups       | List of user's groups discovered via [authentication](https://www.authelia.com/docs/configuration/authentication/) |

### email

This scope includes the email information the authentication backend reports about the user in the token.

|     Claim      |   JWT Type    | Authelia Attribute |                        Description                        |
|:--------------:|:-------------:|:------------------:|:---------------------------------------------------------:|
|     email      |    string     |      email[0]      |       The first email address in the list of emails       |
| email_verified |     bool      |       _N/A_        | If the email is verified, assumed true for the time being |
|   alt_emails   | array[string] |     email[1:]      |  All email addresses that are not in the email JWT field  |

### profile

This scope includes the profile information the authentication backend reports about the user in the token.

|       Claim        | JWT Type | Authelia Attribute |               Description                |
|:------------------:|:--------:|:------------------:|:----------------------------------------:|
| preferred_username |  string  |      username      | The username the user used to login with |
|        name        |  string  |    display_name    |          The users display name          |

## Authentication Method References

Authelia currently supports adding the `amr` claim to the [ID Token](https://openid.net/specs/openid-connect-core-1_0.html#IDToken)
utilizing the [RFC8176] Authentication Method Reference values. 

The values this claim has are not strictly defined by the [OpenID Connect] specification. As such, some backends may
expect a specification other than [RFC8176] for this purpose. If you have such an application and wish for us to support
it then you're encouraged to create an issue.

Below is a list of the potential values we place in the claim and their meaning:

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

The following section documents the endpoints we implement and their respective paths. This information can traditionally
be discovered by relying parties that utilize [discovery](https://openid.net/specs/openid-connect-discovery-1_0.html),
however this information may be useful for clients which do not implement this.

The endpoints can be discovered easily by visiting the Discovery and Metadata endpoints. It is recommended regardless
of your version of Authelia that you utilize this version as it will always produce the correct endpoint URLs. The paths
for the Discovery/Metadata endpoints are part of IANA's well known registration but are also documented in a table below.

These tables document the endpoints we currently support and their paths in the most recent version of Authelia. The paths
are appended to the end of the primary URL used to access Authelia. The tables use the url https://auth.example.com as 
an example of the Authelia root URL which is also the OpenID Connect issuer.

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

[JSON Web Key Sets]: https://datatracker.ietf.org/doc/html/rfc7517#section-5
[OpenID Connect]: https://openid.net/connect/
[Subject Identifier Types]:https://openid.net/specs/openid-connect-core-1_0.html#SubjectIDTypes
[OpenID Connect Discovery]: https://openid.net/specs/openid-connect-discovery-1_0.html
[OAuth 2.0 Authorization Server Metadata]: https://datatracker.ietf.org/doc/html/rfc8414
[Authorization]: https://openid.net/specs/openid-connect-core-1_0.html#AuthorizationEndpoint
[Token]: https://openid.net/specs/openid-connect-core-1_0.html#TokenEndpoint
[Userinfo]: https://openid.net/specs/openid-connect-core-1_0.html#UserInfo
[Introspection]: https://datatracker.ietf.org/doc/html/rfc7662
[Revocation]: https://datatracker.ietf.org/doc/html/rfc7009
[RFC8176]: https://datatracker.ietf.org/doc/html/rfc8176
[RFC4122]: https://datatracker.ietf.org/doc/html/rfc4122
[token lifespan]: https://docs.apigee.com/api-platform/antipatterns/oauth-long-expiration