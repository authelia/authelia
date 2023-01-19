---
title: "authelia crypto hash validate"
description: "Reference for the authelia crypto hash validate command."
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

## authelia crypto hash validate

Perform cryptographic hash validations

### Synopsis

Perform cryptographic hash validations.

This subcommand allows preforming cryptographic hash validations. i.e. checking hash digests against a password.

```
authelia crypto hash validate [flags] -- <digest>
```

### Examples

```
authelia crypto hash validate --help
authelia crypto hash validate '$5$rounds=500000$WFjMpdCQxIkbNl0k$M0qZaZoK8Gwdh8Cw5diHgGfe5pE0iJvxcVG3.CVnQe.' -- 'p@ssw0rd'
```

### Options

```
  -h, --help              help for validate
      --password string   manually supply the password rather than using the terminal prompt
```

### Options inherited from parent commands

```
  -c, --config strings                        configuration files or directories to load, for more information run 'authelia -h authelia config' (default [configuration.yml])
      --config.experimental.filters strings   list of filters to apply to all configuration files, for more information run 'authelia -h authelia filters'
```

### SEE ALSO

* [authelia crypto hash](authelia_crypto_hash.md)	 - Perform cryptographic hash operations

