---
layout: default
title: SQLite
parent: Storage Backends
grand_parent: Configuration
nav_order: 4
---

# SQLite

If you don't have a SQL server, you can use [SQLite](https://en.wikipedia.org/wiki/SQLite).
However please note that this setup will prevent you from running multiple
instances of Authelia since the database will be a local file.

Use of this storage provider leaves Authelia [stateful](../../features/statelessness.md). It's important in highly
available scenarios to use one of the other providers, and we highly recommend it in production environments, but this
requires you setup an external database.

## Configuration

```yaml
storage:
  encryption_key: a_very_important_secret
  local:
    path: /config/db.sqlite3
```

## Options

### path
<div markdown="1">
type: string
{: .label .label-config .label-blue }
required: yes
{: .label .label-config .label-red }
</div>

The path where the SQLite3 database file will be stored. It will be created if the file does not exist.
