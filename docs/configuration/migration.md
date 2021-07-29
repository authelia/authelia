---
layout: default
title: Migration
parent: Configuration
nav_order: 6
---

This section documents changes in the configuration which may require manual migration by the administrator. Typically
this only occurs when a configuration key is renamed or moved to a more appropriate location.

## Format

The migrations are formatted in a table with the old key and the new key. Periods indicate a different section which can
be represented in YAML as a dictionary i.e. it's indented.

Example:

In our table `server.host` with a value of `0.0.0.0` is represented in YAML like this:

```yaml
server:
  host: 0.0.0.0
```


## Migrations

### 4.30.0

The following changes occurred in 4.30.0:

|Previous Key|New Key               |
|:----------:|:--------------------:|
|host        |server.host           |
|port        |server.port           |
|tls_key     |server.tls.key        |
|tls_cert    |server.tls.certificate|
|log_level   |log.level             |
|log_file    |log.file              |
|log_format  |log.format            |