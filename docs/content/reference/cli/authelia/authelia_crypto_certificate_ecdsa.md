---
title: "authelia crypto certificate ecdsa"
description: "Reference for the authelia crypto certificate ecdsa command."
lead: ""
date: 2026-09-12T18:52:11+10:00
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

## authelia crypto certificate ecdsa

Perform ECDSA certificate cryptographic operations

### Synopsis

Perform ECDSA certificate cryptographic operations.

This subcommand allows performing ECDSA certificate cryptographic tasks.

### Examples

```
authelia crypto certificate ecdsa --help
```

### Options

```
  -h, --help   help for ecdsa
```

### Options inherited from parent commands

```
  -c, --config strings                        configuration files or directories to load, for more information run 'authelia -h authelia config' (default [configuration.yml])
      --config.experimental.filters strings   list of filters to apply to all configuration files, for more information run 'authelia -h authelia filters'
      --config.filters.values strings         file paths of values files (.yml, .yaml, .json, .toml) to utilize with configuration file filters; files are loaded in order with later files deep-merged on top, for more information run 'authelia -h authelia filters'
```

### SEE ALSO

* [authelia crypto certificate](authelia_crypto_certificate.md)	 - Perform certificate cryptographic operations
* [authelia crypto certificate ecdsa generate](authelia_crypto_certificate_ecdsa_generate.md)	 - Generate an ECDSA private key and certificate
* [authelia crypto certificate ecdsa request](authelia_crypto_certificate_ecdsa_request.md)	 - Generate an ECDSA private key and certificate signing request

