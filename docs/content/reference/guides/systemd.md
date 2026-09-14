---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

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
  so units ordered `After=authelia.service` do not start early. Note that `After=` only affects ordering when both units
  are part of the same transaction and does not itself start `authelia.service`, so dependent units should also declare
  `Wants=authelia.service` or `Requires=authelia.service`.
- The unit is placed into the deactivating state as soon as a shutdown begins.
- `systemctl reload authelia.service` sends `SIGHUP` which reopens the log file when the
  [log file path](../../configuration/miscellaneous/logging.md) is configured. This is always safe to run regardless of
  the logging configuration.
- The watchdog configured via `WatchdogSec=` is kept alive at half the configured interval. Combined with
  `Restart=on-failure` this means a process which stops responding is restarted.

None of this requires any configuration in Authelia. The service manager sets the `NOTIFY_SOCKET` environment variable
to provide the socket used for readiness and status notifications, and separately sets the `WATCHDOG_USEC` environment
variable to the watchdog timeout when `WatchdogSec=` is configured. Setting `WatchdogSec=0` disables watchdog
supervision. Notifications are only sent when these variables are present, so running Authelia outside of a service
manager is unaffected.

If you do not want the process restarted automatically, or you do not want the watchdog, you can override the relevant
options with a drop-in file. For the `authelia.service` unit:

```shell
systemctl edit authelia.service
```

For the `authelia@.service` template unit, the following applies the drop-in to every instance:

```shell
systemctl edit authelia@.service
```

Alternatively `systemctl edit authelia@<instance>.service` applies the drop-in to a single instance only, where
`<instance>` is the instance name.

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
