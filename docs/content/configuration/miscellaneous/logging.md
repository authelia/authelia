---
title: "Log"
description: "Configuring the Log Settings."
summary: "Authelia performs logging to various locations. This section describes how to configure and tune this."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 199400
toc: true
aliases:
  - /docs/configuration/logging.html
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

The logging section tunes the logging settings.

## Configuration

{{< config-alert-example >}}

```yaml {title="configuration.yml"}
log:
  level: 'info'
  format: 'text'
  file_path: ''
  keep_stdout: false
```

## Options

This section describes the individual configuration options.

### level

{{< confkey type="string" default="info" required="no" >}}

Defines the level of logs used by Authelia. This level can be set to `trace`, `debug`, `info`, `warn`, or `error`. When
setting level to `trace`, you will generate a large amount of log entries and expose the `/debug/vars` and
`/debug/pprof/` endpoints which should not be enabled in production.

```yaml {title="configuration.yml"}
log:
  level: 'debug'
```

### format

{{< confkey type="string" default="text" required="no" >}}

Defines the format of the logs written by Authelia. This format can be set to `json` or `text`.

```yaml {title="configuration.yml"}
log:
  format: 'json'
```

#### JSON format

```json
{"level":"info","msg":"Logging severity set to info","time":"2020-01-01T00:00:00+11:00"}
{"level":"info","msg":"Authelia is listening for non-TLS connections on 0.0.0.0:{{< sitevar name="port" nojs="9091" >}}","time":"2020-01-01T00:00:00+11:00"}
```

#### Text format

```text
time="2020-01-01T00:00:00+11:00" level=info msg="Logging severity set to info"
time="2020-01-01T00:00:00+11:00" level=info msg="Authelia is listening for non-TLS connections on 0.0.0.0:{{< sitevar name="port" nojs="9091" >}}"
```

### file_path

{{< confkey type="string" required="no" >}}

Logs can be stored in a file when file path is provided. Otherwise logs are written to standard output. When setting the
level to `debug` or `trace` this will generate large amount of log entries. Administrators will need to ensure that
they rotate and/or truncate the logs over time to prevent significant long-term disk usage.

There are two replacements that exist in this string for the purpose of including the date. The `%d` value which just
uses the [RFC3339] layout, and the `{datetime}` replacement which by
default uses the [RFC3339] layout, but optionally can be suffixed with the
[Go Layout](https://pkg.go.dev/time#pkg-constants) semantics in the format of `{datetime:<layout>}` where `<layout>` is
the layout supported by Go.

When using a log file sending the Authelia process a SIGHUP will cause it to close and reopen the current log file and
truncate it. This is useful for log rotation services which can force a reopen so the file descriptor open does not
continue to append to the old log file.

#### File Path Examples

__Standard Example:__

```yaml {title="configuration.yml"}
log:
  file_path: '/config/authelia.log'
```

__Date Time Example:__

```yaml {title="configuration.yml"}
log:
  file_path: '/config/authelia.%d.log'
```

__Date Time Example (with custom layout):__

```yaml {title="configuration.yml"}
log:
  file_path: '/config/authelia.{datetime:Mon Jan 2 15:04:05 MST 2006}.log'
```

### keep_stdout

{{< confkey type="boolean" default="false" required="no" >}}

Overrides the behavior to redirect logging only to the `file_path`. If set to `true` logs will be written to both
standard output, and the defined logging location.

```yaml {title="configuration.yml"}
log:
  keep_stdout: true
```

[RFC3339]: https://datatracker.ietf.org/doc/html/rfc3339
