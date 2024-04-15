---
title: "Templating"
description: "A reference guide on the templates system"
summary: "This section contains reference documentation for Authelia's templating capabilities."
date: 2022-12-23T21:58:54+11:00
draft: false
images: []
weight: 220
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

Authelia has several methods where users can interact with templates.

## Enable Templating

By default the [Notification Templates](./notification-templates.md) have templating enabled. To enable templating in configuration files, set the environment variable `X_AUTHELIA_CONFIG_FILTERS` to `template`. For more information see
[Configuration > Methods > Files: File Filters](../../configuration/methods/files.md#file-filters).

## Validation / Debugging

### Notifications

No specific method exists at this time to validate these templates, however a bad template may cause an error before
startup.

### Configuration

Two methods exist to validate the config template output:

1. The [authelia config template](../cli/authelia/authelia_config_template.md) command.
2. The [log level](../../configuration/miscellaneous/logging.md#level) value of `trace` will output the fully rendered
   configuration as a base64 string.

## Functions

Functions can be used to perform specific actions when executing templates. The following is a simple guide on which
functions exist.

### Standard Functions

Go has a set of standard functions which can be used. See the [Go Documentation](https://pkg.go.dev/text/template#hdr-Functions)
for more information.

### Helm-like Functions

The following functions which mimic the behavior of helm exist in most templating areas:

- env
- expandenv
- split
- splitList
- join
- contains
- hasPrefix
- hasSuffix
- lower
- upper
- title
- trim
- trimAll
- trimSuffix
- trimPrefix
- replace
- quote
- sha1sum
- sha256sum
- sha512sum
- squote
- now
- keys
- sortAlpha
- b64enc
- b64dec
- b32enc
- b32dec
- list
- dict
- get
- set
- isAbs
- base
- dir
- ext
- clean
- osBase
- osClean
- osDir
- osExt
- osIsAbs
- deepEqual
- typeOf
- typeIs
- typeIsLike
- kindOf
- kindIs
- default
- empty
- indent
- nindent
- uuidv4
- urlquery
- urlunquery (opposite of urlquery)

See the [Helm Documentation](https://helm.sh/docs/chart_template_guide/function_list/) for more information. Please
note that only the functions listed above are supported and the functions don't necessarily behave exactly the same.

__*Special Note:* The `env` and `expandenv` function automatically excludes environment variables that start with
`AUTHELIA_` or `X_AUTHELIA_` and end with one of `KEY`, `SECRET`, `PASSWORD`, `TOKEN`, or `CERTIFICATE_CHAIN`.__

### Special Functions

The following is a list of special functions and their syntax.

#### iterate

This template function takes a single input and is a positive integer. Returns a slice of uints from 0 to the provided
input.

#### mustEnv

Same as [env](#env) except if the environment variable is not set it returns an error.

#### fileContent

This template function takes a single input and is a string which should be a path. Returns the content of a file.

Example:

```yaml {title="configuration.yml"}
example: |
  {{- fileContent "/absolute/path/to/file" | nindent 2 }}
```

#### secret

Overload for [fileContent](#filecontent) except that tailing newlines will be removed.

##### secret example

```yaml {title="configuration.yml"}
example: '{{ secret "/absolute/path/to/file" }}'
```

#### mindent

Similar function to `nindent` except it skips indenting if there are no newlines, and includes the YAML multiline
formatting string provided. Input is in the format of `(int, string, string)`.

##### mindent example

Input:

```yaml {title="configuration.yml"}
example: {{ secret "/absolute/path/to/file" | mindent 2 "|" | msquote }}
```

Output (with multiple lines):

```yaml {title="configuration.yml"}
example: |
  <content of "/absolute/path/to/file">
```

Output (without multiple lines):

```yaml {title="configuration.yml"}
example: '<content of "/absolute/path/to/file">'
```

#### mquote

Similar to the `quote` function except it skips quoting for strings with multiple lines.

See the [mindent example](#mindent-example) for an example usage (just replace `msquote` with `mquote`, and the expected
quote char is `"` instead of `'`).

#### msquote

Similar to the `squote` function except it skips quoting for strings with multiple lines.

See the [mindent example](#mindent-example) for an example usage.
