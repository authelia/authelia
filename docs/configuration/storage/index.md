---
layout: default
title: Storage backends
parent: Configuration
nav_order: 10
has_children: true
---

# Storage backends

**Authelia** supports multiple storage backends. This backend is used
to store user preferences, 2FA device handles and secrets, authentication
logs, etc...

The available options are:

* [SQLite](./sqlite.html)
* [MariaDB](./mariadb.html)
* ~~MySQL~~ ([#512](https://github.com/authelia/authelia/issues/512))
* [Postgres]((./postgres.html))