---
title: "Files"
description: "Using the YAML File Configuration Method."
lead: "Authelia can be configured via files. This section describes utilizing this method."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  configuration:
    parent: "methods"
weight: 101200
toc: true
---

## Loading Behaviour

There are several options which affect the loading of files:

|       Name        |            Argument             |                                    Description                                     |
|:-----------------:|:-------------------------------:|:----------------------------------------------------------------------------------:|
| Files/Directories |        `--config`, `-c`         | A list of file or directory (non-recursive) paths to load configuration files from |
|      Filters      | `--config.experimental.filters` |   A list of filters applied to every file from the Files or Directories options    |

__*Note:* when specifying directories and files, the individual files specified must not be within any of the
directories specified.__

__*Note:* when specifying directories, all files from the directory (non-recursive) that have the extensions of known
formats will be loaded. As such all of these files must be valid Authelia configuration files.__

## Formats

The only supported configuration file format is [YAML](#yaml), though an experimental implementation of [TOML](#toml)
exists, it's not expressly supported as it is [experimental](../../policies/versioning.md#exceptions).

It's important that you sufficiently validate your configuration file. While we produce console errors for users in many
misconfiguration scenarios it's not perfect. Each file type has recommended methods for validation.

When a directory is specified the following extensions are loaded:

|    Format     |   Extensions    |
|:-------------:|:---------------:|
| [YAML](#yaml) | `.yml`, `.yaml` |
| [TOML](#toml) | `.tml`, `.toml` |

### YAML

*Authelia* loads `configuration.yml` as the configuration if you just run it. You can override this behaviour with the
following syntax:

```bash
authelia --config config.custom.yml
```

#### YAML Validation

We recommend utilizing [VSCodium](https://vscodium.com/) or [VSCode](https://code.visualstudio.com/), both with the
[YAML Extension](https://open-vsx.org/extension/redhat/vscode-yaml) by RedHat to validate this file type.

## Multiple Configuration Files

You can have multiple configuration files which will be merged in the order specified. If duplicate keys are specified
the last one to be specified is the one that takes precedence. Example:

```bash
authelia --config configuration.yml --config config-acl.yml --config config-other.yml
authelia --config configuration.yml,config-acl.yml,config-other.yml
```

Authelia's configuration files use the YAML format. A template with all possible options can be found at the root of the
repository [here](https://github.com/authelia/authelia/blob/master/config.template.yml).

*__Important Note:__ You should not have configuration sections such as Access Control Rules or OpenID Connect clients
configured in multiple files. If you wish to split these into their own files that is fine, but if you have two files that
specify these sections and expect them to merge properly you are asking for trouble.*

### Container

By default, the container looks for a configuration file at `/config/configuration.yml`.

### Docker

This is an example of how to override the configuration files loaded in docker:

```bash
docker run -d --volume /path/to/config:/config authelia:authelia:latest authelia --config=/config/configuration.yaml --config=/config/configuration.acl.yaml
```

See the [Docker Documentation](https://docs.docker.com/engine/reference/commandline/run/) for more information on the
`docker run` command.

### Docker Compose

An excerpt from a docker compose that allows you to specify multiple configuration files is as follows:

```yaml
version: "3.8"
services:
  authelia:
    container_name: authelia
    image: authelia/authelia:latest
    command:
      - "authelia"
      - "--config=/config/configuration.yaml"
      - "--config=/config/configuration.acl.yaml"

```

See the [compose file reference](https://docs.docker.com/compose/compose-file/compose-file-v3/#command) for more
information.

### Kubernetes

An excerpt from a Kubernetes container that allows you to specify multiple configuration files is as follows:

```yaml
kind: Deployment
apiVersion: apps/v1
metadata:
  name: authelia
  namespace: authelia
  labels:
    app.kubernetes.io/instance: authelia
    app.kubernetes.io/name: authelia
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/instance: authelia
      app.kubernetes.io/name: authelia
  template:
    metadata:
      labels:
        app.kubernetes.io/instance: authelia
        app.kubernetes.io/name: authelia
    spec:
      enableServiceLinks: false
      containers:
        - name: authelia
          image: docker.io/authelia/authelia:fix-missing-head-handler
          command:
            - authelia
          args:
            - '--config=/configuration.yaml'
            - '--config=/configuration.acl.yaml'
```

See the Kubernetes [workloads documentation](https://kubernetes.io/docs/concepts/workloads/pods/#pod-templates) or the
[Container API docs](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#container-v1-core) for more
information.

### TOML

[TOML](https://toml.io/) is experimentally supported. No documentation currently exists so users will have to adapt the
other examples to this format.

## File Filters

Experimental file filters exist which allow modification of all configuration files after reading them from the
filesystem but before parsing their content. These filters are _**NOT**_ covered by our
[Standard Versioning Policy](../../policies/versioning.md). There __*WILL*__ be a point where the name of the CLI
argument or environment variable will change and usage of these will either break or just not work.

The filters are configured as a list of filter names by the `--config.experimental.filters` CLI argument and
`X_AUTHELIA_CONFIG_EXPERIMENTAL_FILTERS` environment variable. We recommend using the environment variable as it ensures
commands executed from the container use the same filters. If both the CLI argument and environment variable are used
the environment variable is completely ignored.

Filters can either be used on their own, in combination, or not at all. The filters are processed in order as they are
defined.

Examples:

```bash
authelia --config config.yml --config.experimental.filters expand-env,template
```

```text
X_AUTHELIA_CONFIG_EXPERIMENTAL_FILTERS=expand-env,template
```

### Expand Environment Variable Filter

The name used to enable this filter is `expand-env`.

This filter is the most common filter type used by many other applications. It is similar to using `envsubst` where it
replaces a string like `$EXAMPLE` or `${EXAMPLE}` with the value of the `EXAMPLE` environment variable.

### Go Template Filter

The name used to enable this filter is `template`.

This filter uses the [Go template engine](https://pkg.go.dev/text/template) to render the configuration files. It uses
similar syntax to Jinja2 templates with different function names.

Comprehensive examples are beyond what we support and people wishing to use this should consult the official
[Go template engine](https://pkg.go.dev/text/template) documentation for syntax instructions. We also log the generated
output at each filter stage as a base64 string when trace logging is enabled.

#### Functions

In addition to the standard builtin functions we support several other functions.

##### iterate

The `iterate` function generates a list of numbers from 0 to the input provided. Useful for ranging over a list of
numbers.

Example:

```yaml
numbers:
{{- range $i := iterate 5 }}
  -  {{ $i }}
{{- end }}
```

##### env

The `env` function returns the value of an environment variable or a blank string.

Example:

```yaml
default_redirection_url: 'https://{{ env "DOMAIN" }}'
```

##### split

The `split` function splits a string by the separator.

Example:

```yaml
access_control:
  rules:
    - domain: 'app.{{ env "DOMAIN" }}'
      policy: bypass
      methods:
      {{ range _, $method := split "GET,POST" "," }}
        - {{ $method }}
      {{ end }}
```

##### join

The `join` function is similar to [split](#split) but does the complete oppiste, joining an array of strings with a
separator.

Example:

```yaml
access_control:
  rules:
    - domain: ['app.{{ join (split (env "DOMAINS") ",") "', 'app." }}']
      policy: bypass
```

##### contains

The `contains` function is a test function which checks if one string contains another string.

Example:

```yaml
{{ if contains (env "DOMAIN") "https://" }}
default_redirection_url: '{{ env "DOMAIN" }}'
{{ else }}
default_redirection_url: 'https://{{ env "DOMAIN" }}'
{{ end }}
```

##### hasPrefix

The `hasPrefix` function is a test function which checks if one string is prefixed with another string.

Example:

```yaml
{{ if hasPrefix (env "DOMAIN") "https://" }}
default_redirection_url: '{{ env "DOMAIN" }}'
{{ else }}
default_redirection_url: 'https://{{ env "DOMAIN" }}'
{{ end }}
```

##### hasSuffix

The `hasSuffix` function is a test function which checks if one string is suffixed with another string.

Example:

```yaml
{{ if hasSuffix (env "DOMAIN") "/" }}
default_redirection_url: 'https://{{ env "DOMAIN" }}'
{{ else }}
default_redirection_url: 'https://{{ env "DOMAIN" }}/'
{{ end }}
```

##### lower

The `lower` function is a conversion function which converts a string to all lowercase.

Example:

```yaml
default_redirection_url: 'https://{{ env "DOMAIN" | lower }}'
```

##### upper

The `upper` function is a conversion function which converts a string to all uppercase.

Example:

```yaml
default_redirection_url: 'https://{{ env "DOMAIN" | upper }}'
```
