---
title: "authelia crypto rand"
description: "Reference for the authelia crypto rand command."
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

## authelia crypto rand

Generate a cryptographically secure random string

### Synopsis

Generate a cryptographically secure random string.

This subcommand allows generating cryptographically secure random strings for use for encryption keys, HMAC keys, etc.

```
authelia crypto rand [flags]
```

### Examples

```
authelia crypto rand --help
authelia crypto rand --length 80
authelia crypto rand -n 80
authelia crypto rand --charset alphanumeric
authelia crypto rand --charset alphabetic
authelia crypto rand --charset ascii
authelia crypto rand --charset numeric
authelia crypto rand --charset numeric-hex
authelia crypto rand --characters 0123456789ABCDEF
```

### Options

```
      --characters string   sets the explicit characters for the random string
  -x, --charset string      sets the charset for the random password, options are 'ascii', 'alphanumeric', 'alphabetic', 'numeric', 'numeric-hex', and 'rfc3986' (default "alphanumeric")
  -h, --help                help for rand
  -n, --length int          sets the character length for the random string (default 72)
```

### Options inherited from parent commands

```
  -c, --config strings                        configuration files to load (default [configuration.yml])
      --config.directory string               path to a directory with yml/yaml files to load as part of the configuration
      --config.experimental.filters strings   Applies filters in order to the configuration file before the YAML parser. Options are 'template', 'expand-env'
```

### SEE ALSO

* [authelia crypto](authelia_crypto.md)	 - Perform cryptographic operations

