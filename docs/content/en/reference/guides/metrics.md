---
title: "Telemetry"
description: "A reference guide on the telemetry collection"
lead: "This section contains reference documentation for Authelia's telemetry systems."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  reference:
    parent: "guides"
weight: 220
toc: true
---

No telemetry data is collected by any *Authelia* binaries, tooling, etc by default and all telemetry data is intended
to be used by administrators of their individual *Authelia* installs.

## Metrics

### Prometheus

*Authelia* supports exporting [Prometheus] metrics. These metrics are served on a separate port at the `/metrics` path
when configured. If metrics are enabled the metrics listener listens on `0.0.0.0:9959` as per the officially
[registered port] unless configured otherwise.

#### Recorded Metrics

##### Vectored Histograms

|          Name           | Vectors |                                                    Buckets                                                    |
|:-----------------------:|:-------:|:-------------------------------------------------------------------------------------------------------------:|
| authentication_duration | success | .0005, .00075, .001, .005, .01, .025, .05, .075, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.8, 0.9, 1, 5, 10, 15, 30, 60 |
|    request_duration     |  code   |                   .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 15, 20, 30, 40, 50, 60                    |

##### Vectored Counters

|             Name             |        Vectors        |
|:----------------------------:|:---------------------:|
|           request            |     code, method      |
|        verify_request        |         code          |
| authentication_first_factor  |    success, banned    |
| authentication_second_factor | success, banned, type |


#### Vector Definitions

##### code

The HTTP response status code.

##### method

The HTTP request method.

##### success

If the authentication was successful (`true`) or not (`false`).

##### banned

If the authentication was considered banned (`true`) or not (`false`).

##### type

The authentication type `webauthn`, `totp`, or `duo`.

[Prometheus]: https://prometheus.io/
[registered port]: https://github.com/prometheus/prometheus/wiki/Default-port-allocations
