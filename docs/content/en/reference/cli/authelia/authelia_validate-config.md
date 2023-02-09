---
title: "docs/content/en/reference/cli/authelia/authelia validate-config"
description: "Reference for the docs/content/en/reference/cli/authelia/authelia validate-config command."
lead: ""
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  reference:
    parent: "cli-docs/content/en/reference/cli/authelia/authelia"
weight: 995
toc: true
---

## authelia validate-config

Check a configuration against the internal configuration validation mechanisms

### Synopsis

Check a configuration against the internal configuration validation mechanisms.

This subcommand allows validation of the YAML and Environment configurations so that a configuration can be checked
prior to deploying it.

```
authelia validate-config [flags]
```

### Examples

```
authelia validate-config
authelia validate-config --config config.yml
```

### Options

```
  -h, --help   help for validate-config
```

### Options inherited from parent commands

```
  -c, --config strings                        configuration files or directories to load, for more information run 'authelia -h authelia config' (default [configuration.yml])
      --config.experimental.filters strings   list of filters to apply to all configuration files, for more information run 'authelia -h authelia filters'
```

### SEE ALSO

* [authelia](authelia.md)	 - authelia untagged-unknown-dirty (master, unknown)

