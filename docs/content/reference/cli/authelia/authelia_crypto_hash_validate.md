---
title: "authelia crypto hash validate"
description: "Reference for the authelia crypto hash validate command."
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

## authelia crypto hash validate

Perform cryptographic hash validations

### Synopsis

Perform cryptographic hash validations.

This subcommand allows performing cryptographic hash validations. i.e. checking hash digests against a password.

```
authelia crypto hash validate [flags] -- <digest>
```

### Examples

```
authelia crypto hash validate --help
authelia crypto hash validate --password 'p@ssw0rd' -- '$5$rounds=500000$WFjMpdCQxIkbNl0k$M0qZaZoK8Gwdh8Cw5diHgGfe5pE0iJvxcVG3.CVnQe.'
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

