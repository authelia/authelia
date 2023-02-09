---
title: "docs/content/en/reference/cli/authelia/authelia crypto"
description: "Reference for the docs/content/en/reference/cli/authelia/authelia crypto command."
lead: ""
date: 2022-06-27T18:27:57+10:00
draft: false
images: []
menu:
  reference:
    parent: "cli-docs/content/en/reference/cli/authelia/authelia"
weight: 995
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
  -c, --config strings                        configuration files or directories to load, for more information run 'authelia -h authelia config' (default [configuration.yml])
      --config.experimental.filters strings   list of filters to apply to all configuration files, for more information run 'authelia -h authelia filters'
```

### SEE ALSO

* [authelia](authelia.md)	 - authelia untagged-unknown-dirty (master, unknown)
* [authelia crypto certificate](authelia_crypto_certificate.md)	 - Perform certificate cryptographic operations
* [authelia crypto hash](authelia_crypto_hash.md)	 - Perform cryptographic hash operations
* [authelia crypto pair](authelia_crypto_pair.md)	 - Perform key pair cryptographic operations
* [authelia crypto rand](authelia_crypto_rand.md)	 - Generate a cryptographically secure random string

