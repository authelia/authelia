---
title: "authelia crypto pair ecdsa generate"
description: "Reference for the authelia crypto pair ecdsa generate command."
lead: ""
date: 2022-06-27T18:27:57+10:00
draft: false
images: []
menu:
  reference:
    parent: "cli-authelia"
weight: 905
toc: true
---

## authelia crypto pair ecdsa generate

Generate a cryptographic ECDSA key pair

### Synopsis

Generate a cryptographic ECDSA key pair.

This subcommand allows generating an ECDSA key pair.

```
authelia crypto pair ecdsa generate [flags]
```

### Examples

```
authelia crypto pair ecdsa generate --help
```

### Options

```
  -b, --curve string              Sets the elliptic curve which can be P224, P256, P384, or P521 (default "P256")
  -d, --directory string          directory where the generated keys, certificates, etc will be stored
      --file.private-key string   name of the file to export the private key data to (default "private.pem")
      --file.public-key string    name of the file to export the public key data to (default "public.pem")
  -h, --help                      help for generate
      --pkcs8                     force PKCS #8 ASN.1 format
```

### Options inherited from parent commands

```
  -c, --config strings                        configuration files to load (default [configuration.yml])
      --config.experimental.filters strings   applies filters in order to the configuration file before the YAML parser, options are 'template', 'expand-env'
```

### SEE ALSO

* [authelia crypto pair ecdsa](authelia_crypto_pair_ecdsa.md)	 - Perform ECDSA key pair cryptographic operations

