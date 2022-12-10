---
title: "authelia crypto hash generate scrypt"
description: "Reference for the authelia crypto hash generate scrypt command."
lead: ""
date: 2022-10-17T21:51:59+11:00
draft: false
images: []
menu:
  reference:
    parent: "cli-authelia"
weight: 905
toc: true
---

## authelia crypto hash generate scrypt

Generate cryptographic scrypt hash digests

### Synopsis

Generate cryptographic scrypt hash digests.

This subcommand allows generating cryptographic scrypt hash digests.

```
authelia crypto hash generate scrypt [flags]
```

### Examples

```
authelia crypto hash generate scrypt --help
```

### Options

```
  -r, --block-size int    block size (default 8)
  -h, --help              help for scrypt
  -i, --iterations int    number of iterations (default 16)
  -k, --key-size int      key size in bytes (default 32)
  -p, --parallelism int   parallelism or threads (default 1)
  -s, --salt-size int     salt size in bytes (default 16)
```

### Options inherited from parent commands

```
  -c, --config strings             configuration files to load (default [configuration.yml])
      --config.directory string    path to a directory with yml/yaml files to load as part of the configuration
      --no-confirm                 skip the password confirmation prompt
      --password string            manually supply the password rather than using the terminal prompt
      --random                     uses a randomly generated password
      --random.characters string   sets the explicit characters for the random string
      --random.charset string      sets the charset for the random password, options are 'ascii', 'alphanumeric', 'alphabetic', 'numeric', 'numeric-hex', and 'rfc3986' (default "alphanumeric")
      --random.length int          sets the character length for the random string (default 72)
```

### SEE ALSO

* [authelia crypto hash generate](authelia_crypto_hash_generate.md)	 - Generate cryptographic hash digests

