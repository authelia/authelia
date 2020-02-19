---
layout: default
title: SQLite
parent: Storage backends
grand_parent: Configuration
nav_order: 3
---

# SQLite

If you don't have a SQL server, you can use [SQLite](https://en.wikipedia.org/wiki/SQLite).
However please note that this setup will prevent you from running multiple
instances of Authelia since the database will be a local file.

## Configuration

Just give the path to the sqlite database. It will be created if the file does not exist.

```yaml
storage:
    local:
        path: /var/lib/authelia/db.sqlite3
```
