---
title: "Server Endpoint Rate Limits"
description: "Configuring the Server Authz Endpoint Settings."
summary: "Authelia supports several authorization endpoints on the internal web server. This section describes how to configure and tune them."
date: 2025-03-01T03:28:19+00:00
draft: false
images: []
weight: 199210
toc: true
aliases:
  - /c/authz
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

__Authelia__  imposes default rate limits on specific endpoints which can prevent faulty clients or bad actors from
consuming too many resources or using brute-force to potentially compromise security. This should not be confused with
[Regulation](../security/regulation.md) which is used to silently ban users from using the username / password form.

## Configuration

{{< config-alert-example >}}

```yaml {title=configuration.yml}
server:
  endpoints:
    rate_limits:
      reset_password_start:
        enable: true
        buckets:
          - period: '10 minutes'
            requests: 5
          - period: '15 minutes'
            requests: 10
          - period: '30 minutes'
            requests: 15
      reset_password_finish:
        enable: true
        buckets:
          - period: '1 minute'
            requests: 10
          - period: '2 minutes'
            requests: 15
      second_factor_totp:
        enable: true
        buckets:
          - period: '1 minute'
            requests: 30
          - period: '2 minutes'
            requests: 40
          - period: '10 minutes'
            requests: 50
      second_factor_duo:
        enable: true
        buckets:
          - period: '1 minute'
            requests: 10
          - period: '2 minutes'
            requests: 15
      session_elevation_start:
        enable: true
        buckets:
          - period: '5 minutes'
            requests: 3
          - period: '10 minutes'
            requests: 5
          - period: '1 hour'
            requests: 15
      session_elevation_finish:
        enable: true
        buckets:
          - period: '10 minutes'
            requests: 3
          - period: '20 minutes'
            requests: 5
          - period: '1 hour'
            requests: 15
```

## Common Options

### enable

{{< confkey type="boolean" default="true" required="no" >}}

Enables the given rate limit configuration. These are enabled by default.

### buckets

{{< confkey type="list(object)" required="no" >}}

The list of individual buckets to consider for each request.

#### period

{{< confkey type="string,integer" syntax="duration" required="situational">}}

Configures the period of time the tokenized bucket applies to.

Required if the [buckets](#buckets) have a configuration and [enable](#enable) is true.

#### requests

{{< confkey type="integer" required="situational">}}

Configures the number of requests the tokenized bucket applies to.

Required if the [buckets](#buckets) have a configuration and [enable](#enable) is true.

## Options

### reset_password_start

Configures the rate limiter which applies to the endpoint that initializes the reset password flow.

See [Common Options](#common-options) for the individual options for this section.

### reset_password_finish

Configures the rate limiter which applies to endpoints which consume tokens for the reset password flow.

See [Common Options](#common-options) for the individual options for this section.

### second_factor_totp

Configures the rate limiter which applies to the [TOTP](../second-factor/time-based-one-time-password.md) endpoint code
submissions for the second factor flow.

See [Common Options](#common-options) for the individual options for this section.

### second_factor_duo

Configures the rate limiter which applies to the [Duo / Mobile Push](../second-factor/duo.md) endpoint which initializes
the application authorization flow for the second factor flow.

See [Common Options](#common-options) for the individual options for this section.

### session_elevation_start

Configures the rate limiter which applies to the [Elevated Session](../identity-validation/elevated-session.md) endpoint
which initializes the code generation and notification for the elevated session flow.

See [Common Options](#common-options) for the individual options for this section.

### session_elevation_finish

Configures the rate limiter which applies to the [Elevated Session](../identity-validation/elevated-session.md) endpoint
which consumes the code for the elevated session flow.

See [Common Options](#common-options) for the individual options for this section.
