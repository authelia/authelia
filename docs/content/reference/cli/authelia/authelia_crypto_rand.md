---
title: "authelia crypto rand"
description: "Reference for the authelia crypto rand command."
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
authelia crypto rand directory/file1 directory/file2
authelia crypto rand --file directory/file3,directory/file4
```

### Options

```
      --characters string   sets the explicit characters for the random string
  -x, --charset string      sets the charset for the random password, options are 'ascii', 'alphanumeric', 'alphabetic', 'numeric', 'numeric-hex', and 'rfc3986' (default "alphanumeric")
      --file strings        files to export the unique random values to instead of printing them
  -h, --help                help for rand
  -n, --length int          sets the character length for the random string (default 72)
      --mode-dirs string    mode for the created directories (default "0700")
      --mode-files string   mode for the created files (default "0600")
```

### Options inherited from parent commands

```
  -c, --config strings                        configuration files or directories to load, for more information run 'authelia -h authelia config' (default [configuration.yml])
      --config.experimental.filters strings   list of filters to apply to all configuration files, for more information run 'authelia -h authelia filters'
```

### SEE ALSO

* [authelia crypto](authelia_crypto.md)	 - Perform cryptographic operations

