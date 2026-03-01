---
title: "OpenID Connect 1.0 Provider"
description: "OpenID Connect 1.0 Provider Configuration"
summary: "Authelia can operate as an OpenID Connect 1.0 Provider. This section describes how to configure this."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 110200
toc: true
aliases:
  - /c/oidc
  - /c/oidc/provider
  - /docs/configuration/identity-providers/oidc.html
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

__Authelia__ currently supports the [OpenID Connect 1.0] Provider role as an open
[__beta__](../../../roadmap/active/openid-connect-1.0-provider.md) feature. We currently do not support the [OpenID Connect 1.0] Relying
Party role. This means other applications that implement the [OpenID Connect 1.0] Relying Party role can use Authelia as
an [OpenID Connect 1.0] Provider similar to how you may use social media or development platforms for login.

The [OpenID Connect 1.0] Relying Party role is the role which allows an application to use GitHub, Google, or other
[OpenID Connect 1.0] Providers for authentication and authorization. We do not intend to support this functionality at
this moment in time.

This section covers the [OpenID Connect 1.0] Provider configuration. For information on configuring individual
registered clients see the [OpenID Connect 1.0 Clients](clients.md) documentation.

More information about the beta can be found in the [roadmap](../../../roadmap/active/openid-connect.md) and in the
[integration](../../../integration/openid-connect/introduction.md) documentation.

## OpenID Certified™

Authelia is [OpenID Certified™] to conform to the [OpenID Connect™ protocol].

{{< figure src="/images/oid-certification.jpg" class="center" sizes="40dvh" >}}

