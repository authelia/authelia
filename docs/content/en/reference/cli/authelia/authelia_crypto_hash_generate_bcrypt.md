---
title: "authelia crypto hash generate bcrypt"
description: "Reference for the authelia crypto hash generate bcrypt command."
lead: ""
date: 2022-10-17T21:51:59+11:00
draft: false
images: []
menu:
  reference:
    parent: "cli-authelia"
weight: 330
toc: true
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
  -c, --config strings      configuration files to load (default [configuration.yml])
  -i, --cost int            hashing cost (default 12)
  -h, --help                help for bcrypt
      --no-confirm          skip the password confirmation prompt
      --password string     manually supply the password rather than using the terminal prompt
      --random              uses a randomly generated password
      --random.length int   when using a randomly generated password it configures the length (default 72)
  -v, --variant string      variant, options are 'standard' and 'sha256' (default "standard")
```

### SEE ALSO

* [authelia crypto hash generate](authelia_crypto_hash_generate.md)	 - Generate cryptographic hash digests

