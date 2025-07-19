---
title: "Duo / Mobile Push"
description: "Configuring the Duo Mobile Push Notification Second Factor Method."
summary: ""
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 103200
toc: true
aliases:
  - /docs/configuration/duo-push-notifications.html
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

Authelia supports mobile push notifications relying on [Duo].

Follow the instructions in the dedicated [documentation](../../overview/authentication/push-notification/index.md) for
instructions on how to set up push notifications in Authelia.

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
The configuration options in the following sections are noted as required. They are however only required when
you have this section defined. i.e. if you don't wish to use the [Duo](https://duo.com/) push notifications, you can just not define this
section of the configuration.
{{< /callout >}}

## Configuration

{{< config-alert-example >}}

```yaml {title="configuration.yml"}
duo_api:
  disable: false
  hostname: 'api-123456789.{{< sitevar name="domain" nojs="example.com" >}}'
  integration_key: 'ABCDEF'
  secret_key: '1234567890abcdefghifjkl'
  enable_self_enrollment: false
```

## Options

This section describes the individual configuration options.

### Disable

{{< confkey type="boolean" default="false" required="no" >}}

Disables Duo. If the hostname, integration_key, and secret_key are all empty strings or undefined this is automatically
true.

### hostname

{{< confkey type="string" required="yes" >}}

The [Duo] API hostname. This is provided in the [Duo] dashboard.

### integration_key

{{< confkey type="string" required="yes" >}}

The non-secret [Duo] integration key. Similar to a client identifier. This is provided in the [Duo] dashboard.

### secret_key

{{< confkey type="string" required="yes" secret="yes" >}}

The secret [Duo] key used to verify your application is valid. This is provided in the [Duo] dashboard.

### enable_self_enrollment

{{< confkey type="boolean" default="false" required="no" >}}

Enables [Duo] device self-enrollment from within the Authelia portal.

[Duo]: https://duo.com/
