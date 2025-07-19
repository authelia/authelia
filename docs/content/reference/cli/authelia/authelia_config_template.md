---
title: "authelia config template"
description: "Reference for the authelia config template command."
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

## authelia config template

Template a configuration file or files with enabled filters

### Synopsis

Template a configuration file or files with enabled filters.

This subcommand allows debugging the filtered YAML files with any of the available filters. It should be noted this
command needs to be executed with the same environment variables and working path as when normally running Authelia to
be useful.

```
authelia config template [flags]
```

### Examples

```
authelia config template --config.experimental.filters=template --config=config.yml
```

### Options

```
  -h, --help   help for template
```

### Options inherited from parent commands

```
  -c, --config strings                        configuration files or directories to load, for more information run 'authelia -h authelia config' (default [configuration.yml])
      --config.experimental.filters strings   list of filters to apply to all configuration files, for more information run 'authelia -h authelia filters'
```

### SEE ALSO

* [authelia config](authelia_config.md)	 - Perform config related actions

