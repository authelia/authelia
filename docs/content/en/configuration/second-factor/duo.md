---
title: "Duo / Mobile Push"
description: "Configuring the Duo Mobile Push Notification Second Factor Method."
lead: ""
date: 2022-03-20T12:52:27+11:00
draft: false
images: []
menu:
  configuration:
    parent: "second-factor"
weight: 103200
toc: true
aliases:
  - /docs/configuration/duo-push-notifications.html
---

Authelia supports mobile push notifications relying on [Duo].

Follow the instructions in the dedicated [documentation](../../overview/authentication/push-notification/index.md) for
instructions on how to set up push notifications in Authelia.

__Note:__ The configuration options in the following sections are noted as required. They are however only required when
you have this section defined. i.e. if you don't wish to use the [Duo] push notifications you can just not define this
section of the configuration.

## Configuration

```yaml
duo_api:
  disable: false
  hostname: api-123456789.example.com
  integration_key: ABCDEF
  secret_key: 1234567890abcdefghifjkl
  enable_self_enrollment: false
```

## Options

### Disable

{{< confkey type="boolean" default="false" required="no" >}}

Disables Duo. If the hostname, integration_key, and secret_key are all empty strings or undefined this is automatically
true.

### hostname

{{< confkey type="string" required="yes" >}}

The [Duo] API hostname supplied by [Duo].

### integration_key

{{< confkey type="string" required="yes" >}}

The non-secret [Duo] integration key. Similar to a client identifier.

### secret_key

{{< confkey type="string" required="yes" >}}

The secret [Duo] key used to verify your application is valid. It's recommended to set this via a
[secret](../methods/secrets.md).

### enable_self_enrollment

{{< confkey type="boolean" default="false" required="no" >}}

Enables [Duo] device self-enrollment from within the Authelia portal.

[Duo]: https://duo.com/
