---
title: "Elevated Session"
description: "Elevated Session Identity Validation Configuration"
summary: "Authelia uses multiple methods to verify the identity of users to prevent a malicious user from performing actions on behalf of them. This section describes the Elevated Session method."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 105200
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

The Elevated Session Identity Validation implementation ensures that users cannot perform actions which may adjust the
security characteristics of their account without first ensuring the user is adequately identified. The settings below
therefore can affect the level of security Authelia provides to your users so they should be carefully considered.

Elevated Sessions are initiated by generating a One-Time Code, this One-Time Code is then exchanged for a special status
stored in the session which allows the privileged actions. The elevation itself is anchored to the users Remote IP and
only lasts for a finite amount of time. Users at this time may not revoke the elevated session manually, but may revoke
the One-Time Code so that it cannot be used to create a new elevated session.

## Configuration

{{< config-alert-example >}}

```yaml {title="configuration.yml"}
identity_validation:
  elevated_session:
    code_lifespan: '5 minutes'
    elevation_lifespan: '10 minutes'
    characters: 8
    require_second_factor: false
    skip_second_factor: false
```

## Options

This section describes the individual configuration options.

### code_lifespan

{{< confkey type="string,integer" syntax="duration" default="5 minutes" required="no" >}}

The lifespan of the randomly generated One-Time Code after which it's considered invalid

### elevation_lifespan

{{< confkey type="string,integer" syntax="duration" default="10 minutes" required="no" >}}

The lifespan of the elevation after initially validating the One-Time Code before it expires.

### characters

{{< confkey type="integer" default="8" required="no" >}}

The number of characters the random One-Time Code has. Maximum value is currently 20, but we recommend keeping it
between 8 and 12. It's strongly discouraged to reduce it below 8.

### require_second_factor

{{< confkey type="boolean" default="false" required="no" >}}

Makes second factor authentication a prerequisite for the elevated session process. Users who have only performed
first factor authentication must perform second factor authentication before they can establish an elevated session.
The One-Time Code process is still required in addition to second factor authentication unless
[skip_second_factor](#skip_second_factor) is also enabled.

This option only affects users who have at least one second factor method configured; users without any configured
second factor method perform the One-Time Code process as normal.

### skip_second_factor

{{< confkey type="boolean" default="false" required="no" >}}

Treats sessions which have performed second factor authentication as elevated, skipping the One-Time Code process
entirely. In addition, users who have only performed first factor authentication but have a second factor method
configured are offered the choice to either perform the One-Time Code process or perform second factor authentication
instead.

Can be combined with the [require_second_factor](#require_second_factor) option to make second factor authentication
both necessary and sufficient for elevation: users with a configured second factor method must perform second factor
authentication and are then never asked for a One-Time Code, while users without one perform the One-Time Code process
as normal.

The following table summarizes which process users must complete to perform a protected action depending on these two
options:

|          Configuration          |          User With a Second Factor Method           | User Without a Second Factor Method |
|:-------------------------------:|:---------------------------------------------------:|:-----------------------------------:|
|      both options disabled      |                    One-Time Code                    |            One-Time Code            |
|  `skip_second_factor` enabled   |   One-Time Code *or* second factor authentication   |            One-Time Code            |
| `require_second_factor` enabled |  second factor authentication *and* One-Time Code   |            One-Time Code            |
|      both options enabled       |            second factor authentication             |            One-Time Code            |
