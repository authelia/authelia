---
title: "authelia crypto hash generate pbkdf2"
description: "Reference for the authelia crypto hash generate pbkdf2 command."
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

## authelia crypto hash generate pbkdf2

Generate cryptographic PBKDF2 hash digests

### Synopsis

Generate cryptographic PBKDF2 hash digests.

This subcommand allows generating cryptographic PBKDF2 hash digests.

```
authelia crypto hash generate pbkdf2 [flags]
```

### Examples

```
authelia crypto hash generate pbkdf2 --help
```

### Options

```
  -h, --help             help for pbkdf2
  -i, --iterations int   number of iterations (default 310000)
  -s, --salt-size int    salt size in bytes (default 16)
  -v, --variant string   variant, options are 'sha1', 'sha224', 'sha256', 'sha384', and 'sha512' (default "sha512")
```

### Options inherited from parent commands

```
  -c, --config strings                        configuration files to load (default [configuration.yml])
      --config.directory string               path to a directory with yml/yaml files to load as part of the configuration
      --config.experimental.filters strings   Applies filters in order to the configuration file before the YAML parser. Options are 'template', 'expand-env'
      --no-confirm                            skip the password confirmation prompt
      --password string                       manually supply the password rather than using the terminal prompt
      --random                                uses a randomly generated password
      --random.characters string              sets the explicit characters for the random string
      --random.charset string                 sets the charset for the random password, options are 'ascii', 'alphanumeric', 'alphabetic', 'numeric', 'numeric-hex', and 'rfc3986' (default "alphanumeric")
      --random.length int                     sets the character length for the random string (default 72)
```

### SEE ALSO

* [authelia crypto hash generate](authelia_crypto_hash_generate.md)	 - Generate cryptographic hash digests

