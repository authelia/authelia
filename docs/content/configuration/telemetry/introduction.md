---
title: "Telemetry"
description: "Configuring the Telemetry settings"
summary: "Configuring the Telemetry settings."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 109100
toc: true
---

*Authelia* allows collecting telemetry for the purpose of monitoring it. At the present time we only allow collecting
[metrics](metrics.md). These [metrics](metrics.md) are stored in memory and must be scraped manually by the
administrator.

No metrics or telemetry are reported from an *Authelia* binary to any location the administrator doesn't explicitly
configure. This means by default all metrics are disabled.
