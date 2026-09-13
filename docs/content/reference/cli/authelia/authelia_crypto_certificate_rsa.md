---
title: "authelia crypto certificate rsa"
description: "Reference for the authelia crypto certificate rsa command."
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

## authelia crypto certificate rsa

Perform RSA certificate cryptographic operations

### Synopsis

Perform RSA certificate cryptographic operations.

This subcommand allows performing RSA certificate cryptographic tasks.

### Examples

```
authelia crypto certificate rsa --help
```

### Options

```
  -h, --help   help for rsa
```

### Options inherited from parent commands

```
  -c, --config strings                        configuration files or directories to load, for more information run 'authelia -h authelia config' (default [configuration.yml])
      --config.experimental.filters strings   list of filters to apply to all configuration files, for more information run 'authelia -h authelia filters'
      --config.filters.values strings         file paths of values files (.yml, .yaml, .json, .toml) to utilize with configuration file filters; files are loaded in order with later files deep-merged on top, for more information run 'authelia -h authelia filters'
```

### SEE ALSO

* [authelia crypto certificate](authelia_crypto_certificate.md)	 - Perform certificate cryptographic operations
* [authelia crypto certificate rsa generate](authelia_crypto_certificate_rsa_generate.md)	 - Generate an RSA private key and certificate
* [authelia crypto certificate rsa request](authelia_crypto_certificate_rsa_request.md)	 - Generate an RSA private key and certificate signing request

