---
title: "SQLite3"
description: "SQLite3 Configuration"
summary: "The SQLite3 storage provider."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 107500
toc: true
aliases:
  - /docs/configuration/storage/sqlite.html
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

If you don't have a SQL server, you can use [SQLite](https://en.wikipedia.org/wiki/SQLite).
However please note that this setup will prevent you from running multiple
instances of Authelia since the database will be a local file.

Use of this storage provider leaves Authelia [stateful](../../overview/authorization/statelessness.md). It's important
in highly available scenarios to use one of the other providers, and we highly recommend it in production environments,
but this requires you setup an external database such as [PostgreSQL](postgres.md).

## Configuration

{{< config-alert-example >}}

```yaml {title="configuration.yml"}
storage:
  encryption_key: 'a_very_important_secret'
  local:
    path: '/config/db.sqlite3'
```

## Options

This section describes the individual configuration options.

### encryption_key

See the [encryption_key docs](introduction.md#encryption_key).

### path

{{< confkey type="string" required="yes" >}}

The path where the SQLite3 database file will be stored. It will be created if the file does not exist.
