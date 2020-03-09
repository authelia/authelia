---
layout: default
title: Storage backends
parent: Configuration
nav_order: 10
has_children: true
---

# Storage backends

**Authelia** supports multiple storage backends. The backend is used
to store user preferences, 2FA device handles and secrets, authentication
logs, etc...

The available options are:

* [MariaDB](./mariadb.md)
* [MySQL](./mysql.md)
* [Postgres](./postgres.md)
* [SQLite](./sqlite.md)