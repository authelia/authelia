---
title: "OpenID Connect 1.0 Provider"
description: "OpenID Connect 1.0 Provider Configuration"
lead: "Authelia can operate as an OpenID Connect 1.0 Provider. This section describes how to configure this."
date: 2023-05-08T13:38:08+10:00
draft: false
images: []
menu:
  configuration:
    parent: "openid-connect"
weight: 190200
toc: true
aliases:
  - /c/oidc
  - /docs/configuration/identity-providers/oidc.html
---

__Authelia__ currently supports the [OpenID Connect 1.0] Provider role as an open
[__beta__](../../../roadmap/active/openid-connect.md) feature. We currently do not support the [OpenID Connect 1.0] Relying
Party role. This means other applications that implement the [OpenID Connect 1.0] Relying Party role can use Authelia as
an [OpenID Connect 1.0] Provider similar to how you may use social media or development platforms for login.

The [OpenID Connect 1.0] Relying Party role is the role which allows an application to use GitHub, Google, or other
[OpenID Connect 1.0] Providers for authentication and authorization. We do not intend to support this functionality at
this moment in time.

This section covers the [OpenID Connect 1.0] Provider configuration. For information on configuring individual
registered clients see the [OpenID Connect 1.0 Clients](clients.md) documentation.

More information about the beta can be found in the [roadmap](../../../roadmap/active/openid-connect.md) and in the
[integration](../../../integration/openid-connect/introduction.md) documentation.

## Configuration

The following snippet provides a configuration example for the [OpenID Connect 1.0] Provider. This is not
intended for production use it's used to provide context and an indentation example.

