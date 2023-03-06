---
title: "OpenID Connect"
description: "OpenID Connect Configuration"
lead: "Authelia can operate as an OpenID Connect 1.0 Provider. This section describes how to configure this."
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

__Authelia__ currently supports the [OpenID Connect 1.0] Provider role as an open
[__beta__](../../roadmap/active/openid-connect.md) feature. We currently do not support the [OpenID Connect 1.0] Relying
Party role. This means other applications that implement the [OpenID Connect 1.0] Relying Party role can use Authelia as
an [OpenID Connect 1.0] Provider similar to how you may use social media or development platforms for login.

The [OpenID Connect 1.0] Relying Party role is the role which allows an application to use GitHub, Google, or other
[OpenID Connect 1.0] Providers for authentication and authorization. We do not intend to support this functionality at
this moment in time.

More information about the beta can be found in the [roadmap](../../roadmap/active/openid-connect.md).

## Configuration

The following snippet provides a sample-configuration for the OIDC identity provider explaining each field in detail.

```yaml
identity_providers:
  oidc:
    hmac_secret: this_is_a_secret_abc123abc123abc
    issuer_certificate_chain: |
      -----BEGIN CERTIFICATE-----
      MIIC5jCCAc6gAwIBAgIRAK4Sj7FiN6PXo/urPfO4E7owDQYJKoZIhvcNAQELBQAw
      EzERMA8GA1UEChMIQXV0aGVsaWEwHhcNNzAwMTAxMDAwMDAwWhcNNzEwMTAxMDAw
      MDAwWjATMREwDwYDVQQKEwhBdXRoZWxpYTCCASIwDQYJKoZIhvcNAQEBBQADggEP
      ADCCAQoCggEBAPKv3pSyP4ozGEiVLJ14dIWFCEGEgq7WUMI0SZZqQA2ID0L59U/Q
      /Usyy7uC9gfMUzODTpANtkOjFQcQAsxlR1FOjVBrX5QgjSvXwbQn3DtwMA7XWSl6
      LuYx2rBYSlMSN5UZQm/RxMtXfLK2b51WgEEYDFi+nECSqKzR4R54eOPkBEWRfvuY
      91AMjlhpivg8e4JWkq4LVQUKbmiFYwIdK8XQiN4blY9WwXwJFYs5sQ/UYMwBFi0H
      kWOh7GEjfxgoUOPauIueZSMSlQp7zqAH39N0ZSYb6cS0Npj57QoWZSY3ak87ebcR
      Nf4rCvZLby7LoN7qYCKxmCaDD3x2+NYpWH8CAwEAAaM1MDMwDgYDVR0PAQH/BAQD
      AgWgMBMGA1UdJQQMMAoGCCsGAQUFBwMBMAwGA1UdEwEB/wQCMAAwDQYJKoZIhvcN
      AQELBQADggEBAHSITqIQSNzonFl3DzxHPEzr2hp6peo45buAAtu8FZHoA+U7Icfh
      /ZXjPg7Xz+hgFwM/DTNGXkMWacQA/PaNWvZspgRJf2AXvNbMSs2UQODr7Tbv+Fb4
      lyblmMUNYFMCFVAMU0eIxXAFq2qcwv8UMcQFT0Z/35s6PVOakYnAGGQjTfp5Ljuq
      wsdc/xWmM0cHWube6sdRRUD7SY20KU/kWzl8iFO0VbSSrDf1AlEhnLEkp1SPaxXg
      OdBnl98MeoramNiJ7NT6Jnyb3zZ578fjaWfThiBpagItI8GZmG4s4Ovh2JbheN8i
      ZsjNr9jqHTjhyLVbDRlmJzcqoj4JhbKs6/I^invalid DO NOT USE=
      -----END CERTIFICATE-----
      -----BEGIN CERTIFICATE-----
      MIIDBDCCAeygAwIBAgIRALJsPg21kA0zY4F1wUCIuoMwDQYJKoZIhvcNAQELBQAw
      EzERMA8GA1UEChMIQXV0aGVsaWEwHhcNNzAwMTAxMDAwMDAwWhcNNzEwMTAxMDAw
      MDAwWjATMREwDwYDVQQKEwhBdXRoZWxpYTCCASIwDQYJKoZIhvcNAQEBBQADggEP
      ADCCAQoCggEBAMXHBvVxUzYk0u34/DINMSF+uiOekKOAjOrC6Mi9Ww8ytPVO7t2S
      zfTvM+XnEJqkFQFgimERfG/eGhjF9XIEY6LtnXe8ATvOK4nTwdufzBaoeQu3Gd50
      5VXr6OHRo//ErrGvFXwP3g8xLePABsi/fkH3oDN+ztewOBMDzpd+KgTrk8ysv2ou
      kNRMKFZZqASvCgv0LD5KWvUCnL6wgf1oTXG7aztduA4oSkUP321GpOmBC5+5ElU7
      ysoRzvD12o9QJ/IfEaulIX06w9yVMo60C/h6A3U6GdkT1SiyTIqR7v7KU/IWd/Qi
      Lfftcj91VhCmJ73Meff2e2S2PrpjdXbG5FMCAwEAAaNTMFEwDgYDVR0PAQH/BAQD
      AgKkMA8GA1UdJQQIMAYGBFUdJQAwDwYDVR0TAQH/BAUwAwEB/zAdBgNVHQ4EFgQU
      Z7AtA3mzFc0InSBA5fiMfeLXA3owDQYJKoZIhvcNAQELBQADggEBAEE5hm1mtlk/
      kviCoHH4evbpw7rxPxDftIQlqYTtvMM4eWY/6icFoSZ4fUHEWYyps8SsPu/8f2tf
      71LGgZn0FdHi1QU2H8m0HHK7TFw+5Q6RLrLdSyk0PItJ71s9en7r8pX820nAFEHZ
      HkOSfJZ7B5hFgUDkMtVM6bardXAhoqcMk4YCU96e9d4PB4eI+xGc+mNuYvov3RbB
      D0s8ICyojeyPVLerz4wHjZu68Z5frAzhZ68YbzNs8j2fIBKKHkHyLG1iQyF+LJVj
      2PjCP+auJsj6fQQpMGoyGtpLcSDh+ptcTngUD8JsWipzTCjmaNqdPHAOYmcgtf4b
      qocikt3WAdU^invalid DO NOT USE=
      -----END CERTIFICATE-----
    issuer_private_key: |
      -----BEGIN RSA PRIVATE KEY-----
      MIIEpAIBAAKCAQEA8q/elLI/ijMYSJUsnXh0hYUIQYSCrtZQwjRJlmpADYgPQvn1
      T9D9SzLLu4L2B8xTM4NOkA22Q6MVBxACzGVHUU6NUGtflCCNK9fBtCfcO3AwDtdZ
      KXou5jHasFhKUxI3lRlCb9HEy1d8srZvnVaAQRgMWL6cQJKorNHhHnh44+QERZF+
      +5j3UAyOWGmK+Dx7glaSrgtVBQpuaIVjAh0rxdCI3huVj1bBfAkVizmxD9RgzAEW
      LQeRY6HsYSN/GChQ49q4i55lIxKVCnvOoAff03RlJhvpxLQ2mPntChZlJjdqTzt5
      txE1/isK9ktvLsug3upgIrGYJoMPfHb41ilYfwIDAQABAoIBAQDTOdFf2JjHH1um
      aPgRAvNf9v7Nj5jytaRKs5nM6iNf46ls4QPreXnMhqSeSwj6lpNgBYxOgzC9Q+cc
      Y4ob/paJJPaIJTxmP8K/gyWcOQlNToL1l+eJ20eQoZm23NGr5fIsunSBwLEpTrdB
      ENqqtcwhW937K8Pxy/Q1nuLyU2bc6Tn/ivLozc8n27dpQWWKh8537VY7ancIaACr
      LJJLYxKqhQpjtBWAyCDvZQirnAOm9KnvIHaGXIswCZ4Xbsu0Y9NL+woARPyRVQvG
      jfxy4EmO9s1s6y7OObSukwKDSNihAKHx/VIbvVWx8g2Lv5fGOa+J2Y7o9Qurs8t5
      BQwMTt0BAoGBAPUw5Z32EszNepAeV3E2mPFUc5CLiqAxagZJuNDO2pKtyN29ETTR
      Ma4O1cWtGb6RqcNNN/Iukfkdk27Q5nC9VJSUUPYelOLc1WYOoUf6oKRzE72dkMQV
      R4bf6TkjD+OVR17fAfkswkGahZ5XA7j48KIQ+YC4jbnYKSxZTYyKPjH/AoGBAP1i
      tqXt36OVlP+y84wWqZSjMelBIVa9phDVGJmmhz3i1cMni8eLpJzWecA3pfnG6Tm9
      ze5M4whASleEt+M00gEvNaU9ND+z0wBfi+/DwJYIbv8PQdGrBiZFrPhTPjGQUldR
      lXccV2meeLZv7TagVxSi3DO6dSJfSEHyemd5j9mBAoGAX8Hv+0gOQZQCSOTAq8Nx
      6dZcp9gHlNaXnMsP9eTDckOSzh636JPGvj6m+GPJSSbkURUIQ3oyokMNwFqvlNos
      fTaLhAOfjBZI9WnDTTQxpugWjphJ4HqbC67JC/qIiw5S6FdaEvGLEEoD4zoChywZ
      9oGAn+fz2d/0/JAH/FpFPgsCgYEAp/ipZgPzziiZ9ov1wbdAQcWRj7RaWnssPFpX
      jXwEiXT3CgEMO4MJ4+KWIWOChrti3qFBg6i6lDyyS6Qyls7sLFbUdC7HlTcrOEMe
      rBoTcCI1GqZNlqWOVQ65ZIEiaI7o1vPBZo2GMQEZuq8mDKFsOMThvvTrM5cAep84
      n6HJR4ECgYABWcbsSnr0MKvVth/inxjbKapbZnp2HUCuw87Ie5zK2Of/tbC20wwk
      yKw3vrGoE3O1t1g2m2tn8UGGASeZ842jZWjIODdSi5+icysQGuULKt86h/woz2SQ
      27GoE2i5mh6Yez6VAYbUuns3FcwIsMyWLq043Tu2DNkx9ijOOAuQzw^invalid..
      DO NOT USE==
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
        secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        sector_identifier: ''
        public: false
        authorization_policy: two_factor
        consent_mode: explicit
        pre_configured_consent_duration: 1w
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
[Random Alphanumeric String](../../reference/guides/generating-secure-values.md#generating-a-random-alphanumeric-string)
with 64 or more characters.

### issuer_certificate_chain

{{< confkey type="string" required="no" >}}

The certificate chain/bundle to be used with the [issuer_private_key](#issuer_private_key) DER base64 ([RFC4648])
encoded PEM format used to sign/encrypt the [OpenID Connect 1.0] [JWT]'s. When configured it enables the [x5c] and [x5t]
JSON key's in the JWKs [Discoverable Endpoint](../../integration/openid-connect/introduction.md#discoverable-endpoints)
as per [RFC7517].

[RFC7517]: https://datatracker.ietf.org/doc/html/rfc7517
[x5c]: https://datatracker.ietf.org/doc/html/rfc7517#section-4.7
[x5t]: https://datatracker.ietf.org/doc/html/rfc7517#section-4.8

The first certificate in the chain must have the public key for the [issuer_private_key](#issuerprivatekey), each
certificate in the chain must be valid for the current date, and each certificate in the chain should be signed by the
certificate immediately following it if present.

### issuer_private_key

{{< confkey type="string" required="yes" >}}

*__Important Note:__ This can also be defined using a [secret](../methods/secrets.md) which is __strongly recommended__
especially for containerized deployments.*

The private key used to sign/encrypt the [OpenID Connect 1.0] issued [JWT]'s. The key must be generated by the administrator
and can be done by following the
[Generating an RSA Keypair](../../reference/guides/generating-secure-values.md#generating-an-rsa-keypair) guide.

The private key *__MUST__*:
* Be a PEM block encoded in the DER base64 format ([RFC4648]).
* Be an RSA Key.
* Have a key size of at least 2048 bits.

If the [issuer_certificate_chain](#issuercertificatechain) is provided the private key must include matching public
key data for the first certificate in the chain.

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

[Proof Key for Code Exchange](https://datatracker.ietf.org/doc/html/rfc7636) enforcement policy: if specified, must be
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

### pushed_authorizations

Controls the behaviour of [Pushed Authorization Requests].

#### enforce

{{< confkey type="boolean" default="false" required="no" >}}

When enabled all authorization requests must use the [Pushed Authorization Requests] flow.

#### context_lifespan

{{< confkey type="duration" default="5m" required="no" >}}

The maximum amount of time between the [Pushed Authorization Requests] flow being initiated and the generated
`request_uri` being utilized by a client.

### cors

Some [OpenID Connect 1.0] Endpoints need to allow cross-origin resource sharing, however some are optional. This section allows
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
and the [allowed_origins_from_client_redirect_uris](#allowedoriginsfromclientredirecturis) MUST NOT be enabled. The
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
configured in the application.

This secret must be generated by the administrator and can be done by following the
[Generating Client Secrets](../../integration/openid-connect/specific-information.md#generating-client-secrets) guide.

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
party utilizing the *Authelia* [OpenID Connect 1.0] Authorization Server is foreign and not controlled by the user. It would
prevent the third party utilizing the subject identifier with another third party in order to track the user.

Keep in mind depending on the other claims they may still be able to perform this tracking and it is not a silver
bullet. There are very few benefits when utilizing this in a homelab or business where no third party is utilizing
the server.

#### public

{{< confkey type="bool" default="false" required="no" >}}

This enables the public client type for this client. This is for clients that are not capable of maintaining
confidentiality of credentials, you can read more about client types in [RFC6749 Section 2.1]. This is particularly
useful for SPA's and CLI tools. This option requires setting the [client secret](#secret) to a blank string.

#### redirect_uris

{{< confkey type="list(string)" required="yes" >}}

A list of valid callback URIs this client will redirect to. All other callbacks will be considered unsafe. The URIs are
case-sensitive and they differ from application to application - the community has provided
[a list of URL´s for common applications](../../integration/openid-connect/introduction.md).

Some restrictions that have been placed on clients and
their redirect URIs are as follows:

1. If a client attempts to authorize with Authelia and its redirect URI is not listed in the client configuration the
   attempt to authorize will fail and an error will be generated.
2. The redirect URIs are case-sensitive.
3. The URI must include a scheme and that scheme must be one of `http` or `https`.

#### audience

{{< confkey type="list(string)" required="no" >}}

A list of audiences this client is allowed to request.

#### scopes

{{< confkey type="list(string)" default="openid, groups, profile, email" required="no" >}}

A list of scopes to allow this client to consume. See
[scope definitions](../../integration/openid-connect/introduction.md#scope-definitions) for more information. The
documentation for the application you are trying to configure [OpenID Connect 1.0] for will likely have a list of scopes
or claims required which can be matched with the above guide.

#### grant_types

{{< confkey type="list(string)" default="refresh_token, authorization_code" required="no" >}}

*__Important Note:__ It is recommended that this isn't configured at this time unless you know what you're doing.*

The list of grant types this client is permitted to use in order to obtain access to the relevant tokens.

See the [Grant Types](../../integration/openid-connect/introduction.md#grant-types) section of the
[OpenID Connect 1.0 Integration Guide](../../integration/openid-connect/introduction.md#grant-types) for more information.

#### response_types

{{< confkey type="list(string)" default="code" required="no" >}}

*__Important Note:__ It is recommended that this isn't configured at this time unless you know what you're doing.*

A list of response types this client supports.

See the [Response Types](../../integration/openid-connect/introduction.md#response-types) section of the
[OpenID Connect 1.0 Integration Guide](../../integration/openid-connect/introduction.md#response-types) for more information.

#### response_modes

{{< confkey type="list(string)" default="form_post, query, fragment" required="no" >}}

*__Important Note:__ It is recommended that this isn't configured at this time unless you know what you're doing.*

A list of response modes this client supports.

See the [Response Modes](../../integration/openid-connect/introduction.md#response-modes) section of the
[OpenID Connect 1.0 Integration Guide](../../integration/openid-connect/introduction.md#response-modes) for more information.

#### authorization_policy

{{< confkey type="string" default="two_factor" required="no" >}}

The authorization policy for this client: either `one_factor` or `two_factor`.

#### enforce_par

{{< confkey type="boolean" default="false" required="no" >}}

Enforces the use of a [Pushed Authorization Requests] flow for this client.

#### enforce_pkce

{{< confkey type="bool" default="false" required="no" >}}

This setting enforces the use of [PKCE] for this individual client. To enforce it for all clients see the global
[enforce_pkce](#enforcepkce) setting.

#### pkce_challenge_method

{{< confkey type="string" default="" required="no" >}}

This setting enforces the use of the specified [PKCE] challenge method for this individual client. This setting also
effectively enables the [enforce_pkce](#enforcepkce-1) option for this client.

Valid values are an empty string, `plain`, or `S256`. It should be noted that `S256` is strongly recommended if the
relying party supports it.

#### userinfo_signing_algorithm

{{< confkey type="string" default="none" required="no" >}}

The algorithm used to sign the userinfo endpoint responses. This can either be `none` or `RS256`.

See the [integration guide](../../integration/openid-connect/introduction.md#user-information-signing-algorithm) for
more information.

#### consent_mode

{{< confkey type="string" default="auto" required="no" >}}

*__Important Note:__ the `implicit` consent mode is not technically part of the specification. It theoretically could be
misused in certain conditions specifically with the public client type or when the client credentials (i.e. client
secret) has been exposed to an attacker. For these reasons this mode is discouraged.*

Configures the consent mode. The following table describes the different modes:

|     Value      |                                                                  Description                                                                   |
|:--------------:|:----------------------------------------------------------------------------------------------------------------------------------------------:|
|      auto      | Automatically determined (default). Uses `explicit` unless [pre_configured_consent_duration] is specified in which case uses `pre-configured`. |
|    explicit    |                                   Requires the user provide unique explicit consent for every authorization.                                   |
|    implicit    |                   Automatically assumes consent for every authorization, never asking the user if they wish to give consent.                   |
| pre-configured |                            Allows the end-user to remember their consent for the [pre_configured_consent_duration].                            |

[pre_configured_consent_duration]: #preconfiguredconsentduration

#### pre_configured_consent_duration

{{< confkey type="duration" default="1w" required="no" >}}

*__Note:__ This setting uses the [duration notation format](../prologue/common.md#duration-notation-format). Please see
the [common options](../prologue/common.md#duration-notation-format) documentation for information on this format.*

Specifying this in the configuration without a consent [consent_mode] enables the `pre-configured` mode. If this is
specified as well as the [consent_mode] then it only has an effect if the [consent_mode] is `pre-configured` or `auto`.

The period of time dictates how long a users choice to remember the pre-configured consent lasts.

Pre-configured consents are only valid if the subject, client id are exactly the same and the requested scopes/audience
match exactly with the granted scopes/audience.

[consent_mode]: #consentmode

## Integration

To integrate Authelia's [OpenID Connect 1.0] implementation with a relying party please see the
[integration docs](../../integration/openid-connect/introduction.md).

[token lifespan]: https://docs.apigee.com/api-platform/antipatterns/oauth-long-expiration
[OpenID Connect 1.0]: https://openid.net/connect/
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
