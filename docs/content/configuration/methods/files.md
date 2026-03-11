---
title: "Files"
description: "Using the YAML File Configuration Method."
summary: "Authelia can be configured via files. This section describes utilizing this method."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 101200
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Loading behavior and Discovery

There are several options which affect the loading of files:

|                 Name                  |            Argument             |       Environment Variable       |                                    Description                                     |
|:-------------------------------------:|:-------------------------------:|:--------------------------------:|:----------------------------------------------------------------------------------:|
|          Configuration Paths          |        `--config`, `-c`         |       `X_AUTHELIA_CONFIG`        | A list of file or directory (non-recursive) paths to load configuration files from |
|         Configuration Reload          |               N/A               |    `X_AUTHELIA_CONFIG_RELOAD`    | Enables reloading Authelia on the modification of one of the defined config paths  |
| Configuration Reload Additional Paths |               N/A               | `X_AUTHELIA_CONFIG_RELOAD_PATHS` |    A list of additional paths to the watcher which are not configuration paths     |
|       [Filters](#file-filters)        | `--config.experimental.filters` |   `X_AUTHELIA_CONFIG_FILTERS`    |   A list of filters applied to every file from the Files or Directories options    |

### Configuration Paths

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
When specifying directories and files, the individual files specified **_must not_** be within any of the directories
specified.
{{< /callout >}}

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
If any directory is specified all files in that directory (non-recursive) should be considered part of the effective
Authelia configuration regardless if they handled by a specific configuration parser or not. Storing files not loaded
by Authelia in this directory **_is not supported_** and should it cause an error in the future this
**_is expected behaviour_**. This allows us to add additional file parsers in the future as well as configuration logic.
{{< /callout >}}

Configuration options can be discovered via either the Argument or Environment Variable, but not both at the same time.
If both are specified the Argument takes precedence and the Environment Variable is ignored. It is generally recommended
that if you're using a container that you use the Environment Variable as this will allow you to execute other commands
from the context of the container more easily.

### Configuration Reload

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
If the configuration reload fails to validate the configuration or a provider fails to initialize this will cause an
intentional crash of Authelia. This is expected behaviour and will not be changed. Users are expected to validate their
configuration before writing the changes to disk.

Changing a path that used to be a file into a directory is also not supported. Performing this action could result in
unexpected and undesirable behaviour.
{{< /callout >}}

In the instance of a file system notify event being observed that is a file path defined in the
[Configuration Paths](#configuration-paths), this will trigger a reload.

In the instance of a file system change being
observed that is a file within a directory path defined in the [Configuration Paths](#configuration-paths) this will
also cause a reload, regardless if the file is effectively a configuration file or not, and regardless of the type of
file system notify event that was observed.

In addition to the configuration paths you can define additional paths which will cause a reload with the
`X_AUTHELIA_CONFIG_RELOAD_PATHS` environment variable. These paths are not included in the configuration, and this makes
this option useful for folders that include secrets. This value is a comma separated list of paths which can either be
a file or a directory.

## Formats

The only supported configuration file format is [YAML](#yaml).

It's important that you sufficiently validate your configuration file. While we produce console errors for users in many
misconfiguration scenarios it's not perfect. Each file type has recommended methods for validation.

### YAML

*Authelia* loads `configuration.yml` as the configuration if you just run it. You can override this behavior with the
following syntax:

{{< envTabs "Validate Configuration" >}}
{{< envTab "Docker" >}}
```bash
docker run authelia/authelia:latest authelia --config config.custom.yml
```
{{< /envTab >}}
{{< envTab "Bare-Metal" >}}
```bash
authelia --config config.custom.yml
```
{{< /envTab >}}
{{< /envTabs >}}

#### YAML Validation

We recommend utilizing [VSCodium](https://vscodium.com/) or [VSCode](https://code.visualstudio.com/), both with the
[YAML Extension](https://open-vsx.org/extension/redhat/vscode-yaml) by RedHat to validate this file type.

This extension allows validation of the format and schema of a YAML file. To facilitate schema validation we publish
a set of JSON schemas which you can include as a special comment in order to validate the YAML file further. See the
[JSON Schema reference guide](../../reference/guides/schemas.md#json-schema) for more information including instructions
on how to utilize the schemas.

## Multiple Configuration Files

You can have multiple configuration files which will be merged in the order specified. If duplicate keys are specified
the last one to be specified is the one that takes precedence. Example:

{{< envTabs "Run With Multiple Configurations" >}}
{{< envTab "Docker" >}}
```bash
docker run -d authelia/authelia:latest authelia --config configuration.yml --config config-acl.yml --config config-other.yml
```

```bash
docker run -d authelia/authelia:latest authelia --config configuration.yml,config-acl.yml,config-other.yml
```
{{< /envTab >}}
{{< envTab "Bare-Metal" >}}
```bash
authelia --config configuration.yml --config config-acl.yml --config config-other.yml
```

```bash
authelia --config configuration.yml,config-acl.yml,config-other.yml
```
{{< /envTab >}}
{{< /envTabs >}}

Authelia's configuration files use the YAML format. A template with all possible options can be found at the root of the
repository {{< github-link name="here" path="config.template.yml" >}}.

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
You should not have configuration sections such as Access Control Rules or OpenID Connect 1.0
clients configured in multiple files. If you wish to split these into their own files that is fine, but if you have two
files that specify these sections and expect them to merge properly you are asking for trouble.
{{< /callout >}}

### Container

By default, the container looks for a configuration file at `/config/configuration.yml`.

### Docker

This is an example of how to override the configuration files loaded in docker:

```bash
docker run -d --volume /path/to/config:/config authelia:authelia:latest authelia --config=/config/configuration.yml --config=/config/configuration.acl.yml
```

See the [Docker Documentation](https://docs.docker.com/engine/reference/commandline/run/) for more information on the
`docker run` command.

### Docker Compose

An excerpt from a docker compose that allows you to specify multiple configuration files is as follows:

```yaml {title="compose.yml"}
services:
  authelia:
    container_name: '{{< sitevar name="host" nojs="authelia" >}}'
    image: 'authelia/authelia:latest'
    command:
      - 'authelia'
      - '--config=/config/configuration.yml'
      - '--config=/config/configuration.acl.yml'

```

See the [compose file reference](https://docs.docker.com/compose/compose-file/compose-file-v3/#command) for more
information.

### Kubernetes

An excerpt from a Kubernetes container that allows you to specify multiple configuration files is as follows:

```yaml {title="deployment.yml"}
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
          image: docker.io/authelia/authelia:latest
          command:
            - authelia
          args:
            - '--config=/configuration.yml'
            - '--config=/configuration.acl.yml'
```

See the Kubernetes [workloads documentation](https://kubernetes.io/docs/concepts/workloads/pods/#pod-templates) or the
[Container API docs](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#container-v1-core) for more
information.

## File Filters

File filters exist which allow modification of all configuration files after reading them from the
filesystem but before parsing their content. Unless explicitly specified these filters are _**NOT**_ covered by our
[Standard Versioning Policy](../../policies/versioning.md) and

There __*WILL*__ be a point where:

- The name of the CLI argument will change (we suggest using the environment variable which will not)
- The `expand-env` filter will be removed as it's deprecated

The filters are configured as a list of filter names by the `--config.experimental.filters` CLI argument and
`X_AUTHELIA_CONFIG_FILTERS` environment variable. We recommend using the environment variable as it ensures
commands executed from the container use the same filters and it's likely to be a permanent value whereas the argument
will likely change. If both the CLI argument and environment variable are used the environment variable is completely
ignored.

Filters can either be used on their own, in combination, or not at all. The filters are processed in order as they are
defined. You can preview the output of the YAML files when processed via the filters using the
[authelia config template](../../reference/cli/authelia/authelia_config_template.md) command.

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
The filters are applied in order and thus if the output of one filter outputs a string that
contains syntax for a subsequent filter it will be filtered. It is therefore suggested the template filter is the only
filter and if it isn't that it's last.
{{< /callout >}}

Examples:

{{< envTabs "Filters By Argument" >}}
{{< envTab "Docker" >}}
```bash
docker run -d authelia/authelia:latest authelia --config /config/configuration.yml --config.experimental.filters template
```
{{< /envTab >}}
{{< envTab "Bare-Metal" >}}
```bash
authelia --config /config/configuration.yml --config.experimental.filters template
```
{{< /envTab >}}
{{< /envTabs >}}

{{< envTabs "Filters By Environment" >}}
{{< envTab "Docker" >}}
```bash
docker run -d -e X_AUTHELIA_CONFIG_FILTERS=template -e X_AUTHELIA_CONFIG=/config/configuration.yml authelia/authelia:latest authelia
```
{{< /envTab >}}
{{< envTab "Bare-Metal" >}}
```bash
X_AUTHELIA_CONFIG_FILTERS=template X_AUTHELIA_CONFIG=/config/configuration.yml authelia
```
{{< /envTab >}}
{{< /envTabs >}}

### Go Template Filter

The name used to enable this filter is `template`. This filter is considered stable.

This filter uses the [Go template engine](https://pkg.go.dev/text/template) to render the configuration files. It uses
similar syntax to Jinja2 templates with different function names.

Comprehensive examples are beyond what we support and people wishing to use this should consult the official
[Go template engine](https://pkg.go.dev/text/template) documentation for syntax instructions. We also log the generated
output at each filter stage as a base64 string when trace logging is enabled.

#### Functions

In addition to the standard builtin functions we support several other functions which should operate similar.

See the [Templating Reference Guide](../../reference/guides/templating.md) for more information.

### Expand Environment Variable Filter

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
The Expand Environment Variable filter (i.e. `expand-env`) is officially deprecated. It will be removed in v4.40.0 and
will result in a startup error. This removal is done based on the experimental introduction of this feature and our
[Versioning Policy](../../policies/versioning.md). The removal decision was made due to the fact the
[Go Template Filter](#go-template-filter) can effectively do everything this filter can do without the
[Known Limitations](#known-limitations) which should be read carefully before usage of this filter.
{{< /callout >}}

The name used to enable this filter is `expand-env`.

This filter is the most common filter type used by many other applications. It is similar to using `envsubst` where it
replaces a string like `$EXAMPLE` or `${EXAMPLE}` with the value of the `EXAMPLE` environment variable.

This filter utilizes [os.ExpandEnv](https://pkg.go.dev/os#ExpandEnv) but does not include any environment variables that
look like they're an Authelia secret. This filter is very limited in what we can achieve, and there are known
limitations with this filter which may not be possible for us to work around. We discourage it's usage as the `template`
is much more robust and we have a lot more freedom to make adjustments to this filter compared to the `expand-env`
filter.

#### Known Limitations

The following known limitations exist with the Expand Environment Variable Filter.

- Has no inbuilt way to handle escaping a `$` so treats all `$` values as an expansion value. This can be escaped using
  `$$` as an indication that it should be a `$` literal. However this functionality likely will not work under all
  circumstances and is not guaranteed.