```yaml
identity_providers:
  oidc:
    hmac_secret: this_is_a_secret_abc123abc123abc
    issuer_private_keys:
      - key_id: example
        algorithm: RS256
        use: sig
        key: |
          -----BEGIN RSA PUBLIC KEY-----
          MEgCQQDAwV26ZA1lodtOQxNrJ491gWT+VzFum9IeZ+WTmMypYWyW1CzXKwsvTHDz
          9ec+jserR3EMQ0Rr24lj13FL1ib5AgMBAAE=
          -----END RSA PUBLIC KEY----
        certificate_chain: |
          -----BEGIN CERTIFICATE-----
          MIIBWzCCAQWgAwIBAgIQYAKsXhJOXKfyySlmpKicTzANBgkqhkiG9w0BAQsFADAT
          MREwDwYDVQQKEwhBdXRoZWxpYTAeFw0yMzA0MjEwMDA3NDRaFw0yNDA0MjAwMDA3
          NDRaMBMxETAPBgNVBAoTCEF1dGhlbGlhMFwwDQYJKoZIhvcNAQEBBQADSwAwSAJB
          AK2i7RlJEYo/Xa6mQmv9zmT0XUj3DcEhRJGPVw2qMyadUFxNg/ZFp7aTcToHMf00
          z6T3b7mwdBkCFQOL3Kb7WRcCAwEAAaM1MDMwDgYDVR0PAQH/BAQDAgWgMBMGA1Ud
          JQQMMAoGCCsGAQUFBwMBMAwGA1UdEwEB/wQCMAAwDQYJKoZIhvcNAQELBQADQQB8
          Of2iM7fPadmtChCMna8lYWH+lEplj6BxOJlRuGRawxszLwi78bnq0sCR33LU6xMx
          1oAPwIHNaJJwC4z6oG9E_DO_NOT_USE=
          -----END CERTIFICATE-----
          -----BEGIN CERTIFICATE-----
          MIIBWzCCAQWgAwIBAgIQYAKsXhJOXKfyySlmpKicTzANBgkqhkiG9w0BAQsFADAT
          MREwDwYDVQQKEwhBdXRoZWxpYTAeFw0yMzA0MjEwMDA3NDRaFw0yNDA0MjAwMDA3
          NDRaMBMxETAPBgNVBAoTCEF1dGhlbGlhMFwwDQYJKoZIhvcNAQEBBQADSwAwSAJB
          AK2i7RlJEYo/Xa6mQmv9zmT0XUj3DcEhRJGPVw2qMyadUFxNg/ZFp7aTcToHMf00
          z6T3b7mwdBkCFQOL3Kb7WRcCAwEAAaM1MDMwDgYDVR0PAQH/BAQDAgWgMBMGA1Ud
          JQQMMAoGCCsGAQUFBwMBMAwGA1UdEwEB/wQCMAAwDQYJKoZIhvcNAQELBQADQQB8
          Of2iM7fPadmtChCMna8lYWH+lEplj6BxOJlRuGRawxszLwi78bnq0sCR33LU6xMx
          1oAPwIHNaJJwC4z6oG9E_DO_NOT_USE=
          -----END CERTIFICATE-----
    issuer_private_key: |
      -----BEGIN RSA PUBLIC KEY-----
      MEgCQQDAwV26ZA1lodtOQxNrJ491gWT+VzFum9IeZ+WTmMypYWyW1CzXKwsvTHDz
      9ec+jserR3EMQ0Rr24lj13FL1ib5AgMBAAE=
      -----END RSA PUBLIC KEY----
    issuer_certificate_chain: |
      -----BEGIN CERTIFICATE-----
      MIIBWzCCAQWgAwIBAgIQYAKsXhJOXKfyySlmpKicTzANBgkqhkiG9w0BAQsFADAT
      MREwDwYDVQQKEwhBdXRoZWxpYTAeFw0yMzA0MjEwMDA3NDRaFw0yNDA0MjAwMDA3
      NDRaMBMxETAPBgNVBAoTCEF1dGhlbGlhMFwwDQYJKoZIhvcNAQEBBQADSwAwSAJB
      AK2i7RlJEYo/Xa6mQmv9zmT0XUj3DcEhRJGPVw2qMyadUFxNg/ZFp7aTcToHMf00
      z6T3b7mwdBkCFQOL3Kb7WRcCAwEAAaM1MDMwDgYDVR0PAQH/BAQDAgWgMBMGA1Ud
      JQQMMAoGCCsGAQUFBwMBMAwGA1UdEwEB/wQCMAAwDQYJKoZIhvcNAQELBQADQQB8
      Of2iM7fPadmtChCMna8lYWH+lEplj6BxOJlRuGRawxszLwi78bnq0sCR33LU6xMx
      1oAPwIHNaJJwC4z6oG9E_DO_NOT_USE=
      -----END CERTIFICATE-----
      -----BEGIN CERTIFICATE-----
      MIIBWzCCAQWgAwIBAgIQYAKsXhJOXKfyySlmpKicTzANBgkqhkiG9w0BAQsFADAT
      MREwDwYDVQQKEwhBdXRoZWxpYTAeFw0yMzA0MjEwMDA3NDRaFw0yNDA0MjAwMDA3
      NDRaMBMxETAPBgNVBAoTCEF1dGhlbGlhMFwwDQYJKoZIhvcNAQEBBQADSwAwSAJB
      AK2i7RlJEYo/Xa6mQmv9zmT0XUj3DcEhRJGPVw2qMyadUFxNg/ZFp7aTcToHMf00
      z6T3b7mwdBkCFQOL3Kb7WRcCAwEAAaM1MDMwDgYDVR0PAQH/BAQDAgWgMBMGA1Ud
      JQQMMAoGCCsGAQUFBwMBMAwGA1UdEwEB/wQCMAAwDQYJKoZIhvcNAQELBQADQQB8
      Of2iM7fPadmtChCMna8lYWH+lEplj6BxOJlRuGRawxszLwi78bnq0sCR33LU6xMx
      1oAPwIHNaJJwC4z6oG9E_DO_NOT_USE=
      -----END CERTIFICATE-----
    access_token_lifespan: 1h
    authorize_code_lifespan: 1m
    id_token_lifespan: 1h
    refresh_token_lifespan: 90m
    enable_client_debug_messages: false
    minimum_parameter_entropy: 8
    enforce_pkce: public_clients_only
    enable_pkce_plain_challenge: false
    pushed_authorizations:
      enforce: false
      context_lifespan: 5m
    cors:
      endpoints:
        - authorization
        - token
        - revocation
        - introspection
      allowed_origins:
        - https://example.com
      allowed_origins_from_client_redirect_uris: false
```

## Options

### hmac_secret

{{< confkey type="string" required="yes" >}}

*__Important Note:__ This can also be defined using a [secret](../../methods/secrets.md) which is __strongly recommended__
especially for containerized deployments.*

The HMAC secret used to sign the [JWT]'s. The provided string is hashed to a SHA256 ([RFC6234]) byte string for the
purpose of meeting the required format.

