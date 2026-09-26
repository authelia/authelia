---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Elevated Session"
description: "Configuring the Authelia elevated session identity validation including one-time code lifespan, elevation duration, characters, and second factor options."
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
    require_reauthentication: 'disabled'
    reauthentication_lifespan: '5 minutes'
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

Enabling this option also makes the second factor methods available for registration in the user settings even when
no [access control](../security/access-control.md) rule uses the `two_factor` policy. Without that, an instance which
elevates sessions with a second factor would give users no way to register the method it asks them for.

### skip_second_factor

{{< confkey type="boolean" default="false" required="no" >}}

Treats sessions which have performed second factor authentication as elevated, skipping the One-Time Code process
entirely. In addition, users who have only performed first factor authentication but have a second factor method
configured are offered the choice to either perform the One-Time Code process or perform second factor authentication
instead.

As with [require_second_factor](#require_second_factor), enabling this option makes the second factor methods
available for registration in the user settings even when no access control rule uses the `two_factor` policy.

This option can be combined with the [require_second_factor](#require_second_factor) option to make second factor
authentication both necessary and sufficient for elevation: users with a configured second factor method must perform
second factor authentication and are then never asked for a One-Time Code, while users without one perform the
One-Time Code process as normal.

When [require_reauthentication](#require_reauthentication) is not `disabled`, users must satisfy it before any of the
processes in the following table.

The following table summarizes which process users must complete to perform a protected action depending on these two
options:

|          Configuration          |         User With a Second Factor Method         | User Without a Second Factor Method |
| :-----------------------------: | :----------------------------------------------: | :---------------------------------: |
|      both options disabled      |                  One-Time Code                   |            One-Time Code            |
|  `skip_second_factor` enabled   | One-Time Code _or_ Second Factor Authentication  |            One-Time Code            |
| `require_second_factor` enabled | Second Factor Authentication _and_ One-Time Code |            One-Time Code            |
|      both options enabled       |           Second Factor Authentication           |            One-Time Code            |

### require_reauthentication

{{< confkey type="string" default="disabled" required="no" >}}

Requires that the user has recently authenticated before an elevated session can be used. This is a prerequisite that
is checked before any of the other elevated session requirements; it does not replace the One-Time Code process.
Authentications performed when the user originally logged in count towards this requirement if they are within the
[reauthentication_lifespan](#reauthentication_lifespan).

|      Value      |                                                           Description                                                            |
| :-------------: | :------------------------------------------------------------------------------------------------------------------------------: |
|   `disabled`    |                                                 No re-authentication is required                                                 |
|   `password`    | The user must have recently completed first factor authentication (a password or passkey login, or a password re-authentication) |
| `second_factor` | The user must have recently performed second factor authentication; users without a second factor method must use their password |
|      `any`      |             The user may choose to authenticate with their password or any of their configured second factor methods             |

If Authelia is unable to determine which second factor methods a user has configured, the requirement is enforced and
the user is only offered the methods allowed by the configured value: `password` for `password` and `any`, and second
factor authentication for `second_factor`.

### reauthentication_lifespan

{{< confkey type="string,integer" syntax="duration" default="5 minutes" required="no" >}}

The duration after an authentication during which it satisfies the
[require_reauthentication](#require_reauthentication) option. This lifespan should comfortably exceed the time users need
to complete the One-Time Code step and the protected action.
