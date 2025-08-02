---
title: "authelia crypto hash generate bcrypt"
description: "Reference for the authelia crypto hash generate bcrypt command."
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

## authelia crypto hash generate bcrypt

Generate cryptographic bcrypt hash digests

### Synopsis

Generate cryptographic bcrypt hash digests.

This subcommand allows generating cryptographic bcrypt hash digests.

```
authelia crypto hash generate bcrypt [flags]
```

### Examples

```
authelia crypto hash generate bcrypt --help
```

### Options

```
  -i, --cost int         hashing cost (default 12)
  -h, --help             help for bcrypt
  -v, --variant string   variant, options are 'standard' and 'sha256' (default "standard")
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