It's __strongly recommended__ this is a
[Random Alphanumeric String](../../../reference/guides/generating-secure-values.md#generating-a-random-alphanumeric-string)
with 64 or more characters.

### issuer_private_keys

The key *__MUST__*:

* Be a PEM block encoded in the DER base64 format ([RFC4648]).
* Be either:
  * An RSA public key:
    * With a key size of at least 2048 bits.
  * An ECDSA public key with one of:
    * A P-256 elliptical curve.
    * A P-384 elliptical curve.
    * A P-512 elliptical curve.

### issuer_private_keys

{{< confkey type="list(object" required="no" >}}

The list of JWKS instead of or in addition to the [issuer_private_key](#issuerprivatekey) and
[issuer_certificate_chain](#issuercertificatechain). Can also accept ECDSA Private Key's and Certificates.

#### key_id

{{< confkey type="string" default="<thumbprint of public key>" required="no" >}}

Completely optional, and generally discouraged unless there is a collision between the automatically generated key id's.
If provided must be a unique string with 7 or less alphanumeric characters.

This value is the first 7 characters of the public key thumbprint (SHA1) encoded into hexadecimal.

#### algorithm

{{< confkey type="string" default="RS256" required="no" >}}

The algorithm for this key. This value must be unique. It's automatically detected based on the type of key.

See the response object table in the [integration guide](../../../integration/openid-connect/introduction.md#response-object)
including the algorithm column for the supported values and the key type column for the default algorithm value.

#### use

{{< confkey type="string" default="sig" required="no" >}}

The key usage. Defaults to `sig` which is the only available option at this time.

#### key

{{< confkey type="string" required="yes" >}}

The private key associated with this key entry.

The private key used to sign/encrypt the [OpenID Connect 1.0] issued [JWT]'s. The key must be generated by the administrator
and can be done by following the
[Generating an RSA Keypair](../../../reference/guides/generating-secure-values.md#generating-an-rsa-keypair) guide.

The private key *__MUST__*:
* Be a PEM block encoded in the DER base64 format ([RFC4648]).
* Be one of:
  * An RSA key with a key size of at least 2048 bits.
  * An ECDSA private key with one of the P-256, P-384, or P-521 elliptical curves.

If the [certificate_chain](#certificatechain) is provided the private key must include matching public
key data for the first certificate in the chain.

#### certificate_chain

{{< confkey type="string" required="no" >}}

The certificate chain/bundle to be used with the [key](#key) DER base64 ([RFC4648])
encoded PEM format used to sign/encrypt the [OpenID Connect 1.0] [JWT]'s. When configured it enables the [x5c] and [x5t]
JSON key's in the JWKs [Discoverable Endpoint](../../../integration/openid-connect/introduction.md#discoverable-endpoints)
as per [RFC7517].

[RFC7517]: https://datatracker.ietf.org/doc/html/rfc7517
[x5c]: https://datatracker.ietf.org/doc/html/rfc7517#section-4.7
[x5t]: https://datatracker.ietf.org/doc/html/rfc7517#section-4.8

The first certificate in the chain must have the public key for the [key](#key), each certificate in the chain must be
valid for the current date, and each certificate in the chain should be signed by the certificate immediately following
it if present.

### issuer_private_key

{{< confkey type="string" required="yes" >}}

*__Important Note:__ This can also be defined using a [secret](../../methods/secrets.md) which is __strongly recommended__
especially for containerized deployments.*

The private key used to sign/encrypt the [OpenID Connect 1.0] issued [JWT]'s. The key must be generated by the administrator
and can be done by following the
[Generating an RSA Keypair](../../../reference/guides/generating-secure-values.md#generating-an-rsa-keypair) guide.

This private key is automatically appended to the [issuer_private_keys](#issuerprivatekeys) and assumed to be for the
RS256 algorithm. As such no other key in this list should be RS256 if this is configured.

The issuer private key *__MUST__*:

* Be a PEM block encoded in the DER base64 format ([RFC4648]).
* Be an RSA private key:
    * With a key size of at least 2048 bits.

If the [issuer_certificate_chain](#issuercertificatechain) is provided the private key must include matching public
key data for the first certificate in the chain.

### issuer_certificate_chain

{{< confkey type="string" required="no" >}}

The certificate chain/bundle to be used with the [issuer_private_key](#issuer_private_key) DER base64 ([RFC4648])
encoded PEM format used to sign/encrypt the [OpenID Connect 1.0] [JWT]'s. When configured it enables the [x5c] and [x5t]
JSON key's in the JWKs [Discoverable Endpoint](../../../integration/openid-connect/introduction.md#discoverable-endpoints)
as per [RFC7517].

[RFC7517]: https://datatracker.ietf.org/doc/html/rfc7517
[x5c]: https://datatracker.ietf.org/doc/html/rfc7517#section-4.7
[x5t]: https://datatracker.ietf.org/doc/html/rfc7517#section-4.8

The first certificate in the chain must have the public key for the [issuer_private_key](#issuerprivatekey), each
certificate in the chain must be valid for the current date, and each certificate in the chain should be signed by the
certificate immediately following it if present.

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
[access token lifespan](#accesstokenlifespan), the [authorize code lifespan](#authorizecodelifespan), and the
[id token lifespan](#idtokenlifespan). For instance the default for all of these is 60 minutes, so the default refresh
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
[allowed_origins_from_client_redirect_uris](#allowedoriginsfromclientredirecturis) option is enabled. This means
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

See the [OpenID Connect 1.0 Registered Clients](clients.md) documentation for configuring clients.

## Integration

To integrate Authelia's [OpenID Connect 1.0] implementation with a relying party please see the
[integration docs](../../integration/openid-connect/introduction.md).

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
