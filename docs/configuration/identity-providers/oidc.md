---
layout: default
title: OpenID Connect
parent: Identity Providers
grand_parent: Configuration
nav_order: 2
---

# OpenID Connect

**Authelia** currently supports the [OpenID Connect] OP role as a [**beta**](#roadmap) feature. The OP role is the
[OpenID Connect] Provider role, not the Relying Party or RP role. This means other applications that implement the
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
    clients:
      - id: myapp
        description: My Application
        secret: this_is_a_secret
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

_**Important Note:** The claim `sub` is planned to be changed in the future to a randomly unique value to identify the
individual user. Please use the claim `preferred_username` instead._

|       Claim        |   JWT Type    | Authelia Attribute |                  Description                  |
|:------------------:|:-------------:|:------------------:|:---------------------------------------------:|
|        sub         |    string     |      username      |   The username the user used to login with    |
|       scope        |    string     |       scopes       |       Granted scopes (space delimited)        |
|        scp         | array[string] |       scopes       |                Granted scopes                 |
|        iss         |    string     |      hostname      |      The issuer name, determined by URL       |
|      at_hash       |    string     |       _N/A_        |               Access Token Hash               |
|        aud         | array[string] |       _N/A_        |                   Audience                    |
|        exp         |    number     |       _N/A_        |                    Expires                    |
|     auth_time      |    number     |       _N/A_        | The time the user authenticated with Authelia |
|        rat         |    number     |       _N/A_        |     The time when the token was requested     |
|        iat         |    number     |       _N/A_        |      The time when the token was issued       |
|        jti         | string(uuid)  |       _N/A_        |                JWT Identifier                 |

### groups

This scope includes the groups the authentication backend reports the user is a member of in the token.

| Claim  |   JWT Type    | Authelia Attribute |      Description       |
|:------:|:-------------:|:------------------:|:----------------------:|
| groups | array[string] |       groups       | The users display name |

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

## Endpoint Implementations

This is a table of the endpoints we currently support and their paths. This can be requrired information for some RP's,
particularly those that don't use [discovery](https://openid.net/specs/openid-connect-discovery-1_0.html). The paths are
appended to the end of the primary URL used to access Authelia. For example in the Discovery example provided you access
Authelia via https://auth.example.com, the discovery URL is https://auth.example.com/.well-known/openid-configuration.

|   Endpoint    |                     Path                      |
|:-------------:|:---------------------------------------------:|
|   Discovery   |    [root]/.well-known/openid-configuration    |
|   Metadata    | [root]/.well-known/oauth-authorization-server |
|     JWKS      |             [root]/api/oidc/jwks              |
| Authorization |         [root]/api/oidc/authorization         |
|     Token     |             [root]/api/oidc/token             |
| Introspection |         [root]/api/oidc/introspection         |
|  Revocation   |          [root]/api/oidc/revocation           |
|   Userinfo    |           [root]/api/oidc/userinfo            |

[OpenID Connect]: https://openid.net/connect/
[token lifespan]: https://docs.apigee.com/api-platform/antipatterns/oauth-long-expiration
