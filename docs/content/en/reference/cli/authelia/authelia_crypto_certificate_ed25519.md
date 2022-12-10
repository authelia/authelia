---
title: "authelia crypto certificate ed25519"
description: "Reference for the authelia crypto certificate ed25519 command."
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

## authelia crypto certificate ed25519

Perform Ed25519 certificate cryptographic operations

### Synopsis

Perform Ed25519 certificate cryptographic operations.

This subcommand allows preforming Ed25519 certificate cryptographic tasks.

### Examples

```
authelia crypto certificate ed25519 --help
```

### Options

```
  -h, --help   help for ed25519
```

### Options inherited from parent commands

```
  -c, --config strings                        configuration files to load (default [configuration.yml])
      --config.directory string               path to a directory with yml/yaml files to load as part of the configuration
      --config.experimental.filters strings   Applies filters in order to the configuration file before the YAML parser. Options are 'template', 'expand-env'
```

### SEE ALSO

* [authelia crypto certificate](authelia_crypto_certificate.md)	 - Perform certificate cryptographic operations
* [authelia crypto certificate ed25519 generate](authelia_crypto_certificate_ed25519_generate.md)	 - Generate an Ed25519 private key and certificate
* [authelia crypto certificate ed25519 request](authelia_crypto_certificate_ed25519_request.md)	 - Generate an Ed25519 private key and certificate signing request

