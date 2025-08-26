---
title: "authelia crypto hash generate argon2"
description: "Reference for the authelia crypto hash generate argon2 command."
lead: ""
date: 2025-08-01T16:23:47+10:00
draft: false
images: []
weight: 905
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## authelia crypto hash generate argon2

Generate cryptographic Argon2 hash digests

### Synopsis

Generate cryptographic Argon2 hash digests.

This subcommand allows generating cryptographic Argon2 hash digests.

```
authelia crypto hash generate argon2 [flags]
```

### Examples

```
authelia crypto hash generate argon2 --help
```

### Options

```
  -h, --help              help for argon2
  -i, --iterations int    number of iterations (default 3)
  -k, --key-size int      key size in bytes (default 32)
  -m, --memory int        memory in kibibytes (default 65536)
  -p, --parallelism int   parallelism or threads (default 4)
      --profile string    profile to use, options are low-memory and recommended
  -s, --salt-size int     salt size in bytes (default 16)
  -v, --variant string    variant, options are 'argon2id', 'argon2i', and 'argon2d' (default "argon2id")
```

### Options inherited from parent commands

```
  -c, --config strings                        configuration files or directories to load, for more information run 'authelia -h authelia config' (default [configuration.yml])
      --config.experimental.filters strings   list of filters to apply to all configuration files, for more information run 'authelia -h authelia filters'
      --no-confirm                            skip the password confirmation prompt
      --password string                       manually supply the password rather than using the terminal prompt
      --random                                uses a randomly generated password
      --random.characters string              sets the explicit characters for the random string
      --random.charset string                 sets the charset for the random password, options are 'ascii', 'alphanumeric', 'alphabetic', 'numeric', 'numeric-hex', and 'rfc3986' (default "alphanumeric")
      --random.length int                     sets the character length for the random string (default 72)
```

### SEE ALSO

* [authelia crypto hash generate](authelia_crypto_hash_generate.md)	 - Generate cryptographic hash digests

