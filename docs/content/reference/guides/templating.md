---
title: "Templating"
description: "A reference guide on the templates system"
summary: "This section contains reference documentation for Authelia's templating capabilities."
date: 2024-03-14T06:00:14+11:00
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
- [contains](https://helm.sh/docs/chart_template_guide/function_list/#contains)
- [hasPrefix](https://helm.sh/docs/chart_template_guide/function_list/#hasprefix-and-hassuffix)
- [hasSuffix](https://helm.sh/docs/chart_template_guide/function_list/#hasprefix-and-hassuffix)
- [lower](https://helm.sh/docs/chart_template_guide/function_list/#lower)
- [upper](https://helm.sh/docs/chart_template_guide/function_list/#upper)
- [title](https://helm.sh/docs/chart_template_guide/function_list/#title)
- [trim](https://helm.sh/docs/chart_template_guide/function_list/#trim)
- [trimAll](https://helm.sh/docs/chart_template_guide/function_list/#trimAll)
- [trimSuffix](https://helm.sh/docs/chart_template_guide/function_list/#trimSuffix)
- [trimPrefix](https://helm.sh/docs/chart_template_guide/function_list/#trimPrefix)
- [replace](https://helm.sh/docs/chart_template_guide/function_list/#replace)
- [quote](https://helm.sh/docs/chart_template_guide/function_list/#quote-and-squote)
- [sha1sum](https://helm.sh/docs/chart_template_guide/function_list/#sha1sum)
- [sha256sum](https://helm.sh/docs/chart_template_guide/function_list/#sha256sum)
- sha512sum
- [squote](https://helm.sh/docs/chart_template_guide/function_list/#quote-and-squote)
- [now](https://helm.sh/docs/chart_template_guide/function_list/#now)
- [ago](https://helm.sh/docs/chart_template_guide/function_list/#ago)
- [toDate](https://helm.sh/docs/chart_template_guide/function_list/#toDate)
- [mustToDate](https://helm.sh/docs/chart_template_guide/function_list/#mustToDate)
- [date](https://helm.sh/docs/chart_template_guide/function_list/#date)
- [dateInZone](https://helm.sh/docs/chart_template_guide/function_list/#dateinzone)
- [htmlDate](https://helm.sh/docs/chart_template_guide/function_list/#htmldate)
- [htmlDateInZone](https://helm.sh/docs/chart_template_guide/function_list/#htmldateinzone)
- [duration](https://helm.sh/docs/chart_template_guide/function_list/#duration)
- [unixEpoch](https://helm.sh/docs/chart_template_guide/function_list/#unixepoch)
- [keys](https://helm.sh/docs/chart_template_guide/function_list/#keys)
- [sortAlpha](https://helm.sh/docs/chart_template_guide/function_list/#keys)
- [b64enc](https://helm.sh/docs/chart_template_guide/function_list/#encoding-functions)
- [b64dec](https://helm.sh/docs/chart_template_guide/function_list/#encoding-functions)
- [b32enc](https://helm.sh/docs/chart_template_guide/function_list/#encoding-functions)
- [b32dec](https://helm.sh/docs/chart_template_guide/function_list/#encoding-functions)
- [list](https://helm.sh/docs/chart_template_guide/function_list/#lists-and-list-functions)
- [dict](https://helm.sh/docs/chart_template_guide/function_list/#dict)
- [get](https://helm.sh/docs/chart_template_guide/function_list/#get)
- [set](https://helm.sh/docs/chart_template_guide/function_list/#set)
- [isAbs](https://helm.sh/docs/chart_template_guide/function_list/#isabs)
- [base](https://helm.sh/docs/chart_template_guide/function_list/#base)
- [dir](https://helm.sh/docs/chart_template_guide/function_list/#dir)
- [ext](https://helm.sh/docs/chart_template_guide/function_list/#ext)
- [clean](https://helm.sh/docs/chart_template_guide/function_list/#clean)
- osBase
- osClean
- osDir
- osExt
- osIsAbs
- [deepEqual](https://helm.sh/docs/chart_template_guide/function_list/#deepequal)
- [typeOf](https://helm.sh/docs/chart_template_guide/function_list/#type-functions)
- [typeIs](https://helm.sh/docs/chart_template_guide/function_list/#type-functions)
- [typeIsLike](https://helm.sh/docs/chart_template_guide/function_list/#type-functions)
- [kindOf](https://helm.sh/docs/chart_template_guide/function_list/#kind-functions)
- [kindIs](https://helm.sh/docs/chart_template_guide/function_list/#kind-functions)
- [default](https://helm.sh/docs/chart_template_guide/function_list/#default)
- [empty](https://helm.sh/docs/chart_template_guide/function_list/#empty)
- [indent](https://helm.sh/docs/chart_template_guide/function_list/#indent)
- [nindent](https://helm.sh/docs/chart_template_guide/function_list/#nindent)
- [uuidv4](https://helm.sh/docs/chart_template_guide/function_list/#uuid-functions)
- [urlquery](https://helm.sh/docs/chart_template_guide/function_list/#urlquery)
- urlunquery (opposite of urlquery)
- [fromYaml](https://helm.sh/docs/chart_template_guide/function_list/#fromyaml)
- [toYaml](https://helm.sh/docs/chart_template_guide/function_list/#type-conversion-functions)
- [toYamlPretty](https://helm.sh/docs/chart_template_guide/function_list/#type-conversion-functions)

See the [Helm Documentation](https://helm.sh/docs/chart_template_guide/function_list/) for more information. Please
note that only the functions listed above are supported and while the intention is to make the functions behave exactly
the same they may not necessarily behave exactly the same.

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

#### urlqueryarg

This function takes two parameters; a URL and key. The function parses the URL and returns the given value of the key in
the query string provided it exists.

#### glob

The `glob` function takes a single glob pattern argument and returns a list of files that match that pattern.

Examples:

```yaml {title="Print Names of Files in the '/opt/data' directory which have the .yml extension"}
{{ range (glob "/opt/data/*.yml") }}
{{ . }}
{{ end }}
```

```yaml {title="Print Content of Files in the '/opt/data' directory"}
{{ range (walk "/opt/data/*" "" false) }}
{{ fileContent . }}
{{ end }}
```

#### walk

The `walk` function takes three arguments, `path` (string), `pattern` (string), and `skipDir` (boolean). It's a function
which walks an entire tree of a given path and produces a list of structs with the following spec for every file in
that path:

```go
type WalkInfo struct {
  // Path is the relative path of the file or directory.
  Path string

  // AbsolutePath is the absolute file system path of the file or directory.
  AbsolutePath string

  // Name is the file or directory name.
  Name string

  // Size is the size in bytes of the file or directory.
  Size int64

  // Mode is the os.FileMode of the file or directory.
  Mode os.FileMode

  // Modified is time.Time that the file or directory was last modified.
  Modified time.Time

  // IsDir is truthy if this is a directory otherwise it is not.
  IsDir bool
}
```

The `path` must be defined and must be a valid file path.

If `pattern` is defined it must be a valid go regex pattern and the full file path and name is evaluated against the
pattern.

If `skipDir` is true any directories will be removed from the results.

Examples:

```yaml {title="Print Names of Files in the '/opt/data' directory"}
{{ range (walk "/opt/data" "" false) }}
{{ .AbsolutePath }}
{{ end }}
```

```yaml {title="Print Content of Files in the '/opt/data' directory"}
{{ range (walk "/opt/data" "" false) }}
  {{ if not .IsDir }}
{{ fileContent .AbsolutePath }}
  {{- end }}
{{ end }}
```

```yaml {title="Print Names of Files in the '/opt/data' directory which have the .yml extension"}
{{ range (walk "/opt/data" "^.*\.yml" false) }}
{{ .AbsolutePath }}
{{ end }}
```

```yaml {title="Print Content of Files in the '/opt/data' directory which have the .yml extension"}
{{ range (walk "/opt/data" "^.*\.yml" false) }}
  {{ if not .IsDir }}
{{ fileContent .AbsolutePath }}
  {{- end }}
{{ end }}
```

#### toYamlCustom

Converts an object into a YAML string with custom space indentation. Takes two inputs the first being the same as
`toYaml` the second being the number of spaces to indent the YAML with.
