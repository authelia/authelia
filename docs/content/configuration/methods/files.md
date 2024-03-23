---
title: "Files"
description: "Using the YAML File Configuration Method."
summary: "Authelia can be configured via files. This section describes utilizing this method."
date: 2022-06-15T17:51:47+10:00
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

|       Name        |            Argument             |    Environment Variable     |                                    Description                                     |
|:-----------------:|:-------------------------------:|:---------------------------:|:----------------------------------------------------------------------------------:|
| Files/Directories |        `--config`, `-c`         |     `X_AUTHELIA_CONFIG`     | A list of file or directory (non-recursive) paths to load configuration files from |
|      Filters      | `--config.experimental.filters` | `X_AUTHELIA_CONFIG_FILTERS` |   A list of filters applied to every file from the Files or Directories options    |

__*Note:* when specifying directories and files, the individual files specified must not be within any of the
directories specified.__

Configuration options can be discovered via either the Argument or Environment Variable, but not both at the same time.
If both are specified the Argument takes precedence and the Environment Variable is ignored. It is generally recommended
that if you're using a container that you use the Environment Variable as this will allow you to execute other commands
from the context of the container more easily.

## Formats

The only supported configuration file format is [YAML](#yaml).

It's important that you sufficiently validate your configuration file. While we produce console errors for users in many
misconfiguration scenarios it's not perfect. Each file type has recommended methods for validation.

### YAML

*Authelia* loads `configuration.yml` as the configuration if you just run it. You can override this behavior with the
following syntax:

```bash
authelia --config config.custom.yml
```

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

```bash
authelia --config configuration.yml --config config-acl.yml --config config-other.yml
authelia --config configuration.yml,config-acl.yml,config-other.yml
```

Authelia's configuration files use the YAML format. A template with all possible options can be found at the root of the
repository {{< github-link name="here" path="config.template.yml" >}}.

*__Important Note:__ You should not have configuration sections such as Access Control Rules or OpenID Connect 1.0
clients configured in multiple files. If you wish to split these into their own files that is fine, but if you have two
files that specify these sections and expect them to merge properly you are asking for trouble.*

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

```yaml {title="cocker-compose.yml"}
version: '3.8'
services:
  authelia:
    container_name: 'authelia'
    image: 'authelia/authelia:latest'
    command:
      - 'authelia'
      - '--config=/config/configuration.yaml'
      - '--config=/config/configuration.acl.yaml'

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
          image: docker.io/authelia/authelia:latest
          command:
            - authelia
          args:
            - '--config=/configuration.yaml'
            - '--config=/configuration.acl.yaml'
```

See the Kubernetes [workloads documentation](https://kubernetes.io/docs/concepts/workloads/pods/#pod-templates) or the
[Container API docs](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#container-v1-core) for more
information.

## File Filters

Experimental file filters exist which allow modification of all configuration files after reading them from the
filesystem but before parsing their content. These filters are _**NOT**_ covered by our
[Standard Versioning Policy](../../policies/versioning.md) at least at this time, however we will make every effort
to avoid breaking them unnecessarily and we include several of these filters within our CI testing regiment.

There __*WILL*__ be a point where:
- the name of the CLI argument will change (we suggest using the environment variable which will not)
- the `expand-env` filter might be removed

The filters are configured as a list of filter names by the `--config.experimental.filters` CLI argument and
`X_AUTHELIA_CONFIG_FILTERS` environment variable. We recommend using the environment variable as it ensures
commands executed from the container use the same filters and it's likely to be a permanent value whereas the argument
will likely change. If both the CLI argument and environment variable are used the environment variable is completely
ignored.

Filters can either be used on their own, in combination, or not at all. The filters are processed in order as they are
defined. You can preview the output of the YAML files when processed via the filters using the
[authelia config template](../../reference/cli/authelia/authelia_config_template.md) command.

_**Important Note:** the filters are applied in order and thus if the output of one filter outputs a string that
contains syntax for a subsequent filter it will be filtered. It is therefore suggested the template filter is the only
filter and if it isn't that it's last._

Examples:

```bash
authelia --config config.yml --config.experimental.filters expand-env,template
```

```text
X_AUTHELIA_CONFIG_FILTERS=expand-env,template
```

### Expand Environment Variable Filter

The name used to enable this filter is `expand-env`.

This filter is the most common filter type used by many other applications. It is similar to using `envsubst` where it
replaces a string like `$EXAMPLE` or `${EXAMPLE}` with the value of the `EXAMPLE` environment variable.

This filter utilizes [os.ExpandEnv](https://pkg.go.dev/os#ExpandEnv) but does not include any environment variables that
look like they're an Authelia secret. This filter is very limited in what we can achieve, and there are known
limitations with this filter which may not be possible for us to work around. We discourage it's usage as the `template`
is much more robust and we have a lot more freedom to make adjustments to this filter compared to the `expand-env`
filter.

Known Limitations:

- Has no way to handle escaping a `$` so treats all `$` values as an expansion value. This can be escaped using `$$` as
  an indication that it should be a `$` literal. However this functionality likely will not work under all
  circumstances.

### Go Template Filter

The name used to enable this filter is `template`.

This filter uses the [Go template engine](https://pkg.go.dev/text/template) to render the configuration files. It uses
similar syntax to Jinja2 templates with different function names.

Comprehensive examples are beyond what we support and people wishing to use this should consult the official
[Go template engine](https://pkg.go.dev/text/template) documentation for syntax instructions. We also log the generated
output at each filter stage as a base64 string when trace logging is enabled.

#### Functions

In addition to the standard builtin functions we support several other functions which should operate similar.

See the [Templating Reference Guide](../../reference/guides/templating.md) for more information.
