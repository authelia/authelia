---
title: "authelia crypto"
description: "Reference for the authelia crypto command."
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

## authelia crypto

Perform cryptographic operations

### Synopsis

Perform cryptographic operations.

This subcommand allows preforming cryptographic certificate, key pair, etc tasks.

### Examples

```
authelia crypto --help
```

### Options

```
  -h, --help   help for crypto
```

### Options inherited from parent commands

```
  -c, --config strings                        configuration files or directories to load (default [configuration.yml])
      --config.experimental.filters strings   list of filters to apply to all configuration files, the filters are applied after loading them from disk and before parsing their content, options are 'template', 'expand-env'
```

### SEE ALSO

* [authelia](authelia.md)	 - authelia untagged-unknown-dirty (master, unknown)
* [authelia crypto certificate](authelia_crypto_certificate.md)	 - Perform certificate cryptographic operations
* [authelia crypto hash](authelia_crypto_hash.md)	 - Perform cryptographic hash operations
* [authelia crypto pair](authelia_crypto_pair.md)	 - Perform key pair cryptographic operations
* [authelia crypto rand](authelia_crypto_rand.md)	 - Generate a cryptographically secure random string

