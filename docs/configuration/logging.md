---
layout: default
title: Logging
parent: Configuration
nav_order: 5
---

# Logging

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
<div markdown="1">
type: string
{: .label .label-config .label-purple } 
default: info
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Defines the level of logs used by Authelia. This level can be set to `trace`, `debug`, `info`, `warn`, or `error`. When
setting level to `trace`, you will generate a large amount of log entries and expose the `/debug/vars` and
`/debug/pprof/` endpoints which should not be enabled in production.

```yaml
log:
  level: debug
```

### format
<div markdown="1">
type: string
{: .label .label-config .label-purple } 
default: text
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Defines the format of the logs written by Authelia. This format can be set to `json` or `text`.

```yaml
log:
  format: json
```

#### JSON format
```
{"level":"info","msg":"Logging severity set to info","time":"2020-01-01T00:00:00+11:00"}
{"level":"info","msg":"Authelia is listening for non-TLS connections on 0.0.0.0:9091","time":"2020-01-01T00:00:00+11:00"}
```
#### Text format
```
time="2020-01-01T00:00:00+11:00" level=info msg="Logging severity set to info"
time="2020-01-01T00:00:00+11:00" level=info msg="Authelia is listening for non-TLS connections on 0.0.0.0:9091"
```

### file_path
<div markdown="1">
type: string (path)
{: .label .label-config .label-purple } 
default: ""
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Logs can be stored in a file when file path is provided. Otherwise logs are written to standard output. When setting the
level to `debug` or `trace` this will generate large amount of log entries. Administrators will need to ensure that
they rotate and/or truncate the logs over time to prevent significant long-term disk usage.

```yaml
log:
  file_path: /config/authelia.log
```

### keep_stdout
<div markdown="1">
type: boolean
{: .label .label-config .label-purple } 
default: false
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Overrides the behaviour to redirect logging only to the `file_path`. If set to `true` logs will be written to both
standard output, and the defined logging location.

```yaml
log:
  keep_stdout: true
```