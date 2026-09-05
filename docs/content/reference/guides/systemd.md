---
title: "Systemd"
description: "A reference guide on systemd"
summary: "This section contains reference documentation for Authelia's systemd units."
date: 2025-03-16T21:03:35+11:00
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

# Access to Database Sockets

If you're using database sockets and those sockets don't allow access by most users, after installation add the
`authelia` user to the associated group for that database engine.

Examples:

```shell
usermod -a -G redis authelia
usermod -a -G postgres authelia
```

# Filesystem layout

The units are sandboxed with `ProtectSystem=strict`, which mounts the entire filesystem read-only aside from the
directories listed below. The configuration directory is owned by `root` and readable by the `authelia` group so that a
compromised process cannot rewrite its own configuration or the secrets stored alongside it.

| Path                | Ownership           | Mode   | Purpose                                               |
| :------------------ | :------------------ | :----- | :---------------------------------------------------- |
| `/etc/authelia`     | `root:authelia`     | `0750` | Configuration and secrets. Read-only to the service.  |
| `/var/lib/authelia` | `authelia:authelia` | `0700` | Writable state: SQLite database, users database, etc. |
| `/var/log/authelia` | `authelia:authelia` | `0750` | Log files.                                            |
| `/run/authelia`     | `authelia:authelia` | `0750` | Runtime state such as a unix socket listener.         |

Anything that Authelia must _write_ belongs under `/var/lib/authelia`. This notably includes the
[file](../../configuration/first-factor/file.md) authentication backend's users database, which is rewritten whenever a
user changes their password, as well as the SQLite [storage](../../configuration/storage/sqlite.md) database and the
[file notifier](../../configuration/notifications/file.md).

If you keep writable state outside these directories you must grant the unit access to it with a drop-in:

```shell
systemctl edit authelia.service
```

```ini
[Service]
ReadWritePaths=/srv/authelia
```

# Listening on a privileged port

The units drop all capabilities. If you configure Authelia to listen on a port below 1024 you need to grant it back:

```ini
[Service]
AmbientCapabilities=CAP_NET_BIND_SERVICE
CapabilityBoundingSet=CAP_NET_BIND_SERVICE
```

# Reloading logs

The units implement `ExecReload=` as a `SIGHUP`, which causes Authelia to reopen its log file. Use
`systemctl reload authelia.service` from a logrotate `postrotate` script.

# Overriding tmpfiles.d

The default tmpfiles.d permissions may be overly restrictive for some users. To override them add a file to the
`/etc/tmpfiles.d` directory. Files are processed in lexicographic order by filename regardless of the directory they
reside in, and a file in `/etc/tmpfiles.d` with the same name as one in `/usr/lib/tmpfiles.d` replaces it entirely.

The packages install:

- `/usr/lib/tmpfiles.d/10-authelia.conf` from [authelia.tmpfiles.conf](https://raw.githubusercontent.com/authelia/authelia/refs/heads/master/authelia.tmpfiles.conf)
- `/usr/lib/tmpfiles.d/20-authelia-config.conf` from [authelia.tmpfiles.config.conf](https://raw.githubusercontent.com/authelia/authelia/refs/heads/master/authelia.tmpfiles.config.conf)
