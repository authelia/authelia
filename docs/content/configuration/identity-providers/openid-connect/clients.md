---
title: "OpenID Connect 1.0 Clients"
description: "OpenID Connect 1.0 Registered Clients Configuration"
summary: "Authelia can operate as an OpenID Connect 1.0 Provider. This section describes how to configure the registered clients."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 110220
toc: true
aliases:
  - /c/oidc/registered-clients
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

This section covers specifics regarding configuring the providers registered clients for [OpenID Connect 1.0]. For the
provider specific configuration and information not related to clients see the [OpenID Connect 1.0 Provider](provider.md)
documentation.

More information about OpenID Connect 1.0 can be found in the [roadmap](../../../roadmap/active/openid-connect.md) and
in the [integration](../../../integration/openid-connect/introduction.md) documentation.

## Variables

Some of the values within this page can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

{{< config-alert-example >}}

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    clients:
      - client_id: 'unique-client-identifier'
        client_name: 'My Application'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        sector_identifier_uri: 'https://{{< sitevar name="domain" nojs="example.com" >}}/sector.json'
        public: false
        redirect_uris:
          - 'https://oidc.{{< sitevar name="domain" nojs="example.com" >}}:8080/oauth2/callback'
        request_uris:
          - 'https://oidc.{{< sitevar name="domain" nojs="example.com" >}}:8080/oidc/request-object.jwk'
        audience:
          - 'https://app.{{< sitevar name="domain" nojs="example.com" >}}'
        scopes:
          - 'openid'
          - 'groups'
          - 'email'
          - 'profile'
        grant_types:
          - 'refresh_token'
          - 'authorization_code'
        response_types:
          - 'code'
        response_modes:
          - 'form_post'
          - 'query'
          - 'fragment'
        authorization_policy: 'two_factor'
        lifespan: ''
        claims_policy: ''
        requested_audience_mode: 'explicit'
        consent_mode: 'explicit'
        pre_configured_consent_duration: '1 week'
        require_pushed_authorization_requests: false
        require_pkce: false
        pkce_challenge_method: 'S256'
        authorization_signed_response_key_id: ''
        authorization_signed_response_alg: 'RS256'
        authorization_encrypted_response_key_id: ''
        authorization_encrypted_response_alg: 'none'
        authorization_encrypted_response_enc: 'A128CBC-HS256'
        id_token_signed_response_key_id: ''
        id_token_signed_response_alg: 'RS256'
        id_token_encrypted_response_key_id: ''
        id_token_encrypted_response_alg: 'none'
        id_token_encrypted_response_enc: 'A128CBC-HS256'
        access_token_signed_response_key_id: ''
        access_token_signed_response_alg: 'none'
        access_token_encrypted_response_key_id: ''
        access_token_encrypted_response_alg: 'none'
        access_token_encrypted_response_enc: 'A128CBC-HS256'
        userinfo_signed_response_key_id: ''
        userinfo_signed_response_alg: 'none'
        userinfo_encrypted_response_key_id: ''
        userinfo_encrypted_response_alg: 'none'
        userinfo_encrypted_response_enc: 'A128CBC-HS256'
        introspection_signed_response_key_id: ''
        introspection_signed_response_alg: 'none'
        introspection_encrypted_response_key_id: ''
        introspection_encrypted_response_alg: 'none'
        introspection_encrypted_response_enc: 'A128CBC-HS256'
        request_object_signing_alg: 'RS256'
        request_object_encryption_alg: ''
        request_object_encryption_enc: ''
        token_endpoint_auth_method: 'client_secret_basic'
        token_endpoint_auth_signing_alg: 'RS256'
        revocation_endpoint_auth_method: 'client_secret_basic'
        revocation_endpoint_auth_signing_alg: 'RS256'
        introspection_endpoint_auth_method: 'client_secret_basic'
        introspection_endpoint_auth_signing_alg: 'RS256'
        pushed_authorization_request_endpoint_auth_method: 'client_secret_basic'
        pushed_authorization_request_endpoint_auth_signing_alg: 'RS256'
        jwks_uri: ''
        jwks:
          - key_id: 'example'
            algorithm: 'RS256'
            use: 'sig'
            key: |
              -----BEGIN RSA PUBLIC KEY-----
              ...
              -----END RSA PUBLIC KEY-----
            certificate_chain: |
              -----BEGIN CERTIFICATE-----
              ...
              -----END CERTIFICATE-----
              -----BEGIN CERTIFICATE-----
              ...
              -----END CERTIFICATE-----
```

## Options

### client_id

{{< confkey type="string" required="yes" >}}

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
We generally recommend using a semi-long random alphanumeric string for this value. See below for
specific limitations.
{{< /callout >}}

The Client ID for this client. It must exactly match the Client ID configured in the application consuming this client.

Valid Client ID's have the following characteristics:

- Less than or equal to 100 characters.
- Only contains [RFC3986 Unreserved Characters](https://datatracker.ietf.org/doc/html/rfc3986#section-2.3).
- Completely unique from other configured clients.

### client_name

{{< confkey type="string" default="*same as id*" required="no" >}}

A friendly name for this client shown in the UI. This defaults to the same as the ID.

### client_secret

{{< confkey type="string" required="situational" >}}

The shared secret between Authelia and the application consuming this client. This secret must match the secret
configured in the application.

This secret must be generated by the administrator and can be done by following the
[How Do I Generate a Client Identifier or Client Secret](../../../integration/openid-connect/frequently-asked-questions.md#how-do-i-generate-a-client-identifier-or-client-secret) FAQ.

This must be provided when the client is a confidential client type provided with the exception if you've configured a
[token_endpoint_auth_method](#token_endpoint_auth_method) that uses a credential type that isn't a secret (i.e. a key), and
must be blank when using the public client type. To set the client type to public see the [public](#public)
configuration option.

### sector_identifier_uri

{{< confkey type="string" required="no" >}}

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
Because adjusting this option will inevitably change the `sub` claim of all tokens generated for
the specified client, changing this should cause the relying party to detect all future authorizations as completely new
users.
{{< /callout >}}

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
This **must** either not be configured at all i.e. commented or completely absent from the
configuration, or it must be an absolute HTTPS URL which contains a valid sector identifier JSON document. Configuration
of this option with the `https://` scheme per the requirements will cause Authelia to validate this JSON document.
{{< /callout >}}

