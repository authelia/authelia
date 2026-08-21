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

# Service Manager Notifications

The units we distribute use `Type=notify`. Authelia notifies the service manager directly using the `sd_notify`
protocol which means:

- The unit is only considered started once every listener is accepting connections and every service has been started,
  so units ordered `After=authelia.service` do not start early.
- The unit is placed into the deactivating state as soon as a shutdown begins.
- `systemctl reload authelia.service` sends `SIGHUP` which reopens the log file when the
  [log file path](../../configuration/miscellaneous/logging.md) is configured. This is always safe to run regardless of
  the logging configuration.
- The watchdog configured via `WatchdogSec=` is kept alive at half the configured interval. Combined with
  `Restart=on-failure` this means a process which stops responding is restarted.

None of this requires any configuration in Authelia. The notifications are only sent when the service manager requests
them via the `NOTIFY_SOCKET` and `WATCHDOG_USEC` environment variables, so running Authelia outside of a service
manager is unaffected.

If you do not want the process restarted automatically, or you do not want the watchdog, you can override the relevant
options with a drop-in file:

```shell
systemctl edit authelia.service
```

```ini
[Service]
Restart=no
WatchdogSec=0
```

# Overriding tmpfiles.d

The default tmpfiles.d permissions may be overly restrictive for some users. To override them you can just add files to
the `/etc/tmpfiles.d` directory. You can see the default tmpfiles.d configurations here:

- [authelia.conf](https://raw.githubusercontent.com/authelia/authelia/refs/heads/master/authelia.tmpfiles.conf)
- [authelia.config.conf](https://raw.githubusercontent.com/authelia/authelia/refs/heads/master/authelia.tmpfiles.config.conf)
