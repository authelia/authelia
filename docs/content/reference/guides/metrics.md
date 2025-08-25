---
title: "Telemetry"
description: "A reference guide on the telemetry collection"
summary: "This section contains reference documentation for Authelia's telemetry systems."
date: 2024-03-14T06:00:14+11:00
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

No telemetry data is collected by any *Authelia* binaries, tooling, etc by default and all telemetry data is intended
to be used by administrators of their individual *Authelia* installs.

## Metrics

### Prometheus

*Authelia* supports exporting [Prometheus] metrics. These metrics are served on a separate port at the `/metrics` path
when configured. If metrics are enabled the metrics listener listens on `:9959` as per the officially
[registered port] unless configured otherwise.

#### Example Prometheus Job

```yaml
# Authelia
  - job_name: 'authelia'
    scrape_interval: '15s'
    scheme: 'http'
    static_configs:
    - targets: ['{{< sitevar name="host" nojs="authelia" >}}:9959']
```

*Notes: Replace '{{< sitevar name="host" nojs="authelia" >}}' with the URL or IP of your Authelia container.*


#### Recorded Metrics

##### Vectored Counters

|        Name         |           Vectors           |       Description        |
|:-------------------:|:---------------------------:|:------------------------:|
|       request       |      `code`, `method`       |       All Requests       |
|        authz        |           `code`            |      Authz Requests      |
|        authn        |     `success`, `banned`     |   Authn Requests (1FA)   |
|    authn_passkey    |          `success`          | Authn Requests (Passkey) |
| authn_second_factor | `success`, `banned`, `type` |   Authn Requests (2FA)   |

##### Vectored Histograms

|              Name               |      Vectors       |                                                    Buckets                                                    |
|:-------------------------------:|:------------------:|:-------------------------------------------------------------------------------------------------------------:|
|         authn_duration          |     `success`      | .0005, .00075, .001, .005, .01, .025, .05, .075, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.8, 0.9, 1, 5, 10, 15, 30, 60 |
|        request_duration         |       `code`       |                   .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 15, 20, 30, 40, 50, 60                    |
| request_duration_openid_connect | `endpoint`, `code` |                   .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 15, 20, 30, 40, 50, 60                    |

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

##### endpoint

The endpoint name.

OpenID Connect 1.0 Endpoint Names:

- consent
- authorization
- pushed_authorization_request
- token
- userinfo
- revocation
- introspection
- openid_configuration
- oauth_configuration
- jwks

### Grafana

Metrics collected by [Prometheus] can be displayed and analyzed in [Grafana] by creating a new dashboard or by
importing an existing one.

#### Community Dashboard

*Authelia* provides a community-maintained [Grafana] dashboard, which is intended to serve as an example to explore
the available metrics.

##### Installation

To import the dashboard into [Grafana], either download the JSON file
[here](https://github.com/authelia/authelia/blob/master/examples/grafana-dashboards/simple.json) or copy its contents.

[Prometheus]: https://prometheus.io/
[Grafana]: https://grafana.com/
[registered port]: https://github.com/prometheus/prometheus/wiki/Default-port-allocations