A valid `sector_identifier_uri` will:
  1. Have the scheme `https://`.
  2. Be the absolute URI of a JSON document which:
     1. Is a JSON array of strings (URIs).
     2. Has every URI registered with this clients [redirect_uris](#redirect_uris) when compared using an exact string
        match as defined in [OAuth 2.0 Security Best Current Practice Section 2.1].
     3. May or may not have additional [redirect_uris](#redirect_uris) from other clients.

[OAuth 2.0 Security Best Current Practice Section 2.1]: https://datatracker.ietf.org/doc/html/draft-ietf-oauth-security-topics#section-2.1

Authelia utilizes UUID version 4 subject identifiers. By default the public [Subject Identifier Type] is utilized for
all clients. This means the subject identifiers will be the same for all clients. This configuration option enables
[Pairwise Identifier Algorithm] for this client, and configures the sector identifier utilized for both the storage and
the lookup of the subject identifier.

1. All clients who do not have this configured will generate the same subject identifier for a particular user
   regardless of which client obtains the ID token.
2. All clients which have the same sector identifier will:
   1. Have the same subject identifier for a particular user when compared to clients with the same sector identifier.
   2. Have a completely different subject identifier for a particular user when compared to:
      1. Any client with the public subject identifier type.
      2. Any client with a differing `sector_identifier_uri`.

In specific but limited scenarios this option is beneficial for privacy reasons. In particular this is useful when the
party utilizing the *Authelia* [OpenID Connect 1.0] Authorization Server is foreign and not controlled by the user. It would
prevent the third party utilizing the subject identifier with another third party in order to track the user.

Keep in mind depending on the other claims they may still be able to perform this tracking and it is not a silver
bullet. There are very few benefits when utilizing this in a homelab or business where no third party is utilizing
the server.

### public

{{< confkey type="boolean" default="false" required="no" >}}

This enables the public client type for this client. This is for clients that are not capable of maintaining
confidentiality of credentials, you can read more about client types in [RFC6749 Section 2.1]. This is particularly
useful for SPA's and CLI tools. This option requires setting the [client secret](#client_secret) to a blank string.

### redirect_uris

{{< confkey type="list(string)" required="yes" >}}

A list of valid callback URIs this client will redirect to. All other callbacks will be considered unsafe. The URIs are
case-sensitive and they differ from application to application - the community has provided
[a list of URL´s for common applications](../../../integration/openid-connect/introduction.md).

Some restrictions that have been placed on clients and
their redirect URIs are as follows:

1. If a client attempts to authorize with Authelia and its redirect URI is not listed in the client configuration the
   attempt to authorize will fail and an error will be generated.
2. The redirect URIs are case-sensitive.
3. The URI must include a scheme and that scheme must be one of `http` or `https`.

### request_uris

{{< confkey type="list(string)" required="no" >}}

A list of URIs which can be used for the OpenID Connect 1.0 Request Object to pass Authorize Request parameters via a
JSON Web Token remote URI using the `request_uri` parameter.

These URIs must have the `https` scheme.

### audience

{{< confkey type="list(string)" required="no" >}}

A whitelist of audiences this client is allowed to request. These audiences were previously automatically granted to all
access requests unless specifically requested otherwise. The current behavior is only those requested by the client in
the `audience` parameter are granted. This behavior can be tuned using the
[requested_audience_mode](#requested_audience_mode).

This value does not generally affect the minted ID Tokens as they are always issued with the client identifier being the
audience unless the [claims policy](#claims_policy) changes this behaviour.

### scopes

{{< confkey type="list(string)" default="openid,groups,profile,email" required="no" >}}

A list of scopes to allow this client to consume. See
[scope definitions](../../../integration/openid-connect/openid-connect-1.0-claims.md#scope-definitions) for more information. The
documentation for the application you are trying to configure [OpenID Connect 1.0] for will likely have a list of scopes
or claims required which can be matched with the above guide.

The scope values should generally be one of those documented in the
[scope definitions](../../../integration/openid-connect/openid-connect-1.0-claims.md#scope-definitions) with the exception of when a client requires a specific scope we do not define. Users should
expect to see a warning in the logs if they configure a scope not in our definitions with the exception of a client
where the configured [grant_types](#grant_types) includes the `client_credentials` grant in which case arbitrary scopes are
expected,

### grant_types

{{< confkey type="list(string)" default="authorization_code" required="no" >}}

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
It is recommended that this isn't configured at this time unless you know what you're doing.
{{< /callout >}}

The list of grant types this client is permitted to use in order to obtain access to the token endpoint to obtain the
granted tokens.

See the [Grant Types](../../../integration/openid-connect/introduction.md#grant-types) section of the
[OpenID Connect 1.0 Integration Guide](../../../integration/openid-connect/introduction.md#grant-types) for more information.

### response_types

{{< confkey type="list(string)" default="code" required="no" >}}

{{< callout context="danger" title="Security Note" icon="outline/alert-octagon" >}}
It is recommended that only the `code` response type (i.e. the default) is used. The other response
types are not as secure as this response type.
{{< /callout >}}

A list of response types this client supports. If a response type not in this list is requested by a client then an
error will be returned to the client. The response type indicates the types of values that are returned to the client.

See the [Response Types](../../../integration/openid-connect/introduction.md#response-types) section of the
[OpenID Connect 1.0 Integration Guide](../../../integration/openid-connect/introduction.md#response-types) for more information.

### response_modes

{{< confkey type="list(string)" default="form_post,query" required="no" >}}

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
It is recommended that this isn't configured at this time unless you know what you're doing.
{{< /callout >}}

A list of response modes this client supports. If a response mode not in this list is requested by a client then an
error will be returned to the client. The response mode controls how the response type is returned to the client.

See the [Response Modes](../../../integration/openid-connect/introduction.md#response-modes) section of the
[OpenID Connect 1.0 Integration Guide](../../../integration/openid-connect/introduction.md#response-modes) for more
information.

The default values are based on the [response_types](#response_types) values. When the [response_types](#response_types)
values include the `code` type then the `query` response mode will be included. When any other type is included the
`fragment` response mode will be included. It's important to note at this time we do not support the `none` response
type, but when it is supported it will include the `query` response mode.

### authorization_policy

{{< confkey type="string" default="two_factor" required="no" >}}

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
This option is aimed at providing authorization customization for this particular client. This option should not be
confused with the [Access Control Rules](../../security/access-control.md#rules) section and this option is distinctly
and intentionally different. The reasons for the differences are clearly explained in the
[OpenID Connect 1.0 FAQ](../../../integration/openid-connect/frequently-asked-questions.md#why-doesnt-the-access-control-configuration-work-with-openid-connect-10)
and [ADR1](../../../reference/architecture-decision-log/1.md). This policy specifically applies solely to Authorization Requests and
should not be used as a crutch for applications which do not implement the most basic
level of access control on their end.
{{< /callout >}}


The authorization policy for this client: either `one_factor`, `two_factor`, or one of the ones configured in the
provider [authorization_policies](./provider.md#authorization_policies) section.

The follow example shows a policy named `policy_name` which will `deny` access to users in the `services` group, with
a default policy of `two_factor` for everyone else. This policy is applied to the client with id
`client_with_policy_name`. You should refer to the [authorization_policies](./provider.md#authorization_policies)
section for more in depth information.

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    authorization_policies:
      policy_name:
        default_policy: 'two_factor'
        rules:
          - policy: 'deny'
            subject: 'group:services'
    clients:
      - client_id: 'client_with_policy_name'
        authorization_policy: 'policy_name'
```

### lifespan

{{< confkey type="string" default="" required="no" >}}

The name of the custom lifespan that this client uses. A custom lifespan is named and configured globally via the
[custom](provider.md#custom) section within [lifespans](provider.md#lifespans).

### claims_policy

{{< confkey type="string" default="" required="no" >}}

The name of the claims policy that this client uses. A claims policy is named and configured globally via the
[claims_policies](provider.md#claims_policies) for the OpenID Connect 1.0 Provider.

### requested_audience_mode

{{< confkey type="string" default="explicit" required="no" >}}

Controls the effective audience the client has requested. The following table describes the possible values and their
behavior. This value does not affect the issued ID Tokens as they are always issued with the client identifier being
the audience.

|  Value   |                                                   Description                                                    |
|:--------:|:----------------------------------------------------------------------------------------------------------------:|
| explicit |     Requires the client explicitly requests an audiences for an audience to be included in the issued tokens     |
| implicit | Assumes if the client is requesting all audiences it is permitted to request if the audience parameter is absent |

### consent_mode

{{< confkey type="string" default="auto" required="no" >}}

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
The `implicit` consent mode is not technically part of the specification. It theoretically could be
misused in certain conditions specifically with the public client type or when the client credentials (i.e. client
secret) has been exposed to an attacker. For these reasons this mode is discouraged.
{{< /callout >}}

Configures the fallback consent mode. If explicit consent or a condition that requires explicit consent is present this
setting has no effect. The following table describes the different modes:

|     Value      |                                                                  Description                                                                   |
|:--------------:|:----------------------------------------------------------------------------------------------------------------------------------------------:|
|      auto      | Automatically determined (default). Uses `explicit` unless [pre_configured_consent_duration] is specified in which case uses `pre-configured`. |
|    explicit    |                                   Requires the user provide unique explicit consent for every authorization.                                   |
|    implicit    |    Automatically assumes consent for every authorization, never asking the user if they wish to give consent. See the specific notes below.    |
| pre-configured |                            Allows the end-user to remember their consent for the [pre_configured_consent_duration].                            |

[pre_configured_consent_duration]: #pre_configured_consent_duration

See the [Frequently Asked Questions](../../../integration/openid-connect/frequently-asked-questions.md#why-does-authelia-ask-for-consent-when-ive-asked-for-my-consent-to-be-remembered-or-used-the-implicit-consent-policy)
for more information on specific behaviour around why consent may be required despite this configuration option.

### pre_configured_consent_duration

{{< confkey type="string,integer" syntax="duration" default="1 week" required="no" >}}

Specifying this in the configuration without a consent [consent_mode] enables the `pre-configured` mode. If this is
specified as well as the [consent_mode] then it only has an effect if the [consent_mode] is `pre-configured` or `auto`.

The period of time dictates how long a users choice to remember the pre-configured consent lasts.

Pre-configured consents are only valid if the subject, client id are exactly the same and the requested scopes/audience
match exactly with the granted scopes/audience.

[consent_mode]: #consent_mode

### require_pushed_authorization_requests

{{< confkey type="boolean" default="false" required="no" >}}

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
A majority of clients will not support this option as it requires a special but highly secure
authorization flow.
{{< /callout >}}

This configuration option enforces the use of a [Pushed Authorization Requests] flow for this registered client.
To enforce it for all clients see the global [pushed_authorizations enforce](provider.md#enforce) provider configuration
option.

### require_pkce

{{< confkey type="boolean" default="false" required="no" >}}

This configuration option enforces the use of [PKCE] for this registered client. To enforce it for all clients see the
global [enforce_pkce](provider.md#enforce_pkce) provider configuration option.

### pkce_challenge_method

{{< confkey type="string" default="" required="no" >}}

This setting enforces the use of the specified [PKCE] challenge method for this individual client. This setting also
effectively enables the [require_pkce](#require_pkce) option for this client.

Valid values are an empty string, `plain`, or `S256`. It should be noted that `S256` is strongly recommended if the
relying party supports it.

### authorization_signed_response_key_id

{{< confkey type="string" required="no" >}}

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
A majority of clients will not support this option with any value other than `none` as it implements the
[JARM](https://openid.net/specs/oauth-v2-jarm.html) specification where the whole authorization response becomes a
signed JWT and the signed JWT is optionally nested within an encrypted JWT.
{{< /callout >}}

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
This value automatically configures the [authorization_signed_response_alg](#authorization_signed_response_alg)
value with the algorithm of the specified key.
{{< /callout >}}

The key id of the [JSON Web Key] used to sign
[Authorization Responses](https://openid.net/specs/oauth-v2-jarm.html#name-response-encoding) for this client.

To be considered valid:

1. The chosen value must have a [JSON Web Key] configured in the [jwks] section with the use value `sig` and a
   matching `kid` to this value.

### authorization_signed_response_alg

{{< confkey type="string" default="RS256" required="no" >}}

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
A majority of clients will not support this option with any value other than `none` as it implements the
[JARM](https://openid.net/specs/oauth-v2-jarm.html) specification where the whole authorization response becomes a
signed JWT and the signed JWT is optionally nested within an encrypted JWT.
{{< /callout >}}

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
This value is completely ignored if the
[authorization_signed_response_key_id](#authorization_signed_response_key_id) is defined.
{{< /callout >}}

The algorithm of the [JSON Web Key] used to sign
[Authorization Responses](https://openid.net/specs/oauth-v2-jarm.html#name-response-encoding) for this client.
To be considered valid with exclusion of the value `none`:

1. The chosen value must have a [JSON Web Key] configured in the [jwks] section with the use value `sig` and a
   matching `alg` to this value.

See the response object section of the
[integration guide](../../../integration/openid-connect/introduction.md#response-object) for more information. The
supported values come from the algorithm column with a use of `sig`.

### authorization_encrypted_response_key_id

{{< confkey type="string" required="no" >}}

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
A majority of clients will not support this option with any value other than `none` as it implements the
[JARM](https://openid.net/specs/oauth-v2-jarm.html) specification where the whole authorization response becomes a
signed JWT and the signed JWT is optionally nested within an encrypted JWT.
{{< /callout >}}

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
This value is completely ignored if the
[authorization_encrypted_response_alg](#authorization_encrypted_response_alg) is defined.
{{< /callout >}}

The key id of the [JSON Web Key] used to encrypt
[Authorization Responses](https://openid.net/specs/oauth-v2-jarm.html#name-response-encoding) for this client.

To be considered valid:

1. The chosen value must have a [JSON Web Key] configured in the [jwks] section with the use value `enc` and a
   matching `kid` to this value.

### authorization_encrypted_response_alg

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
A majority of clients will not support this option with any value other than `none` as it implements the
[JARM](https://openid.net/specs/oauth-v2-jarm.html) specification where the whole authorization response becomes a
signed JWT and the signed JWT is optionally nested within an encrypted JWT.
{{< /callout >}}

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
This value is completely ignored if the
[authorization_encrypted_response_key_id](#authorization_encrypted_response_key_id) is defined.
{{< /callout >}}

The key algorithm of the [JSON Web Key] used to encrypt
[Authorization Responses](https://openid.net/specs/oauth-v2-jarm.html#name-response-encoding) for this client.

To be considered valid with exclusion of the value `none`:

1. The chosen value must have a [JSON Web Key] configured in the [jwks] section with the use value `enc` and a
   matching `alg` to this value.
2. The [authorization_signed_response_alg](#authorization_signed_response_alg) or
   [authorization_response_key_id](#authorization_signed_response_key_id) option must be configured.

See the response object section of the
[integration guide](../../../integration/openid-connect/introduction.md#response-object) for more information. The
supported values come from the algorithm column with a use of `enc`.

### authorization_encrypted_response_enc

{{< confkey type="string" default="A128CBC-HS256" required="no" >}}

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
A majority of clients will not support this option with any value other than `none` as it implements the
[JARM](https://openid.net/specs/oauth-v2-jarm.html) specification where the whole authorization response becomes a
signed JWT and the signed JWT is optionally nested within an encrypted JWT.
{{< /callout >}}

The content encryption algorithm used to encrypt the authorization responses.

See the encryption algorithms section of the
[integration guide](../../../integration/openid-connect/introduction.md#encryption-algorithms) for more information
including the algorithm column for supported values.

### id_token_signed_response_key_id

{{< confkey type="string" required="no" >}}

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
This value automatically configures the [id_token_signed_response_alg](#id_token_signed_response_alg)
value with the algorithm of the specified key.
{{< /callout >}}

The key id of the [JSON Web Key] used to sign ID Tokens for this client.

To be considered valid:

1. The chosen value must have a [JSON Web Key] configured in the [jwks] section with the use value `sig` and a
   matching `kid` to this value.

### id_token_signed_response_alg

{{< confkey type="string" default="RS256" required="no" >}}

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
This value is completely ignored if the
[id_token_signed_response_key_id](#id_token_signed_response_key_id) is defined.
{{< /callout >}}

The algorithm of the [JSON Web Key] used to sign ID Tokens for this client.
To be considered valid:

1. The chosen value must have a [JSON Web Key] configured in the [jwks] section with the use value `sig` and a
   matching `alg` to this value.

See the response object section of the
[integration guide](../../../integration/openid-connect/introduction.md#response-object) for more information. The
supported values come from the algorithm column with a use of `sig`.

### id_token_encrypted_response_key_id

{{< confkey type="string" required="no" >}}

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
A majority of clients will not support this option as the ID Token will be a nested within an
encrypted JWT.
{{< /callout >}}

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
This value is completely ignored if the
[id_token_encrypted_response_alg](#id_token_encrypted_response_alg) is defined.
{{< /callout >}}

The key id of the [JSON Web Key] used to encrypt ID Tokens for this client.

To be considered valid:

1. The chosen value must have a [JSON Web Key] configured in the [jwks] section with the use value `enc` and a
   matching `kid` to this value.

### id_token_encrypted_response_alg

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
A majority of clients will not support this option as the ID Token will be a nested within an
encrypted JWT.
{{< /callout >}}

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
This value is completely ignored if the
[id_token_encrypted_response_key_id](#id_token_encrypted_response_key_id) is defined.
{{< /callout >}}

The key algorithm of the [JSON Web Key] used to encrypt ID Tokens for this client.

To be considered valid with exclusion of the value `none`:

1. The chosen value must have a [JSON Web Key] configured in the [jwks] section with the use value `enc` and a
   matching `alg` to this value.
2. The [id_token_signed_response_alg](#id_token_signed_response_alg) or
   [id_token_response_key_id](#id_token_signed_response_key_id) option must be configured.

See the response object section of the
[integration guide](../../../integration/openid-connect/introduction.md#response-object) for more information. The
supported values come from the algorithm column with a use of `enc`.

### id_token_encrypted_response_enc

{{< confkey type="string" default="A128CBC-HS256" required="no" >}}

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
A majority of clients will not support this option as the ID Token will be a nested within an
encrypted JWT.
{{< /callout >}}

The content encryption algorithm used to encrypt the authorization responses.

See the encryption algorithms section of the
[integration guide](../../../integration/openid-connect/introduction.md#encryption-algorithms) for more information
including the algorithm column for supported values.

### access_token_signed_response_key_id

{{< confkey type="string" required="no" >}}

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
Using any value other than `none` for this option enables encoding Access Tokens as JWT's per
[RFC9068](https://datatracker.ietf.org/doc/html/rfc9068). It is critical to note that these Access Tokens should not be
treated as an ID Token and the semantics of validating these token types differ. The JWT Profile Access Token is
intended for resource servers to perform stateless validation of the Access Tokens and they should not be used to prove
identity. We therefore only recommend implementing these tokens in heavier use cases where the cost of validating the
Access Tokens against a database is too costly.
{{< /callout >}}

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
This value automatically configures the [access_token_signed_response_alg](#access_token_signed_response_alg)
value with the algorithm of the specified key.
{{< /callout >}}

The key id of the [JSON Web Key] used to sign Access Tokens for this client.

To be considered valid:

1. The chosen value must have a [JSON Web Key] configured in the [jwks] section with the use value `sig` and a
   matching `kid` to this value.

### access_token_signed_response_alg

{{< confkey type="string" default="none" required="no" >}}

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
Using any value other than `none` for this option enables encoding Access Tokens as JWT's per
[RFC9068](https://datatracker.ietf.org/doc/html/rfc9068). It is critical to note that these Access Tokens should not be
treated as an ID Token and the semantics of validating these token types differ. The JWT Profile Access Token is
intended for resource servers to perform stateless validation of the Access Tokens and they should not be used to prove
identity. We therefore only recommend implementing these tokens in heavier use cases where the cost of validating the
Access Tokens against a database is too costly.
{{< /callout >}}

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
This value is completely ignored if the
[access_token_signed_response_key_id](#access_token_signed_response_key_id) is defined.
{{< /callout >}}

The algorithm of the [JSON Web Key] used to sign Access Tokens for this client.
To be considered valid with exclusion of the value `none`:

1. The chosen value must have a [JSON Web Key] configured in the [jwks] section with the use value `sig` and a
   matching `alg` to this value.

See the response object section of the
[integration guide](../../../integration/openid-connect/introduction.md#response-object) for more information. The
supported values come from the algorithm column with a use of `sig`.

### access_token_encrypted_response_key_id

{{< confkey type="string" required="no" >}}

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
Using any value other than `none` for this option enables encoding Access Tokens as JWT's per
[RFC9068](https://datatracker.ietf.org/doc/html/rfc9068). It is critical to note that these Access Tokens should not be
treated as an ID Token and the semantics of validating these token types differ. The JWT Profile Access Token is
intended for resource servers to perform stateless validation of the Access Tokens and they should not be used to prove
identity. We therefore only recommend implementing these tokens in heavier use cases where the cost of validating the
Access Tokens against a database is too costly.
{{< /callout >}}

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
This value is completely ignored if the
[access_token_encrypted_response_alg](#access_token_encrypted_response_alg) is defined.
{{< /callout >}}

The key id of the [JSON Web Key] used to encrypt Access Tokens for this client.

To be considered valid:

1. The chosen value must have a [JSON Web Key] configured in the [jwks] section with the use value `enc` and a
   matching `kid` to this value.

### access_token_encrypted_response_alg

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
Using any value other than `none` for this option enables encoding Access Tokens as JWT's per
[RFC9068](https://datatracker.ietf.org/doc/html/rfc9068). It is critical to note that these Access Tokens should not be
treated as an ID Token and the semantics of validating these token types differ. The JWT Profile Access Token is
intended for resource servers to perform stateless validation of the Access Tokens and they should not be used to prove
identity. We therefore only recommend implementing these tokens in heavier use cases where the cost of validating the
Access Tokens against a database is too costly.
{{< /callout >}}

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
This value is completely ignored if the
[access_token_encrypted_response_key_id](#access_token_encrypted_response_key_id) is defined.
{{< /callout >}}

The key algorithm of the [JSON Web Key] used to encrypt Access Tokens for this client.

To be considered valid with exclusion of the value `none`:

1. The chosen value must have a [JSON Web Key] configured in the [jwks] section with the use value `enc` and a
   matching `alg` to this value.
2. The [access_token_signed_response_alg](#access_token_signed_response_alg) or
   [access_token_response_key_id](#access_token_signed_response_key_id) option must be configured.

See the response object section of the
[integration guide](../../../integration/openid-connect/introduction.md#response-object) for more information. The
supported values come from the algorithm column with a use of `enc`.

### access_token_encrypted_response_enc

{{< confkey type="string" default="A128CBC-HS256" required="no" >}}

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
Using any value other than `none` for this option enables encoding Access Tokens as JWT's per
[RFC9068](https://datatracker.ietf.org/doc/html/rfc9068). It is critical to note that these Access Tokens should not be
treated as an ID Token and the semantics of validating these token types differ. The JWT Profile Access Token is
intended for resource servers to perform stateless validation of the Access Tokens and they should not be used to prove
identity. We therefore only recommend implementing these tokens in heavier use cases where the cost of validating the
Access Tokens against a database is too costly.
{{< /callout >}}

The content encryption algorithm used to encrypt the authorization responses.

See the encryption algorithms section of the
[integration guide](../../../integration/openid-connect/introduction.md#encryption-algorithms) for more information
including the algorithm column for supported values.

### userinfo_signed_response_key_id

{{< confkey type="string" required="no" >}}

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
A majority of clients will not support this option with any value other than `none` as the whole User Information
response rather than being a JSON document becomes a signed JWT and the signed JWT is optionally nested within an
encrypted JWT.
{{< /callout >}}

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
This value automatically configures the [userinfo_signed_response_alg](#userinfo_signed_response_alg)
value with the algorithm of the specified key.
{{< /callout >}}

The key id of the [JSON Web Key] used to sign User Information responses for this client.

To be considered valid:

1. The chosen value must have a [JSON Web Key] configured in the [jwks] section with the use value `sig` and a
   matching `kid` to this value.

### userinfo_signed_response_alg

{{< confkey type="string" default="none" required="no" >}}

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
A majority of clients will not support this option with any value other than `none` as the whole User Information
response rather than being a JSON document becomes a signed JWT and the signed JWT is optionally nested within an
encrypted JWT.
{{< /callout >}}

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
This value is completely ignored if the
[userinfo_signed_response_key_id](#userinfo_signed_response_key_id) is defined.
{{< /callout >}}

The algorithm of the [JSON Web Key] used to sign User Information responses for this client.
To be considered valid with exclusion of the value `none`:

1. The chosen value must have a [JSON Web Key] configured in the [jwks] section with the use value `sig` and a
   matching `alg` to this value.

See the response object section of the
[integration guide](../../../integration/openid-connect/introduction.md#response-object) for more information. The
supported values come from the algorithm column with a use of `sig`.

### userinfo_encrypted_response_key_id

{{< confkey type="string" required="no" >}}

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
A majority of clients will not support this option with any value other than `none` as the whole User Information
response rather than being a JSON document becomes a signed JWT and the signed JWT is optionally nested within an
encrypted JWT.
{{< /callout >}}

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
This value is completely ignored if the
[userinfo_encrypted_response_alg](#userinfo_encrypted_response_alg) is defined.
{{< /callout >}}

The key id of the [JSON Web Key] used to encrypt User Information responses for this client.

To be considered valid:

1. The chosen value must have a [JSON Web Key] configured in the [jwks] section with the use value `enc` and a
   matching `kid` to this value.

### userinfo_encrypted_response_alg

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
A majority of clients will not support this option with any value other than `none` as the whole User Information
response rather than being a JSON document becomes a signed JWT and the signed JWT is optionally nested within an
encrypted JWT.
{{< /callout >}}

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
This value is completely ignored if the
[userinfo_encrypted_response_key_id](#userinfo_encrypted_response_key_id) is defined.
{{< /callout >}}

The key algorithm of the [JSON Web Key] used to encrypt User Information responses for this client.

To be considered valid with exclusion of the value `none`:

1. The chosen value must have a [JSON Web Key] configured in the [jwks] section with the use value `enc` and a
   matching `alg` to this value.
2. The [userinfo_signed_response_alg](#userinfo_signed_response_alg) or
   [userinfo_response_key_id](#userinfo_signed_response_key_id) option must be configured.

See the response object section of the
[integration guide](../../../integration/openid-connect/introduction.md#response-object) for more information. The
supported values come from the algorithm column with a use of `enc`.

### userinfo_encrypted_response_enc

{{< confkey type="string" default="A128CBC-HS256" required="no" >}}

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
A majority of clients will not support this option with any value other than `none` as the whole User Information
response rather than being a JSON document becomes a signed JWT and the signed JWT is optionally nested within an
encrypted JWT.
{{< /callout >}}

The content encryption algorithm used to encrypt the authorization responses.

See the encryption algorithms section of the
[integration guide](../../../integration/openid-connect/introduction.md#encryption-algorithms) for more information
including the algorithm column for supported values.

### introspection_signed_response_key_id

{{< confkey type="string" required="no" >}}

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
A majority of clients will not support this option with any value other than `none` as this enables encoding the
[Introspection Response](https://datatracker.ietf.org/doc/html/rfc7662#section-2.2) as a JWT as per
[JWT Response for OAuth Token Introspection](https://www.ietf.org/archive/id/draft-ietf-oauth-jwt-introspection-response-12.html)
i.e. rather than being a JSON document the
[Introspection Response](https://datatracker.ietf.org/doc/html/rfc7662#section-2.2) becomes a signed JWT in the
`application/token-introspection+jwt` format and the signed JWT is optionally nested within an encrypted JWT.
{{< /callout >}}

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
This value automatically configures the [introspection_signed_response_alg](#introspection_signed_response_alg)
value with the algorithm of the specified key.
{{< /callout >}}

The key id of the [JSON Web Key] used to sign Introspection responses for this client.

To be considered valid:

1. The chosen value must have a [JSON Web Key] configured in the [jwks] section with the use value `sig` and a
   matching `kid` to this value.

### introspection_signed_response_alg

{{< confkey type="string" default="none" required="no" >}}

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
A majority of clients will not support this option with any value other than `none` as this enables encoding the
[Introspection Response](https://datatracker.ietf.org/doc/html/rfc7662#section-2.2) as a JWT as per
[JWT Response for OAuth Token Introspection](https://www.ietf.org/archive/id/draft-ietf-oauth-jwt-introspection-response-12.html)
i.e. rather than being a JSON document the
[Introspection Response](https://datatracker.ietf.org/doc/html/rfc7662#section-2.2) becomes a signed JWT in the
`application/token-introspection+jwt` format and the signed JWT is optionally nested within an encrypted JWT.
{{< /callout >}}

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
This value is completely ignored if the
[introspection_signed_response_key_id](#introspection_signed_response_key_id) is defined.
{{< /callout >}}

The algorithm of the [JSON Web Key] used to sign Introspection responses for this client.
To be considered valid with exclusion of the value `none`:

1. The chosen value must have a [JSON Web Key] configured in the [jwks] section with the use value `sig` and a
   matching `alg` to this value.

See the response object section of the
[integration guide](../../../integration/openid-connect/introduction.md#response-object) for more information. The
supported values come from the algorithm column with a use of `sig`.

### introspection_encrypted_response_key_id

{{< confkey type="string" required="no" >}}

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
A majority of clients will not support this option with any value other than `none` as this enables encoding the
[Introspection Response](https://datatracker.ietf.org/doc/html/rfc7662#section-2.2) as a JWT as per
[JWT Response for OAuth Token Introspection](https://www.ietf.org/archive/id/draft-ietf-oauth-jwt-introspection-response-12.html)
i.e. rather than being a JSON document the
[Introspection Response](https://datatracker.ietf.org/doc/html/rfc7662#section-2.2) becomes a signed JWT in the
`application/token-introspection+jwt` format and the signed JWT is optionally nested within an encrypted JWT.
{{< /callout >}}

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
This value is completely ignored if the
[introspection_encrypted_response_alg](#introspection_encrypted_response_alg) is defined.
{{< /callout >}}

The key id of the [JSON Web Key] used to encrypt Introspection responses for this client.

To be considered valid:

1. The chosen value must have a [JSON Web Key] configured in the [jwks] section with the use value `enc` and a
   matching `kid` to this value.

### introspection_encrypted_response_alg

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
A majority of clients will not support this option with any value other than `none` as this enables encoding the
[Introspection Response](https://datatracker.ietf.org/doc/html/rfc7662#section-2.2) as a JWT as per
[JWT Response for OAuth Token Introspection](https://www.ietf.org/archive/id/draft-ietf-oauth-jwt-introspection-response-12.html)
i.e. rather than being a JSON document the
[Introspection Response](https://datatracker.ietf.org/doc/html/rfc7662#section-2.2) becomes a signed JWT in the
`application/token-introspection+jwt` format and the signed JWT is optionally nested within an encrypted JWT.
{{< /callout >}}

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
This value is completely ignored if the
[introspection_encrypted_response_key_id](#introspection_encrypted_response_key_id) is defined.
{{< /callout >}}

The key algorithm of the [JSON Web Key] used to encrypt Introspection responses for this client.

To be considered valid with exclusion of the value `none`:

1. The chosen value must have a [JSON Web Key] configured in the [jwks] section with the use value `enc` and a
   matching `alg` to this value.
2. The [introspection_signed_response_alg](#introspection_signed_response_alg) or
   [introspection_response_key_id](#introspection_signed_response_key_id) option must be configured.

See the response object section of the
[integration guide](../../../integration/openid-connect/introduction.md#response-object) for more information. The
supported values come from the algorithm column with a use of `enc`.

### introspection_encrypted_response_enc

{{< confkey type="string" default="A128CBC-HS256" required="no" >}}

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
A majority of clients will not support this option with any value other than `none` as this enables encoding the
[Introspection Response](https://datatracker.ietf.org/doc/html/rfc7662#section-2.2) as a JWT as per
[JWT Response for OAuth Token Introspection](https://www.ietf.org/archive/id/draft-ietf-oauth-jwt-introspection-response-12.html)
i.e. rather than being a JSON document the
[Introspection Response](https://datatracker.ietf.org/doc/html/rfc7662#section-2.2) becomes a signed JWT in the
`application/token-introspection+jwt` format and the signed JWT is optionally nested within an encrypted JWT.
{{< /callout >}}

The content encryption algorithm used to encrypt the authorization responses.

See the encryption algorithms section of the
[integration guide](../../../integration/openid-connect/introduction.md#encryption-algorithms) for more information
including the algorithm column for supported values.

### request_object_signing_alg

{{< confkey type="string" default="RS256" required="no" >}}

The JWT signing algorithm accepted for request objects.

See the request object section of the
[integration guide](../../../integration/openid-connect/introduction.md#request-object) for more information including
the algorithm column for supported values.

### request_object_encryption_alg

{{< confkey type="string" required="no" >}}

The JWT content encryption algorithm accepted for request objects.

See the request object section of the
[integration guide](../../../integration/openid-connect/introduction.md#request-object) for more information including
the algorithm column for supported values.

### request_object_encryption_enc

{{< confkey type="string" required="no" >}}

The JWT encryption algorithm accepted for request objects.

See the request object section of the
[integration guide](../../../integration/openid-connect/introduction.md#request-object) for more information including
the algorithm column for supported values.

### token_endpoint_auth_method

{{< confkey type="string" default="client_secret_basic" required="no" >}}

The registered client authentication mechanism used by this client for the [Token Endpoint]. If no method is defined
the confidential client type will default to `client_secret_basic` as this is required by the specification. The public
client type defaults to `none` as this is required by the specification. Supported values are `client_secret_basic`,
`client_secret_post`, `client_secret_jwt`, `private_key_jwt`, and `none`.

See the [integration guide](../../../integration/openid-connect/introduction.md#client-authentication-method) for
more information.

### token_endpoint_auth_signing_alg

{{< confkey type="string" default="RS256" required="no" >}}

The JWT signing algorithm accepted when the [token_endpoint_auth_method](#token_endpoint_auth_method) is configured as
`client_secret_jwt` or `private_key_jwt`.

See the request object section of the [integration guide](../../../integration/openid-connect/introduction.md#request-object)
for more information including the algorithm column for supported values.

It's recommended that you specifically configure this when the following options are configured to specific values
otherwise we assume the default value:

|                   Configuration Option                    |        Value        | Default |
|:---------------------------------------------------------:|:-------------------:|:-------:|
| [token_endpoint_auth_method](#token_endpoint_auth_method) |  `private_key_jwt`  | `RS256` |
| [token_endpoint_auth_method](#token_endpoint_auth_method) | `client_secret_jwt` | `HS256` |

### revocation_endpoint_auth_method

{{< confkey type="string" default="client_secret_basic" required="no" >}}

The registered client authentication mechanism used by this client for the [Revocation Endpoint]. If no method is defined
the confidential client type will default to `client_secret_basic` as this is required by the specification. The public
client type defaults to `none` as this is required by the specification. Supported values are `client_secret_basic`,
`client_secret_post`, `client_secret_jwt`, `private_key_jwt`, and `none`.

See the [integration guide](../../../integration/openid-connect/introduction.md#client-authentication-method) for
more information.

### revocation_endpoint_auth_signing_alg

{{< confkey type="string" default="RS256" required="no" >}}

The JWT signing algorithm accepted when the [revocation_endpoint_auth_method](#revocation_endpoint_auth_method) is
configured as `client_secret_jwt` or `private_key_jwt`.

See the request object section of the [integration guide](../../../integration/openid-connect/introduction.md#request-object)
for more information including the algorithm column for supported values.

It's recommended that you specifically configure this when the following options are configured to specific values
otherwise we assume the default value:

|                        Configuration Option                         |        Value        | Default |
|:-------------------------------------------------------------------:|:-------------------:|:-------:|
| [revocation_endpoint_auth_method](#revocation_endpoint_auth_method) |  `private_key_jwt`  | `RS256` |
| [revocation_endpoint_auth_method](#revocation_endpoint_auth_method) | `client_secret_jwt` | `HS256` |

### introspection_endpoint_auth_method

{{< confkey type="string" default="client_secret_basic" required="no" >}}

The registered client authentication mechanism used by this client for the [Introspection Endpoint]. If no method is
defined the confidential client type will default to `client_secret_basic` as this is required by the specification. The
public client type defaults to `none` as this is required by the specification. Supported values are
`client_secret_basic`, `client_secret_post`, `client_secret_jwt`, `private_key_jwt`, and `none`.

See the [integration guide](../../../integration/openid-connect/introduction.md#client-authentication-method) for
more information.

### introspection_endpoint_auth_signing_alg

{{< confkey type="string" default="RS256" required="no" >}}

The JWT signing algorithm accepted when the [introspection_endpoint_auth_method](#introspection_endpoint_auth_method) is
configured as `client_secret_jwt` or `private_key_jwt`.

See the request object section of the [integration guide](../../../integration/openid-connect/introduction.md#request-object)
for more information including the algorithm column for supported values.

It's recommended that you specifically configure this when the following options are configured to specific values
otherwise we assume the default value:

|                           Configuration Option                            |        Value        | Default |
|:-------------------------------------------------------------------------:|:-------------------:|:-------:|
| [introspection_endpoint_auth_method](#introspection_endpoint_auth_method) |  `private_key_jwt`  | `RS256` |
| [introspection_endpoint_auth_method](#introspection_endpoint_auth_method) | `client_secret_jwt` | `HS256` |

### pushed_authorization_request_endpoint_auth_method

{{< confkey type="string" default="client_secret_basic" required="no" >}}

The registered client authentication mechanism used by this client for the [Pushed Authorization Request Endpoint]. If
no method is defined the confidential client type will default to `client_secret_basic` as this is required by the
specification. The public client type defaults to `none` as this is required by the specification. Supported values are
`client_secret_basic`, `client_secret_post`, `client_secret_jwt`, `private_key_jwt`, and `none`.

See the [integration guide](../../../integration/openid-connect/introduction.md#client-authentication-method) for
more information.

### pushed_authorization_request_endpoint_auth_signing_alg

{{< confkey type="string" default="RS256" required="no" >}}

The JWT signing algorithm accepted when the
[pushed_authorization_request_endpoint_auth_method](#pushed_authorization_request_endpoint_auth_method) is configured as
`client_secret_jwt` or `private_key_jwt`.

See the request object section of the [integration guide](../../../integration/openid-connect/introduction.md#request-object)
for more information including the algorithm column for supported values.

It's recommended that you specifically configure this when the following options are configured to specific values
otherwise we assume the default value:

|                                          Configuration Option                                           |        Value        | Default |
|:-------------------------------------------------------------------------------------------------------:|:-------------------:|:-------:|
| [pushed_authorization_request_endpoint_auth_method](#pushed_authorization_request_endpoint_auth_method) |  `private_key_jwt`  | `RS256` |
| [pushed_authorization_request_endpoint_auth_method](#pushed_authorization_request_endpoint_auth_method) | `client_secret_jwt` | `HS256` |

### allow_multiple_auth_methods

{{< confkey type="boolean" default="false" required="no" >}}

[RFC6749: Section 2.3](https://datatracker.ietf.org/doc/html/rfc6749#section-2.3) clearly indicates that clients have no
option but to use a single authentication method in any single request. Authelia by default enforces this behavior, this
is an escape hatch to turn this policy off for a particular client.

Per the text:

{{< callout context="danger" title="RFC6749: Section 2.3" icon="outline/alert-octagon" >}}
The client MUST NOT use more than one authentication method in each request.
{{< /callout >}}

### jwks_uri

{{< confkey type="string" required="situational" >}}

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
The URL given in this value MUST be resolvable by Authelia and MUST present a certificate signed by
a certificate trusted by your environment. It is beyond our intentions to support anything other than this.
{{< /callout >}}

The fully qualified, `https` scheme, and appropriately signed URI for the JWKs endpoint that implements
[RFC7517 Section 5](https://datatracker.ietf.org/doc/html/rfc7517#section-5).This is mutually exclusive with [jwks](#jwks), meaning they must not be configured at the same
time. It's recommended that you configure this option to account for key rotation instead of [jwks](#jwks).

This option or the [jwks](#jwks) option configures the trusted JSON Web Keys or JWKs for this registered client.
This section is situationally required. These are used to validate the [JWT] assertions from clients.

Required when the following options are configured:

- [request_object_signing_alg](#request_object_signing_alg)
- [token_endpoint_auth_signing_alg](#token_endpoint_auth_signing_alg)

Required when the following options are configured to specific values:

- [token_endpoint_auth_method](#token_endpoint_auth_method): `private_key_jwt`

The following is a contextual example (see below for information regarding each option):

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    clients:
      - client_id: 'example'
        jwks_uri: 'https://oidc.{{< sitevar name="domain" nojs="example.com" >}}:8080/oauth2/jwks.json'
        jwks:
          - key_id: 'example'
            algorithm: 'RS256'
            use: 'sig'
            key: |
              -----BEGIN RSA PUBLIC KEY-----
              ...
              -----END RSA PUBLIC KEY-----
            certificate_chain: |
              -----BEGIN CERTIFICATE-----
              ...
              -----END CERTIFICATE-----
              -----BEGIN CERTIFICATE-----
              ...
              -----END CERTIFICATE-----
```

### jwks

{{< confkey type="list(object)" required="situational" >}}

A list of manually configured JSON Web Keys. This is mutually exclusive with [jwks_uri](#jwks_uri), meaning they must
not be configured at the same time. It's recommended that you configure the [jwks_uri](#jwks_uri) option to account for
key rotation instead of this option.

This option or the [jwks_uri](#jwks_uri) option configures the trusted JSON Web Keys or JWKs for this registered client.
This section is situationally required. These are used to validate the [JWT] assertions from clients.

Required when the following options are configured:

- [request_object_signing_alg](#request_object_signing_alg)
- [token_endpoint_auth_signing_alg](#token_endpoint_auth_signing_alg)

Required when the following options are configured to specific values:

- [token_endpoint_auth_method](#token_endpoint_auth_method): `private_key_jwt`

#### key_id

{{< confkey type="string" required="yes" >}}

The Key ID used to match the request object's JWT header `kid` value against.

#### use

{{< confkey type="string" default="sig" required="no" >}}

The key usage. Defaults to `sig` which is the only available option at this time.

#### algorithm

{{< confkey type="string" default="RS256" required="situational" >}}

The algorithm for this key. This value typically optional as it can be automatically detected based on the type of key
in some situations. It is however strongly recommended this is set.

See the request object table in the [integration guide](../../../integration/openid-connect/introduction.md#request-object)
for more information. The `Algorithm` column lists supported values, the `Key` column references the required
[key](#key) type constraints that exist for the algorithm, and the `JWK Default Conditions` column briefly explains the
conditions under which it's the default algorithm.

#### key

{{< confkey type="string" required="yes" >}}

The public key portion of the JSON Web Key.

The public key the clients use to sign/encrypt the [OpenID Connect 1.0] asserted [JWT]'s. The key is generated by the
client application or the administrator of the client application.

The key *__MUST__*:

* Be a PEM block encoded in the DER base64 format ([RFC4648]).
* Be either:
  * An RSA public key:
    * Encoded in conformance to the [PKCS#8] or [PKCS#1] specifications.
    * Have a key size of at least 2048 bits.
  * An ECDSA public key:
    * Encoded in conformance to the [PKCS#8] or [SECG1] specifications.
    * Use one of the following elliptical curves:
      * P-256.
      * P-384.
      * P-512.

[PKCS#8]: https://datatracker.ietf.org/doc/html/rfc5208
[PKCS#1]: https://datatracker.ietf.org/doc/html/rfc8017
[SECG1]: https://datatracker.ietf.org/doc/html/rfc5915

If the [certificate_chain](#certificate_chain) is provided the private key must include matching public
key data for the first certificate in the chain.

#### certificate_chain

{{< confkey type="string" required="no" >}}

The certificate chain/bundle to be used with the [key](#key) DER base64 ([RFC4648])
encoded PEM format used to sign/encrypt the [OpenID Connect 1.0] [JWT]'s.

## Integration

To integrate Authelia's [OpenID Connect 1.0] implementation with a relying party please see the
[integration docs](../../../integration/openid-connect/introduction.md).

[token lifespan]: https://docs.apigee.com/api-platform/antipatterns/oauth-long-expiration
[OpenID Connect 1.0]: https://openid.net/connect/
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
[jwks]: provider.md#jwks
[JSON Web Key]: provider.md#jwks
