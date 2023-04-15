---
title: "Frequently Asked Questions"
description: "Frequently Asked Questions regarding integrating the Authelia OpenID Connect Provider with an OpenID Connect relying party"
lead: "Frequently Asked Questions regarding integrating the Authelia OpenID Connect Provider with an OpenID Connect relying party."
date: 2022-10-20T15:27:09+11:00
draft: false
images: []
menu:
  integration:
    parent: "openid-connect"
weight: 615
toc: true
---

## How do I generate client secrets?

We strongly recommend the following guidelines for generating client secrets:

1. Each client should have a unique secret.
2. Each secret should be randomly generated.
3. Each secret should have a length above 40 characters.
4. The secrets should be stored in the configuration in a supported hash format. *__Note:__ This does not mean you
   configure the relying party / client application with the hashed version, just the secret value in the Authelia
   configuration.*
5. Secrets should only have alphanumeric characters as some implementations do not appropriately encode the secret
   when using it to access the token endpoint.

Authelia provides an easy way to perform such actions via the [Generating a Random Password Hash] guide. Users can
perform a command such as
`authelia crypto hash generate pbkdf2 --variant sha512 --random --random.length 72 --random-charset rfc3986` command to
both generate a client secret with 72 characters which is printed and is to be used with the relying party and hash it
using PBKDF2 which can be stored in the Authelia configuration. This random command also avoids issues with a relying
party / client application encoding the characters correctly as it uses the
[RFC3986 Unreserved Characters](https://datatracker.ietf.org/doc/html/rfc3986#section-2.3).

[Generating a Random Password Hash]: ../../reference/guides/generating-secure-values.md#generating-a-random-password-hash

### Plaintext

Authelia *technically* supports storing the plaintext secret in the configuration. This will likely be completely
unavailable in the future as it was a mistake to implement it like this in the first place. While some other OpenID
Connect 1.0 providers operate in this way, it's more often than not that they operating in this way in error. The
current *technical support* for this is only to prevent massive upheaval to users and give them time to migrate.

As per [RFC6819 Section 5.1.4.1.3](https://datatracker.ietf.org/doc/html/rfc6819#section-5.1.4.1.3) the secret should
only be stored by the authorization server as hashes / digests unless there is a very specific specification or protocol
that is implemented by the authorization server which requires access to the secret in the clear to operate properly in
which case the secret should be encrypted and not be stored in plaintext. The most likely long term outcome is that the
client configurations will be stored in the database with the secret both salted and peppered.

Authelia currently does not implement any of the specifications or protocols which require secrets being accessible in
the clear such as most notably the `client_secret_jwt` grant, we will however likely soon implement `client_secret_jwt`.
We are however *__strongly discouraging__* and formally deprecating the use of plaintext client secrets for purposes
outside those required by specifications. We instead recommended that users remove this from their configuration
entirely and use the [How Do I Generate Client Secrets](#how-do-i-generate-client-secrets) FAQ.

Plaintext is either denoted by the `$plaintext$` prefix where everything after the prefix is the secret. In addition if
the secret does not start with the `$` character it's considered as a plaintext secret for the time being but is
deprecated as is the `$plaintext$` prefix.

## Why isn't my application able to retrieve the token even though I've consented?

The most common cause for this issue is when the affected application can not make requests to the Token [Endpoint].
This becomes obvious when the log level is set to `debug` or `trace` and a presence of requests to the Authorization
[Endpoint] without errors but an absence of requests made to the Token [Endpoint].

These requests can be identified by looking at the `path` field in the logs, or by messages prefixed with
`Authorization Request` indicating a request to the Authorization [Endpoint] and `Access Request` indicating a request
to the Token [Endpoint].

All causes should be clearly logged by the client application, and all errors that do not match this scenario are
clearly logged by Authelia. It's not possible for us to log requests that never occur however.

[Endpoint]: ./introduction.md#discoverable-endpoints
