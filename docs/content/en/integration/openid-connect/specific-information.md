---
title: "Specific Information"
description: "Specific information regarding integrating the Authelia OpenID Connect Provider with an OpenID Connect relying party"
lead: "Specific information regarding integrating the Authelia OpenID Connect Provider with an OpenID Connect relying party."
date: 2022-10-20T15:27:09+11:00
draft: false
images: []
menu:
  integration:
    parent: "openid-connect"
weight: 615
toc: true
---

## Generating Client Secrets

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
perform a command such as `authelia crypto hash generate pbkdf2 --variant sha512 --random --random.length 72` command to
both generate a client secret with 72 characters which is printed and is to be used with the relying party and hash it
using PBKDF2 which can be stored in the Authelia configuration.

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
the clear and currently has no plans to implement any of these. As such it's *__strongly discouraged and heavily
deprecated__* and we instead recommended that users remove this from their configuration entirely and use the
[Generating Client Secrets](#generating-client-secrets) guide.

Plaintext is either denoted by the `$plaintext$` prefix where everything after the prefix is the secret. In addition if
the secret does not start with the `$` character it's considered as a plaintext secret for the time being but is
deprecated as is the `$plaintext$` prefix.

## Frequently Asked Questions

### Why isn't my application able to retrieve the token even though I've consented?

The most common cause for this issue is when the affected application can not make requests to the Token [Endpoint].
This becomes obvious when the log level is set to `debug` or `trace` and a presence of requests to the Authorization
[Endpoint] without errors but an absence of requests made to the Token [Endpoint].

These requests can be identified by looking at the `path` field in the logs, or by messages prefixed with
`Authorization Request` indicating a request to the Authorization [Endpoint] and `Access Request` indicating a request
to the Token [Endpoint].

All causes should be clearly logged by the client application, and all errors that do not match this scenario are
clearly logged by Authelia. It's not possible for us to log requests that never occur however.

[Endpoint]: ./introduction.md#discoverable-endpoints
