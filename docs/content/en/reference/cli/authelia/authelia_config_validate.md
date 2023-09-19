---
title: "authelia config validate"
description: "Reference for the authelia config validate command."
lead: ""
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  reference:
    parent: "cli-authelia"
weight: 905
toc: true
---

## authelia config validate

Check a configuration against the internal configuration validation mechanisms

### Synopsis

Check a configuration against the internal configuration validation mechanisms.

This subcommand allows validation of the YAML and Environment configurations so that a configuration can be checked
prior to deploying it.

```
authelia config validate [flags]
```

### Examples

```
authelia config validate
authelia config validate --config config.yml
```

### Options

```
  -h, --help   help for validate
```

### Options inherited from parent commands

```
  -c, --config strings                        configuration files or directories to load, for more information run 'authelia -h authelia config' (default [configuration.yml])
      --config.experimental.filters strings   list of filters to apply to all configuration files, for more information run 'authelia -h authelia filters'
```

### SEE ALSO

* [authelia config](authelia_config.md)	 - Perform config related actions

