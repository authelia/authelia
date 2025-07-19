---
title: "authelia config"
description: "Reference for the authelia config command."
lead: ""
date: 2024-03-14T06:00:14+11:00
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

## authelia config

Perform config related actions

### Synopsis

Perform config related actions.

This subcommand contains other subcommands related to the configuration.

```
authelia config [flags]
```

### Examples

```
authelia config --help
```

### Options

```
  -h, --help   help for config
```

### Options inherited from parent commands

```
  -c, --config strings                        configuration files or directories to load, for more information run 'authelia -h authelia config' (default [configuration.yml])
      --config.experimental.filters strings   list of filters to apply to all configuration files, for more information run 'authelia -h authelia filters'
```

### SEE ALSO

* [authelia](authelia.md)	 - authelia untagged-unknown-dirty (master, unknown)
* [authelia config template](authelia_config_template.md)	 - Template a configuration file or files with enabled filters
* [authelia config validate](authelia_config_validate.md)	 - Check a configuration against the internal configuration validation mechanisms

