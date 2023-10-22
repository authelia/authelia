---
title: "Elevated Session"
description: "Elevated Session Identity Validation Configuration"
lead: "Authelia uses multiple methods to verify the identity of users to prevent a malicious user from performing actions on behalf of them. This section describes the Elevated Session method."
date: 2023-10-21T17:51:47+10:00
draft: false
images: []
menu:
  configuration:
    parent: "identity-validation"
weight: 105200
toc: true
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

```yaml
identity_validation:
  elevated_session:
    expiration: '5 minutes'
    elevation_expiration: '10 minutes'
    characters: 8
    require_second_factor: false
    skip_second_factor: false
```

## Options

This section describes the individual configuration options.

### expiration

{{< confkey type="string,integer" syntax="duration" default="5 minutes" required="no" >}}

The duration of time the emailed One-Time Code is considered valid.

### elevation_expiration

{{< confkey type="string,integer" syntax="duration" default="10 minutes" required="no" >}}

The duration of time that the session elevation lasts after validating the One-Time Code.

### characters

{{< confkey type="integer" default="8" required="no" >}}

The number of characters the random One-Time Code has.

### require_second_factor

{{< confkey type="boolean" default="false" required="no" >}}

Requires second factor authentication for all protected actions in addition to the elevated session provided the user
has configured a second factor authentication method.

### skip_second_factor

{{< confkey type="boolean" default="false" required="no" >}}

Skips the elevated session requirement if the user has performed second factor authentication. Can be combined with the
[require_second_factor](#requiresecondfactor) option to always (and only) require second factor authentication.
