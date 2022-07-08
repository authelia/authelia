---
title: "Log"
description: "Configuring the Log Settings."
lead: "Authelia performs logging to various locations. This section describes how to configure and tune this."
date: 2021-06-01T14:09:50+10:00
draft: false
images: []
menu:
  configuration:
    parent: "miscellaneous"
weight: 199400
toc: true
aliases:
  - /docs/configuration/logging.html
---

The logging section tunes the logging settings.

## Configuration

```yaml
log:
  level: info
  format: text
  file_path: ""
  keep_stdout: false
```

## Options

### level

{{< confkey type="string" default="info" required="no" >}}

Defines the level of logs used by Authelia. This level can be set to `trace`, `debug`, `info`, `warn`, or `error`. When
setting level to `trace`, you will generate a large amount of log entries and expose the `/debug/vars` and
`/debug/pprof/` endpoints which should not be enabled in production.

```yaml
log:
  level: debug
```

### format

{{< confkey type="string" default="text" required="no" >}}

Defines the format of the logs written by Authelia. This format can be set to `json` or `text`.

```yaml
log:
  format: json
```

#### JSON format

```json
{"level":"info","msg":"Logging severity set to info","time":"2020-01-01T00:00:00+11:00"}
{"level":"info","msg":"Authelia is listening for non-TLS connections on 0.0.0.0:9091","time":"2020-01-01T00:00:00+11:00"}
```

#### Text format

```text
time="2020-01-01T00:00:00+11:00" level=info msg="Logging severity set to info"
time="2020-01-01T00:00:00+11:00" level=info msg="Authelia is listening for non-TLS connections on 0.0.0.0:9091"
```

### file_path

{{< confkey type="string" required="no" >}}

Logs can be stored in a file when file path is provided. Otherwise logs are written to standard output. When setting the
level to `debug` or `trace` this will generate large amount of log entries. Administrators will need to ensure that
they rotate and/or truncate the logs over time to prevent significant long-term disk usage.

If you include the value `%d` in the filename it will replace this value with a date time indicative of the time
the logger was initialized using `2006-02-01T150405Z` as the format.

#### File Path Examples

__Standard Example:__

```yaml
log:
  file_path: /config/authelia.log
```

__Date Time Example:__

```yaml
log:
  file_path: /config/authelia.%d.log
```

### keep_stdout

{{< confkey type="boolean" default="false" required="no" >}}

Overrides the behaviour to redirect logging only to the `file_path`. If set to `true` logs will be written to both
standard output, and the defined logging location.

```yaml
log:
  keep_stdout: true
```
