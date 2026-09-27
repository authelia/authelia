---
title: "authelia config validate"
description: "Reference for the authelia config validate command."
lead: ""
date: 2026-09-27T14:52:21+10:00
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
  -c, --config strings                                   configuration files or directories to load, for more information run 'authelia -h authelia config' (default [configuration.yml])
      --config.filters strings                           list of filters to apply to all configuration files, for more information run 'authelia -h authelia filters'
      --config.filters.template.delimiter.left string    sets the left delimiter for the 'template' filter
      --config.filters.template.delimiter.right string   sets the right delimiter for the 'template' filter
      --config.filters.values strings                    file paths of values files (.yml, .yaml, .json, .toml) to utilize with configuration file filters; files are loaded in order with later files deep-merged on top, for more information run 'authelia -h authelia filters'
```

### SEE ALSO

* [authelia config](authelia_config.md)	 - Perform config related actions