For more information please see the
[OpenID Connect 1.0 Integration Documentation](../../../integration/openid-connect/introduction.md#openid-certified).

## Variables

Some of the values within this page can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

{{< config-alert-example >}}

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    hmac_secret: 'this_is_a_secret_abc123abc123abc'
    jwks:
      - key_id: 'example'
        algorithm: 'RS256'
        use: 'sig'
        key: |
          -----BEGIN PRIVATE KEY-----
          ...
          -----END PRIVATE KEY-----
        certificate_chain: |
          -----BEGIN CERTIFICATE-----
          ...
          -----END CERTIFICATE-----
          -----BEGIN CERTIFICATE-----
          ...
          -----END CERTIFICATE-----
    enable_client_debug_messages: false
    minimum_parameter_entropy: 8
    enforce_pkce: 'public_clients_only'
    enable_pkce_plain_challenge: false
    enable_jwt_access_token_stateless_introspection: false
    discovery_signed_response_alg: 'none'
    discovery_signed_response_key_id: ''
    require_pushed_authorization_requests: false
    authorization_policies:
      policy_name:
        default_policy: 'two_factor'
        rules:
          - policy: 'deny'
            subject: 'group:services'
            networks:
              - '192.168.1.0/24'
              - '192.168.2.51'
    lifespans:
      access_token: '1h'
      authorize_code: '1m'
      id_token: '1h'
      refresh_token: '90m'
    claims_policies:
      policy_name:
        id_token: []
        access_token: []
        id_token_audience_mode: 'specification'
        custom_claims:
          claim_name:
            name: 'claim_name'
            attribute: 'attribute_name'
    scopes:
      scope_name:
        claims: []
    cors:
      endpoints:
        - 'authorization'
        - 'token'
        - 'revocation'
        - 'introspection'
      allowed_origins:
        - 'https://{{< sitevar name="domain" nojs="example.com" >}}'
      allowed_origins_from_client_redirect_uris: false
```

## Options

### hmac_secret

{{< confkey type="string" required="yes" secret="yes" >}}

The HMAC secret used to sign the [JWT]'s. The provided string is hashed to a SHA256 ([RFC6234]) byte string for the
purpose of meeting the required format.

It's __strongly recommended__ this is a
[Random Alphanumeric String](../../../reference/guides/generating-secure-values.md#generating-a-random-alphanumeric-string)
with 64 or more characters.

### jwks

{{< confkey type="list(object)" required="yes" >}}

The list of issuer JSON Web Keys. At least one of these must be an RSA Private key and be configured with the RS256
algorithm. Can also be used to configure many types of JSON Web Keys for the issuer such as the other RSA based JSON Web
Key formats and ECDSA JSON Web Key formats.

The default key for each algorithm is decided based on the order of this list. The first key for each algorithm is
considered the default if a client is not configured to use a specific key id. For example if a client has
[id_token_signed_response_alg](clients.md#id_token_signed_response_alg) `ES256` and
[id_token_signed_response_key_id](clients.md#id_token_signed_response_key_id) is not specified then the first `ES256`
key in this list is used.

The following is a contextual example (see below for information regarding each option):

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    jwks:
      - key_id: 'example'
        algorithm: 'RS256'
        use: 'sig'
        key: |
          -----BEGIN PRIVATE KEY-----
          ...
          -----END PRIVATE KEY-----
        certificate_chain: |
          -----BEGIN CERTIFICATE-----
          ...
          -----END CERTIFICATE-----
          -----BEGIN CERTIFICATE-----
          ...
          -----END CERTIFICATE-----
```

#### key_id

{{< confkey type="string" default="<thumbprint of public key>" required="no" >}}

Completely optional, and generally discouraged unless there is a collision between the automatically generated key id's.
If provided must be a unique string with 100 or fewer characters, with a recommendation to use a length less
than 15. In addition, it must meet the following rules:

- Match the regular expression `^[a-zA-Z0-9](([a-zA-Z0-9._~-]*)([a-zA-Z0-9]))?$` which should enforce the following rules:
  - Start with an alphanumeric character.
  - End with an alphanumeric character.
  - Only contain the [RFC3986 Unreserved Characters](https://datatracker.ietf.org/doc/html/rfc3986#section-2.3).

The default if this value is omitted is the first 7 characters of the public key SHA256 thumbprint encoded into
hexadecimal, followed by a hyphen, then followed by the lowercase algorithm value.

#### use

{{< confkey type="string" default="sig" required="no" >}}

The key usage. Defaults to `sig`. Available options are `sig` and `enc`.

#### algorithm

{{< confkey type="string" default="RS256" required="situational" >}}

The algorithm for this key. This value typically optional as it can be automatically detected based on the type of key
in some situations.

See the response object table in the [integration guide](../../../integration/openid-connect/introduction.md#response-object)
for more information. The `Algorithm` column lists supported values, the `Key` column references the required
[key](#key) type constraints that exist for the algorithm, and the `JWK Default Conditions` column briefly explains the
conditions under which it's the default algorithm.

At least one `RSA256` key must be provided.

#### key

{{< confkey type="string" required="yes" >}}

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
The methods described to generate this value often output both a Private Key and Public Key. This option only accepts
a Private Key. See the [certificate_chain](#certificate_chain) for Public Key configuration options; though this isn't
necessary in most situations.
{{< /callout >}}

The private key used to sign or encrypt the [OpenID Connect 1.0] issued [JWT]'s when using this JWK. The key must be
generated by the administrator and can be done by following the
[Generating an RSA Keypair](../../../reference/guides/generating-secure-values.md#generating-an-rsa-keypair) guide.

The key *__MUST__*:

* Be a PEM block encoded in the DER base64 format ([RFC4648]).
* Be either:
  * An RSA private key:
    * Encoded in conformance to the [PKCS#8] or [PKCS#1] specifications.
    * Have a key size of at least 2048 bits.
  * An ECDSA private key:
    * Encoded in conformance to the [PKCS#8] or [SECG1] specifications.
    * Use one of the following elliptical curves:
      * P-256.
      * P-384.
      * P-512.
* Include matching public key data if the [certificate_chain](#certificate_chain) is provided for the first certificate in the chain.

[PKCS#8]: https://datatracker.ietf.org/doc/html/rfc5208
[PKCS#1]: https://datatracker.ietf.org/doc/html/rfc8017
[SECG1]: https://datatracker.ietf.org/doc/html/rfc5915

It is recommended that you use a file to specify this particular option. In particular we recommend enabling the
`template` [file filter](../../methods/files.md#file-filters) and using the following example to assuming the path to
the file is `/config/secrets/oidc/jwks/rsa.2048.key`:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    jwks:
      - key: {{ secret "/config/secrets/oidc/jwks/rsa.2048.key" | mindent 10 "|" | msquote }}
```

#### certificate_chain

{{< confkey type="string" required="no" >}}

{{< callout context="tip" title="Not Required" icon="outline/alert-triangle" >}}
This option is rarely required as most clients do not support validating these values in the JSON Web Key Set document.
{{< /callout >}}

The certificate chain/bundle to be used with the [key](#key) DER base64 ([RFC4648])
encoded PEM format used to sign/encrypt the [OpenID Connect 1.0] [JWT]'s. When configured it enables the [x5c] and [x5t]
JSON Web Key's in the JSON Web Key Set
[Discoverable Endpoint](../../../integration/openid-connect/introduction.md#discoverable-endpoints) as per [RFC7517].

[RFC7517]: https://datatracker.ietf.org/doc/html/rfc7517
[x5c]: https://datatracker.ietf.org/doc/html/rfc7517#section-4.7
[x5t]: https://datatracker.ietf.org/doc/html/rfc7517#section-4.8

The certificate chain *__MUST__*:

* Include matching public key data for the [key](#key).
* Include only certificates valid for the current date.
* Contain only generally valid certificates.
* Include only sequentially signed certificates i.e. the first certificate must be signed by the second certificate
  (if provided) and the second certificate must be signed by the third (if provided), and so on.

### enable_client_debug_messages

{{< confkey type="boolean" default="false" required="no" >}}

Allows additional debug messages to be sent to the clients.

### minimum_parameter_entropy

{{< confkey type="integer" default="8" required="no" >}}

{{< callout context="danger" title="Security Note" icon="outline/alert-octagon" >}}
Changing this value is generally discouraged, reducing it from the default can theoretically
make certain scenarios less secure. It is highly encouraged that if your OpenID Connect 1.0 Relying Party does not send
these parameters or sends parameters with a lower length than the default that they implement a change rather than
changing this value.
{{< /callout >}}

This controls the minimum length of the `nonce` and `state` parameters. Setting this value to `-1` completely disables
this validation.

### enforce_pkce

{{< confkey type="string" default="public_clients_only" required="no" >}}

{{< callout context="danger" title="Security Note" icon="outline/alert-octagon" >}}
Changing this value to `never` is generally discouraged, reducing it from the default can
theoretically make certain client-side applications (mobile applications, SPA) vulnerable to CSRF and authorization code
interception attacks.
{{< /callout >}}

[Proof Key for Code Exchange](https://datatracker.ietf.org/doc/html/rfc7636) enforcement policy: if specified, must be
either `never`, `public_clients_only` or `always`.

If set to `public_clients_only` (default), [PKCE] will be required for public clients using the
[Authorization Code Flow].

When set to `always`, [PKCE] will be required for all clients using the Authorization Code flow.

### enable_pkce_plain_challenge

{{< confkey type="boolean" default="false" required="no" >}}

{{< callout context="danger" title="Security Note" icon="outline/alert-octagon" >}}
Changing this value is generally discouraged. Applications should use the `S256`
[PKCE](https://datatracker.ietf.org/doc/html/rfc7636) challenge method instead.
{{< /callout >}}

Allows [PKCE] `plain` challenges when set to `true`.

### enable_jwt_access_token_stateless_introspection

{{< confkey type="boolean" default="false" required="no" >}}

Allows [JWT Access Tokens](https://oauth.net/2/jwt-access-tokens/) to be introspected using a stateless model where
the JWT claims have all of the required introspection information, and assumes that they have not been revoked. This is
strongly discouraged unless you have a very specific use case.

A client with an [access_token_signed_response_alg](clients.md#access_token_signed_response_alg) or
[access_token_signed_response_key_id](clients.md#access_token_signed_response_key_id) must be configured for this option to
be enabled.

### discovery_signed_response_alg

{{< confkey type="string" default="none" required="no" >}}

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
Many clients do not support this option and it has a performance cost. It's therefore recommended
unless you have a specific need that you do not enable this option.
{{< /callout >}}

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
This value is completely ignored if the
[discovery_signed_response_key_id](#discovery_signed_response_key_id) is defined.
{{< /callout >}}

The algorithm used to sign the [OAuth 2.0 Authorization Server Metadata] and [OpenID Connect Discovery 1.0] responses.
Per the specifications this Signed JSON Web Token is stored in the `signed_metadata` value using the compact encoding.

See the response object section of the
[integration guide](../../../integration/openid-connect/introduction.md#response-object) for more information including
the algorithm column for supported values.

With the exclusion of `none` which excludes the `signed_metadata` value, the algorithm chosen must have a key
configured in the [jwks](#jwks) section to be considered valid.

See the response object section of the [integration guide](../../../integration/openid-connect/introduction.md#response-object)
for more information including the algorithm column for supported values.

### discovery_signed_response_key_id

{{< confkey type="string" required="no" >}}

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
Many clients do not support this option and it has a performance cost. It's therefore recommended
unless you have a specific need that you do not enable this option.
{{< /callout >}}

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
This value automatically configures the [discovery_signed_response_alg](#discovery_signed_response_alg)
value with the algorithm of the specified key.
{{< /callout >}}

The algorithm used to sign the [OAuth 2.0 Authorization Server Metadata] and [OpenID Connect Discovery 1.0] responses.
The value of this must one of those provided or calculated in the [jwks](#jwks). Per the specifications this Signed JSON
Web Token is stored in the `signed_metadata` value using the compact encoding.

### require_pushed_authorization_requests

{{< confkey type="boolean" default="false" required="no" >}}

When enabled all authorization requests must use the [Pushed Authorization Requests] flow.

### authorization_policies

{{< confkey type="dictionary(object)" required="no" >}}

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
This section is aimed at providing authorization customization for various
[OpenID Connect 1.0 Registered Clients](clients.md#authorization_policy). This section should not be confused with the
[Access Control Rules](../../security/access-control.md#rules) section, the way these policies are used and the options
available are distinctly and intentionally different to those of the
[Access Control Rules](../../security/access-control.md#rules) unless explicitly specified in this section. The reasons
for the differences are clearly explained in the
[OpenID Connect 1.0 FAQ](../../../integration/openid-connect/frequently-asked-questions.md#why-doesnt-the-access-control-configuration-work-with-openid-connect-10)
and [ADR1](../../../reference/architecture-decision-log/1.md). These policies specifically apply solely to Authorization Requests and
should not be used as a crutch for applications which do not implement the most basic
level of access control on their end.
{{< /callout >}}

The authorization policies section allows creating custom authorization policies which can be applied to clients. This
is useful if you wish to only allow specific users to access specific clients i.e. RBAC. It's generally recommended
however that users rely on the [OpenID Connect 1.0] relying party to provide RBAC controls based on the available
claims.

Each policy applies one of the effective policies which can be either `one_factor` or `two_factor` as per the standard
policies, or also the `deny` policy which is exclusively available via these configuration options.

Each rule within a policy is matched in order where the first fully matching rule is the applied policy. If the `deny`
rule is matched the user is not asked for consent and it is considered a rejected consent and returns an
[OpenID Connect 1.0] `access_denied` error.

The key for the policy itself is the name of the policy, which is used when configuring the client
[authorization_policy](clients.md#authorization_policy) option. In the example we name the policy `policy_name`.

The follow example shows a policy named `policy_name` which will `deny` access to users in the `services` group, with
a default policy of `two_factor` for everyone else. This policy is applied to the client with id
`client_with_policy_name`. You should refer to the below headings which describe each configuration key in more detail.

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    authorization_policies:
      policy_name:
        default_policy: 'two_factor'
        rules:
          - policy: 'deny'
            subject: 'group:services'
            networks:
              - '192.168.1.0/24'
              - '192.168.2.51'
    clients:
      - client_id: 'client_with_policy_name'
        authorization_policy: 'policy_name'
```

#### default_policy

{{< confkey type="string" default="two_factor" required="no" >}}

The default effective policy if none of the rules are able to determine the effective policy.

#### rules

{{< confkey type="list(object)" required="yes" >}}

The list of rules which this policy should consider when choosing the effective policy. This must be included for the
policy to be considered valid.

##### policy

{{< confkey type="string" default="two_factor" required="no" >}}

The policy which is applied if this rule matches. Valid values are `one_factor`, `two_factor`, and `deny`.

##### subject

{{< confkey type="list(list(string))" required="situational" >}}

_**Situational Note:** Either this option or the [networks](#networks) must be configured or this rule is considered
invalid._

The subjects criteria as per the [Access Control Configuration](../../security/access-control.md#subject).

##### networks

{{< confkey type="list(string)" syntax="network" required="situational" >}}
{{< callout context="danger" title="Security Note" icon="outline/rocket" >}}
The rules can only apply to the Authorization Code Flow when the resource owner is optionally providing
consent to the Authorization Request. While this is not a major issue for the [subject](#subject) criteria, the users
IP address may change and there is no technical way to enforce this check after consent has been granted and the tokens
have been issued. See [ADR1](../../../reference/architecture-decision-log/1.md) for more information.
{{< /callout >}}

_**Situational Note:** Either this option or the [subject](#subject) must be configured or this rule is considered
invalid._

The list of networks this rule applies to. Items in this list can also be named
[Network Definitions](../../definitions/network.md).

### lifespans

Token lifespans configuration. It's generally recommended keeping these values similar to the default values and to
utilize refresh tokens. For more information read this documentation about the [token lifespan].

#### access_token

{{< confkey type="string,integer" syntax="duration" default="1 hour" required="no" >}}

The default maximum lifetime of an access token.

#### refresh_token

{{< confkey type="string,integer" syntax="duration" default="1 hour 30 minutes" required="no" >}}

The default maximum lifetime of a refresh token. The refresh token can be used to obtain new refresh tokens as well as
access tokens or id tokens with an up-to-date expiration.

A good starting point is 50% more or 30 minutes more (which ever is less) time than the highest lifespan out of the
[access token](#access_token) lifespan and the [id token](#id_token) lifespan. For instance the default for all of these
is 60 minutes, so the default refresh token lifespan is 90 minutes.

#### id_token

{{< confkey type="string,integer" syntax="duration" default="1 hour" required="no" >}}

The default maximum lifetime of an ID token.

#### authorize_code

{{< confkey type="string,integer" syntax="duration" default="1 minute" required="no" >}}

The default maximum lifetime of an authorize code.

#### device_code

{{< confkey type="string,integer" syntax="duration" default="10 minutes" required="no" >}}

The default maximum lifetime of an device code.

#### custom

{{< confkey type="dictionary(object)" required="no" >}}

The custom lifespan configuration allows customizing the lifespans per-client. The custom lifespans must be utilized
with the client [lifespan](clients.md#lifespan) option which applies those settings to that client. Custom lifespans
can be configured in a very granular way, either solely by the token type, or by the token type for each grant type.
If a value is omitted it automatically uses the next value in the precedence tree. The tree is as follows:

1. Custom by token type and by grant.
2. Custom by token type.
3. Global default value.

The key for the custom lifespan itself is the name of the lifespan, which is used when configuring the client
[lifespan](clients.md#lifespan) option. In the example we name the lifespan `lifespan_name`.

##### Example

The following is an exhaustive example of all of the options available. Each of these options must follow all of the
same rules as the [access_token](#access_token), [authorize_code](#authorize_code), [id_token](#id_token), and
[refresh_token](#refresh_token) global default options. The global lifespan options are included for reference purposes.

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    lifespans:
      access_token: '1h'
      refresh_token: '90m'
      id_token: '1h'
      authorize_code: '1m'
      device_code: '10m'
      custom:
        lifespan_name:
          access_token: '1h'
          refresh_token: '90m'
          id_token: '1h'
          authorize_code: '1m'
          device_code: '10m'
          grants:
            authorize_code:
              access_token: '1h'
              refresh_token: '90m'
              id_token: '1h'
            device_code:
              access_token: '1h'
              refresh_token: '90m'
              id_token: '1h'
            implicit:
              access_token: '1h'
              refresh_token: '90m'
              id_token: '1h'
            client_credentials:
              access_token: '1h'
              refresh_token: '90m'
              id_token: '1h'
            refresh_token:
              access_token: '1h'
              refresh_token: '90m'
              id_token: '1h'
            jwt_bearer:
              access_token: '1h'
              refresh_token: '90m'
              id_token: '1h'
```

### claims_policies

{{< confkey type="string" syntax="dictionary" common="dictionary-reference" required="no" >}}

The claims policies are policies which allow customizing the behaviour of claims and the available claims for a
particular client.

The keys under `claims_policies` is an arbitrary value that can be used in the
[OpenID Connect 1.0 Client](clients.md#claims_policy) as the [claims_policy](clients.md#claims_policy) value.

#### id_token

{{< confkey type="list(string)" required="no" >}}

{{< callout context="danger" title="Security Notice" icon="outline/alert-octagon" >}}
This option is a escape hatch which should not normally be used. It allows confidential personally identifiable
information to be hydrated into the ID Token which is not normally encrypted. In addition this behaviour is only
necessary for clients which do not actually support OpenID Connect 1.0 and indicates a significant bug with the client.

This also is a common indicator that the client uses claims other than `iss` and `sub` to link users with the provider,
which is a fairly significant security issue.

For these reasons this option is highly discouraged and it's recommended the client in question fixes this significant
bug instead. This option is provided only on a best effort basis
{{< /callout >}}

The list of claims automatically copied to the ID Token in addition to the standard ID Token claims provided the
relevant scope was granted.

#### id_token_audience_mode

{{< confkey type="string" default="specification" required="no" >}}

The ID Token audience derivation mode for clients using this claims policy. It's recommended this is not configured
as the default mode is the correct mode in almost all situations, and if are considering changing this first read
the section on audiences in the [Integration Guide](../../../integration/openid-connect/introduction.md#audiences),
as there may be unintended security issues caused for relying parties that trust Authelia as a provider if you're not
cautious.

The following table describes all of the modes. Please note that any mode value prefixed with `experimental-` may be
removed or renamed without notice, and it's suggested if you're using these modes that you start a
[Discussion](https://github.com/authelia/authelia/discussions/new?category=show-and-tell) showcasing how you're using
a specific mode so we can adequately gauge its overall value.

|         Value         |                                                   Description                                                   |
|:---------------------:|:---------------------------------------------------------------------------------------------------------------:|
|    `specification`    |           This is the specification compliant mode where only the client id is recorded in the claim.           |
| `experimental-merged` | This mode includes the same value as `specification` but also merges the granted audience from the Access Token |

#### access_token

{{< confkey type="list(string)" required="no" >}}

The list of claims automatically copied to the Access Token in addition to the standard JWT Profile claims provided the
relevant scope was granted.

#### custom_claims

{{< confkey type="string" syntax="dictionary" common="dictionary-reference" required="no" >}}

The list of claims available in this policy in addition to the standard claims. These claims are anchored to attributes
which can either be concrete attributes from the [first factor](../../first-factor/introduction.md) backend or can be
those defined via [definitions](../../definitions/user-attributes.md).

The keys under `custom_claims` are arbitrary values, and by default are the claim name and attribute values.

##### name

{{< confkey type="string" required="no" >}}

The claim name for this claim. By default it's the same as the dictionary key.

##### attribute

{{< confkey type="string" required="no" >}}

The attribute name that this claim returns. By default it's the same as the dictionary key.

### scopes

{{< confkey type="string" syntax="dictionary" common="dictionary-reference" required="no" >}}

A list of scope definitions available in addition to the standard ones.

The keys under `scopes` are arbitrary values which are the names of the scopes.

#### claims

{{< confkey type="list(string)" required="no" >}}

The claims to be available to this scope.

If the scope is configured in a [OpenID Connect 1.0 Client](clients.md#scopes) in the [scopes](clients.md#scopes) then
every claim available in this list must either be a Standard Claim or must be fulfilled by the
[claims_policy](clients.md#claims_policy).

### cors

Some [OpenID Connect 1.0] Endpoints need to allow cross-origin resource sharing; however, some are optional. This section allows
you to configure the optional parts. We reply with CORS headers when the request includes the Origin header.

#### endpoints

{{< confkey type="list(string)" required="no" >}}

A list of endpoints to configure with cross-origin resource sharing headers. It is recommended that the `userinfo`
option is at least in this list. The potential endpoints which this can be enabled on are as follows:

* authorization
* pushed-authorization-request
* token
* revocation
* introspection
* userinfo

#### allowed_origins

{{< confkey type="list(string)" required="no" >}}

A list of permitted origins.

Any origin with https is permitted unless this option is configured or the
[allowed_origins_from_client_redirect_uris](#allowed_origins_from_client_redirect_uris) option is enabled. This means
you must configure this option manually if you want http endpoints to be permitted to make cross-origin requests to the
[OpenID Connect 1.0] endpoints, however this is not recommended.

Origins must only have the scheme, hostname and port, they may not have a trailing slash or path.

In addition to an Origin URI, you may specify the wildcard origin in the allowed_origins. It MUST be specified by itself
and the [allowed_origins_from_client_redirect_uris](#allowed_origins_from_client_redirect_uris) MUST NOT be enabled. The
wildcard origin is denoted as `*`. Examples:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    cors:
      allowed_origins: "*"
```

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    cors:
      allowed_origins:
        - "*"
```

#### allowed_origins_from_client_redirect_uris

{{< confkey type="boolean" default="false" required="no" >}}

Automatically adds the origin portion of all redirect URI's on all clients to the list of
[allowed_origins](#allowed_origins), provided they have the scheme http or https and do not have the hostname of
localhost.

### clients

{{< confkey type="list(object)" required="yes" >}}

See the [OpenID Connect 1.0 Registered Clients](clients.md) documentation for configuring clients.

## Integration

To integrate Authelia's [OpenID Connect 1.0] implementation with a relying party please see the
[integration docs](../../../integration/openid-connect/introduction.md).

[token lifespan]: https://docs.apigee.com/api-platform/antipatterns/oauth-long-expiration
[OpenID Connect 1.0]: https://openid.net/connect/
[OAuth 2.0 Authorization Server Metadata]: https://oauth.net/2/authorization-server-metadata/
[OpenID Connect Discovery 1.0]: https://openid.net/specs/openid-connect-discovery-1_0.html
[Token Endpoint]: https://openid.net/specs/openid-connect-core-1_0.html#TokenEndpoint
[JWT]: https://datatracker.ietf.org/doc/html/rfc7519
[RFC6234]: https://datatracker.ietf.org/doc/html/rfc6234
[RFC4648]: https://datatracker.ietf.org/doc/html/rfc4648
[RFC7468]: https://datatracker.ietf.org/doc/html/rfc7468
[RFC6749 Section 2.1]: https://datatracker.ietf.org/doc/html/rfc6749#section-2.1
[PKCE]: https://datatracker.ietf.org/doc/html/rfc7636
[Authorization Code Flow]: https://openid.net/specs/openid-connect-core-1_0.html#CodeFlowAuth
[Subject Identifier Type]: https://openid.net/specs/openid-connect-core-1_0.html#SubjectIDTypes
[Pairwise Identifier Algorithm]: https://openid.net/specs/openid-connect-core-1_0.html#PairwiseAlg
[Pushed Authorization Requests]: https://datatracker.ietf.org/doc/html/rfc9126
[OpenID Certified™]: https://openid.net/certification/
[OpenID Connect™ protocol]: https://openid.net/developers/how-connect-works/
