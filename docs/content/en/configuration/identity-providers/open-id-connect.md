---
title: "OpenID Connect"
description: "OpenID Connect Configuration"
lead: "Authelia can operate as an OpenID Connect provider. This section describes how to configure this."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  configuration:
    parent: "identity-providers"
weight: 190200
toc: true
aliases:
  - /c/oidc
  - /docs/configuration/identity-providers/oidc.html
---

__Authelia__ currently supports the [OpenID Connect] OP role as a [__beta__](../../roadmap/active/openid-connect.md)
feature. The OP role is the [OpenID Connect] Provider role, not the Relying Party or RP role. This means other
applications that implement the [OpenID Connect] RP role can use Authelia as an authentication and authorization backend
similar to how you may use social media or development platforms for login.

The Relying Party role is the role which allows an application to use GitHub, Google, or other [OpenID Connect]
providers for authentication and authorization. We do not intend to support this functionality at this moment in time.

More information about the beta can be found in the [roadmap](../../roadmap/active/openid-connect.md).

## Configuration

The following snippet provides a sample-configuration for the OIDC identity provider explaining each field in detail.

```yaml
identity_providers:
  oidc:
    hmac_secret: this_is_a_secret_abc123abc123abc
    issuer_private_key: |
      -----BEGIN RSA PRIVATE KEY-----
      MXIEogIB$AKCAQEAxZVJP3WF//PG2fLQoEC9DtdiFG/+00vqlbVzz47nyxKONIPI
      lmL3UdmqpGTKMe/5Brqse4ZAKlQHiDbwzK9ypnfigtHuvh/JO0S7ChP70RC67ed1
      HV1nyfz5eW3llbtGJPrlYLqITNgctHp6zmRUFtSzPj9qFvozI93LJi492yL1+vu8
      Un3Dm8+Qq6XM2tPdEcldB/dtBwOWoF+8eOOVsu0TDuB5bwlhBVGJuSAuzBPRS2bF
      Ga4uk0JDdkDOMCEQxC5uWDFxgfERSMFyfLVWD47woDbuWEBq10c0z+dpWPMp7Ain
      YnnkqicwCN88Z0zid6MmMQ65F4+9Hc+qC/p6xwIDAQABAoIBAGlhaAHKor+Su3o/
      AXqXTL5/rbYMzbLQiLt0XeJT69jpeqMTroZXHmWvXE3128mqnf0yzw/K2Ko6yxGh
      i+j/onya8FqpsVYCCgfsbn2/js1AyRJeIp6Y1ORsYnqbXJnxmkXa80AV/OBPW2/+
      60TtSdQrebY3iFPc+i2k+9bPTvpyyDLKlz8UwdZG+k5uyYNIyQTccz+PjwsIvDij
      7tKYamhhLN3QXt3/aZTFpjTgezP4WyriZxjWrddHowc47q2rwNS95ND39JcysJAc
      0Pcbu8A5lVa7Fx33uOtzDfKWIW7xVEN+OtPgN+FbTjXcXk5IZedl+pW5lU5P++G/
      ZPvz+WECgYEA9g6HwdODW3e68bOqsFoKg35+vfUFMzlyMF8HFylNVfnLpTEDr637
      owzMFvcUxVd71b+gV5nnnbI+riUFIgyR8vhCjhy4moopDPahC4/KwN4NG6uz+i1h
      AB6D5+zn2BjnO/5xMMFGlApWtRNmJVGYlNDj3bXKh2VXzzy03VNeD8kCgYEAzZFL
      OlzoRB1HKpTWIECcuvxofMxLOLb3zs0k2t/FYNYIpovmGWCCAULz13y53e5+/+5m
      7I9VUZJFaIhaZ36qVBApCKdru69pZMkWCcQO9jELFcx51Ez7OgJWzu7GS1QJCPKC
      fEDxI0rZK21j93/Sl/nUnEir7CYpQ+wvCaGuHg8CgYAXgbncfY1+DokwkB6NbHy2
      pT4Mfbz6cNGE538w6kQ2I4AeDvmwLentYMqaow478CinegAiflSPTzkHwAemghbr
      ZGZPV1UXhn13fJRUG2+eT1hnPVcbXnx223N0k8Bud6qXo65CnyRT/kzcTbcjd5Eh
      Hne2daicmMTzynPo9Q72aQKBgBmobO9X8VWvIdbaxO85oVZlctVA2pK1o7CYQmVf
      UM+JZ4MCKzI3rYJizPS0iK5+ujNPmmEkcs2/qBIoEsCgOrpLWhPOcc/3UPxXbPzD
      D+sCrBOIdhxdj23qJNOnUfDNCGOpgUfpAzAYg4q8GKInvi1h7XukRnEvQi9MJ4LY
      P1dZAoGASGcGnTMkmeSXP8ux+dvQJAiJskn/sJIgBZ5uq5GRCeLBUosRSVxM75UK
      vAh/c/RBj+pYXVKuPuHGZCQJxsdcRXzXNGouUtgbaYML5Me/Hagt20QzDRBfuGBg
      qeZBJaXhjElvw6PUWtg4x+LYRCBpq/bS3LK3ozZrSTukVkKDegw=
      -----END RSA PRIVATE KEY-----
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
        pre_configured_consent_duration: ''
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

{{< confkey type="string" required="yes" >}}

*__Important Note:__ This can also be defined using a [secret](../methods/secrets.md) which is __strongly recommended__
especially for containerized deployments.*

The HMAC secret used to sign the [JWT]'s. The provided string is hashed to a SHA256 ([RFC6234]) byte string for the
purpose of meeting the required format.

It's __strongly recommended__ this is a
[Random Alphanumeric String](../miscellaneous/guides.md#generating-a-random-alphanumeric-string) with 64 or more
characters.

### issuer_private_key

{{< confkey type="string" required="yes" >}}

*__Important Note:__ This can also be defined using a [secret](../methods/secrets.md) which is __strongly recommended__
especially for containerized deployments.*

The private key in DER base64 ([RFC4648]) encoded PEM format used to encrypt the [OpenID Connect] [JWT]'s. The key must
be generated by the administrator and can be done by following the
[Generating an RSA Keypair](../miscellaneous/guides.md#generating-an-rsa-keypair) guide.

### access_token_lifespan

{{< confkey type="duration" default="1h" required="no" >}}

The maximum lifetime of an access token. It's generally recommended keeping this short similar to the default.
For more information read these docs about [token lifespan].

### authorize_code_lifespan

{{< confkey type="duration" default="1m" required="no" >}}

The maximum lifetime of an authorize code. This can be rather short, as the authorize code should only be needed to
obtain the other token types. For more information read these docs about [token lifespan].

### id_token_lifespan

{{< confkey type="duration" default="1h" required="no" >}}

The maximum lifetime of an ID token. For more information read these docs about [token lifespan].

### refresh_token_lifespan

{{< confkey type="string" default="90m" required="no" >}}

The maximum lifetime of a refresh token. The
refresh token can be used to obtain new refresh tokens as well as access tokens or id tokens with an
up-to-date expiration. For more information read these docs about [token lifespan].

A good starting point is 50% more or 30 minutes more (which ever is less) time than the highest lifespan out of the
[access token lifespan](#access_token_lifespan), the [authorize code lifespan](#authorize_code_lifespan), and the
[id token lifespan](#id_token_lifespan). For instance the default for all of these is 60 minutes, so the default refresh
token lifespan is 90 minutes.

### enable_client_debug_messages

{{< confkey type="boolean" default="false" required="no" >}}

Allows additional debug messages to be sent to the clients.

### minimum_parameter_entropy

{{< confkey type="integer" default="8" required="no" >}}

This controls the minimum length of the `nonce` and `state` parameters.

*__Security Notice:__* Changing this value is generally discouraged, reducing it from the default can theoretically
make certain scenarios less secure. It is highly encouraged that if your OpenID Connect RP does not send these
parameters or sends parameters with a lower length than the default that they implement a change rather than changing
this value.

### enforce_pkce

{{< confkey type="string" default="public_clients_only" required="no" >}}

[Proof Key for Code Exchange](https://www.rfc-editor.org/rfc/rfc7636.html) enforcement policy: if specified, must be
either `never`, `public_clients_only` or `always`.

If set to `public_clients_only` (default), [PKCE] will be required for public clients using the
[Authorization Code Flow].

When set to `always`, [PKCE] will be required for all clients using the Authorization Code flow.

*__Security Notice:__* Changing this value to `never` is generally discouraged, reducing it from the default can
theoretically make certain client-side applications (mobile applications, SPA) vulnerable to CSRF and authorization code
interception attacks.

### enable_pkce_plain_challenge

{{< confkey type="boolean" default="false" required="no" >}}

Allows [PKCE] `plain` challenges when set to `true`.

*__Security Notice:__* Changing this value is generally discouraged. Applications should use the `S256` [PKCE] challenge
method instead.

### cors

Some [OpenID Connect] Endpoints need to allow cross-origin resource sharing, however some are optional. This section allows
you to configure the optional parts. We reply with CORS headers when the request includes the Origin header.

#### endpoints

{{< confkey type="list(string)" required="no" >}}

A list of endpoints to configure with cross-origin resource sharing headers. It is recommended that the `userinfo`
option is at least in this list. The potential endpoints which this can be enabled on are as follows:

* authorization
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
[OpenID Connect] endpoints, however this is not recommended.

Origins must only have the scheme, hostname and port, they may not have a trailing slash or path.

In addition to an Origin URI, you may specify the wildcard origin in the allowed_origins. It MUST be specified by itself
and the [allowed_origins_from_client_redirect_uris](#allowed_origins_from_client_redirect_uris) MUST NOT be enabled. The
wildcard origin is denoted as `*`. Examples:

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

{{< confkey type="boolean" default="false" required="no" >}}

Automatically adds the origin portion of all redirect URI's on all clients to the list of
[allowed_origins](#allowed_origins), provided they have the scheme http or https and do not have the hostname of
localhost.

### clients

{{< confkey type="list" required="yes" >}}

A list of clients to configure. The options for each client are described below.

#### id

{{< confkey type="string" required="yes" >}}

The Client ID for this client. It must exactly match the Client ID configured in the application
consuming this client.

#### description

{{< confkey type="string" default="*same as id*" required="no" >}}

A friendly description for this client shown in the UI. This defaults to the same as the ID.

#### secret

{{< confkey type="string" required="situational" >}}

The shared secret between Authelia and the application consuming this client. This secret must match the secret
configured in the application. Currently this is stored in plain text.

This secret must be generated by the administrator and can be done by following the
[Generating a Random Alphanumeric String](../miscellaneous/guides.md#generating-a-random-alphanumeric-string) guide.

This must be provided when the client is a confidential client type, and must be blank when using the public client
type. To set the client type to public see the [public](#public) configuration option.

#### sector_identifier

{{< confkey type="string" required="no" >}}

*__Important Note:__ because adjusting this option will inevitably change the `sub` claim of all tokens generated for
the specified client, changing this should cause the relying party to detect all future authorizations as completely new
users.*

Must be an empty string or the host component of a URL. This is commonly just the domain name, but may also include a
port.

Authelia utilizes UUID version 4 subject identifiers. By default the public [Subject Identifier Type] is utilized for
all clients. This means the subject identifiers will be the same for all clients. This configuration option enables
[Pairwise Identifier Algorithm] for this client, and configures the sector identifier utilized for both the storage and
the lookup of the subject identifier.

1. All clients who do not have this configured will generate the same subject identifier for a particular user
   regardless of which client obtains the ID token.
2. All clients which have the same sector identifier will:
   1. have the same subject identifier for a particular user when compared to clients with the same sector identifier.
   2. have a completely different subject identifier for a particular user whe compared to:
      1. any client with the public subject identifier type.
      2. any client with a differing sector identifier.

In specific but limited scenarios this option is beneficial for privacy reasons. In particular this is useful when the
party utilizing the *Authelia* [OpenID Connect] Authorization Server is foreign and not controlled by the user. It would
prevent the third party utilizing the subject identifier with another third party in order to track the user.

Keep in mind depending on the other claims they may still be able to perform this tracking and it is not a silver
bullet. There are very few benefits when utilizing this in a homelab or business where no third party is utilizing
the server.

#### public

{{< confkey type="bool" default="false" required="no" >}}

This enables the public client type for this client. This is for clients that are not capable of maintaining
confidentiality of credentials, you can read more about client types in [RFC6749 Section 2.1]. This is particularly
useful for SPA's and CLI tools. This option requires setting the [client secret](#secret) to a blank string.

In addition to the standard rules for redirect URIs, public clients can use the `urn:ietf:wg:oauth:2.0:oob` redirect
URI.

#### authorization_policy

{{< confkey type="string" default="two_factor" required="no" >}}

The authorization policy for this client: either `one_factor` or `two_factor`.

#### pre_configured_consent_duration

{{< confkey type="duration" required="no" >}}

*__Note:__ This setting uses the [duration notation format](../prologue/common.md#duration-notation-format). Please see
the [common options](../prologue/common.md#duration-notation-format) documentation for information on this format.*

Configuring this enables users of this client to remember their consent as a pre-configured consent. The period of time
dictates how long a users choice to remember the pre-configured consent lasts.

Pre-configured consents are only valid if the subject, client id are exactly the same and the requested scopes/audience
match exactly with the granted scopes/audience.

#### audience

{{< confkey type="list(string)" required="no" >}}

A list of audiences this client is allowed to request.

#### scopes

{{< confkey type="list(string)" default="openid, groups, profile, email" required="no" >}}

A list of scopes to allow this client to consume. See
[scope definitions](../../integration/openid-connect/introduction.md#scope-definitions) for more information. The
documentation for the application you want to use with Authelia will most-likely provide you with the scopes to allow.

#### redirect_uris

{{< confkey type="list(string)" required="yes" >}}

A list of valid callback URIs this client will redirect to. All other callbacks will be considered unsafe. The URIs are
case-sensitive and they differ from application to application - the community has provided
[a list of URL´s for common applications](../../integration/openid-connect/introduction.md).

Some restrictions that have been placed on clients and
their redirect URIs are as follows:

1. If a client attempts to authorize with Authelia and its redirect URI is not listed in the client configuration the
   attempt to authorize wil fail and an error will be generated.
2. The redirect URIs are case-sensitive.
3. The URI must include a scheme and that scheme must be one of `http` or `https`.
4. The client can ignore rule 3 and use `urn:ietf:wg:oauth:2.0:oob` if it is a [public](#public) client type.

#### grant_types

{{< confkey type="list(string)" default="refresh_token, authorization_code" required="no" >}}

A list of grant types this client can return. *It is recommended that this isn't configured at this time unless you
know what you're doing*. Valid options are: `implicit`, `refresh_token`, `authorization_code`, `password`,
`client_credentials`.

#### response_types

{{< confkey type="list(string)" default="code" required="no" >}}

A list of response types this client can return. *It is recommended that this isn't configured at this time unless you
know what you're doing*. Valid options are: `code`, `code id_token`, `id_token`, `token id_token`, `token`,
`token id_token code`.

#### response_modes

{{< confkey type="list(string)" default="form_post, query, fragment" required="no" >}}

A list of response modes this client can return. It is recommended that this isn't configured at this time unless you
know what you're doing. Potential values are `form_post`, `query`, and `fragment`.

#### userinfo_signing_algorithm

{{< confkey type="string" default="none" required="no" >}}

The algorithm used to sign the userinfo endpoint responses. This can either be `none` or `RS256`.

## Integration

To integrate Authelia's [OpenID Connect] implementation with a relying party please see the
[integration docs](../../integration/openid-connect/introduction.md).

[token lifespan]: https://docs.apigee.com/api-platform/antipatterns/oauth-long-expiration
[OpenID Connect]: https://openid.net/connect/
[JWT]: https://www.rfc-editor.org/rfc/rfc7519.html
[RFC6234]: https://www.rfc-editor.org/rfc/rfc6234.html
[RFC4648]: https://www.rfc-editor.org/rfc/rfc4648.html
[RFC7468]: https://www.rfc-editor.org/rfc/rfc7468.html
[RFC6749 Section 2.1]: https://www.rfc-editor.org/rfc/rfc6749.html#section-2.1
[PKCE]: https://www.rfc-editor.org/rfc/rfc7636.html
[Authorization Code Flow]: https://openid.net/specs/openid-connect-core-1_0.html#CodeFlowAuth
[Subject Identifier Type]: https://openid.net/specs/openid-connect-core-1_0.html#SubjectIDTypes
[Pairwise Identifier Algorithm]: https://openid.net/specs/openid-connect-core-1_0.html#PairwiseAlg
