---
title: "authelia crypto certificate"
description: "Reference for the authelia crypto certificate command."
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

## authelia crypto certificate

Perform certificate cryptographic operations

### Synopsis

Perform certificate cryptographic operations.

This subcommand allows preforming certificate cryptographic tasks.

### Examples

```
authelia crypto certificate --help
```

### Options

```
  -h, --help   help for certificate
```

### Options inherited from parent commands

```
  -c, --config strings                        configuration files to load (default [configuration.yml])
      --config.directory string               configuration directory to load configuration files from
      --config.experimental.filters strings   list of filters to apply to all configuration files between loading them from disk and parsing their content, options are 'template', 'expand-env'
```

### SEE ALSO

* [authelia crypto](authelia_crypto.md)	 - Perform cryptographic operations
* [authelia crypto certificate ecdsa](authelia_crypto_certificate_ecdsa.md)	 - Perform ECDSA certificate cryptographic operations
* [authelia crypto certificate ed25519](authelia_crypto_certificate_ed25519.md)	 - Perform Ed25519 certificate cryptographic operations
* [authelia crypto certificate rsa](authelia_crypto_certificate_rsa.md)	 - Perform RSA certificate cryptographic operations

