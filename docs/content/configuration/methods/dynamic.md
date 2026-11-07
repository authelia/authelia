---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Dynamic Configuration"
description: "Configuring the sections of Authelia which can be reloaded at runtime without a restart, including how reloads are triggered, validated, and applied to requests."
summary: "Some sections of the configuration can be placed in separate files which are reloaded without restarting Authelia. This section describes using this method."
date: 2026-11-07T14:00:00-08:00
draft: false
images: []
weight: 101500
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

Most of the configuration is only read when Authelia starts. A small number of sections can instead be placed in
separate _dynamic configuration_ files. These are validated and applied while Authelia keeps running, without dropping
connections or restarting any of its services.

The sections which are currently supported are:

- [access_control](../security/access-control.md)
- `definitions.network` (named networks which can be used by the `access_control` rules)

## Configuration

```yaml {title="configuration.yml"}
dynamic_configuration:
  paths:
    - '/config/access-control.yml'
  watch: false
```

```yaml {title="access-control.yml"}
access_control:
  default_policy: 'deny'
  rules:
    - domain: 'app.example.com'
      policy: 'one_factor'
```

When `paths` is configured the sections listed above **must be removed** from the main configuration, otherwise
Authelia refuses to start. When `paths` is not configured these sections work exactly as they always have, in the main
configuration, and are not reloaded.

## Options

### paths

{{< confkey type="list(string)" required="no" >}}

The files or directories to load the dynamic configuration from. The same file formats (YAML, TOML, and JSON), directory
behavior, and [filters](./files.md) as the main configuration files are supported. Environment variables and secrets
are not used for the dynamic configuration. Only the supported sections can be present; keys for any other section are
an error rather than being silently ignored.

### watch

{{< confkey type="boolean" default="false" required="no" >}}

Automatically reloads the dynamic configuration when the files or directories listed in [paths](#paths) change.
Changes which happen close together, such as an editor writing a file several times, are combined and the latest
content of the files is what's loaded. Replacing a file, or the Kubernetes style atomic replacement of a ConfigMap
volume, is detected.

## Reloading

A reload happens when:

1. A file changes and [watch](#watch) is enabled.
1. Authelia receives the `SIGHUP` signal. When the dynamic configuration is in use this reloads _only_ the dynamic
   configuration, instead of the full restart that `SIGHUP` otherwise performs. The main configuration is not reloaded
   and the full restart is not performed; to apply changes to other sections restart Authelia.

On every reload the files are loaded and fully validated, in the same way as at startup. Then:

- If there are **any errors** nothing is changed. The previous configuration stays in use, and the errors are logged at
  the `error` level. Fixing the files and reloading again recovers without a restart.
- If the result is identical to the one in use nothing happens.
- Otherwise requests which arrive from that point onward are evaluated against the new configuration. Requests which
  were already being evaluated finish using the configuration they started with, and a single request never uses a
  mix of the two.

A successful reload logs `Dynamic configuration applied` along with the new generation number, which starts at 1 when
Authelia starts and increases with every applied reload.

### OpenID Connect 1.0 client subjects

The `access_control` rules can reference OpenID Connect 1.0 clients with the `oauth2:client:` subject prefix. On every
reload the main configuration is read again to determine which client identifiers exist, so a rule can only reference
clients which exist in the main configuration at that time. The clients themselves are not reloaded, so adding or
removing one still requires a restart.

## Checking a policy

The `authelia access-control check-policy` command reads the dynamic configuration when it is configured.
