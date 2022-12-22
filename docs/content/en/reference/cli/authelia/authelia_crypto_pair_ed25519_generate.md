---
title: "authelia crypto pair ed25519 generate"
description: "Reference for the authelia crypto pair ed25519 generate command."
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

## authelia crypto pair ed25519 generate

Generate a cryptographic Ed25519 key pair

### Synopsis

Generate a cryptographic Ed25519 key pair.

This subcommand allows generating an Ed25519 key pair.

```
authelia crypto pair ed25519 generate [flags]
```

### Examples

```
authelia crypto pair ed25519 generate --help
```

### Options

```
  -d, --directory string          directory where the generated keys, certificates, etc will be stored
      --file.private-key string   name of the file to export the private key data to (default "private.pem")
      --file.public-key string    name of the file to export the public key data to (default "public.pem")
  -h, --help                      help for generate
      --pkcs8                     force PKCS #8 ASN.1 format
```

### Options inherited from parent commands

```
  -c, --config strings                        configuration files or directories to load (default [configuration.yml])
      --config.experimental.filters strings   list of filters to apply to all configuration files, for more information: authelia --help authelia filters
```

### SEE ALSO

* [authelia crypto pair ed25519](authelia_crypto_pair_ed25519.md)	 - Perform Ed25519 key pair cryptographic operations

