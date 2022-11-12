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

Authelia supports storing the plaintext secret in the configuration. This may be discontinued in the future. Plaintext
is either denoted by the `$plaintext$` prefix where everything after the prefix is the secret. In addition if the secret
does not start with the `$` character it's considered as a plaintext secret for the time being but is deprecated.
