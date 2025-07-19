---
title: "First Factor"
name: "test"
description: "Configuring Authelia First Factor Authentication."
summary: "Authelia uses a username and password for a first factor method. This section describes configuring this."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 102100
toc: true
aliases:
  - /c/1fa
  - /docs/configuration/authentication/
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

There are two ways to integrate *Authelia* with an authentication backend:

* [LDAP](ldap.md): users are stored in remote servers like [OpenLDAP], [OpenDJ], [FreeIPA], or
  [Microsoft Active Directory].
* [File](file.md): users are stored in [YAML] file with a hashed version of their password.

## Configuration

{{< config-alert-example >}}

```yaml {title="configuration.yml"}
authentication_backend:
  refresh_interval: '5m'
  password_reset:
    disable: false
    custom_url: ''
  password_change:
    disable: false
```

## Options

This section describes the individual configuration options.

### refresh_interval

{{< confkey type="string,integer" syntax="duration" default="5 minutes" required="no">}}

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
When using the [File Provider](#file) this value has a default value of `always` as the cost in this
scenario is basically not measurable, users can, however, override this setting by setting an explicit value.
{{< /callout >}}

This setting controls the interval at which details are refreshed from the backend. The details refreshed in order of
importance are the groups, email address, and display name. This is particularly useful for the [File Provider](#file)
when [watch](file.md#watch) is enabled or generally with the [LDAP Provider](#ldap).

In addition to the duration values this option accepts `always` and `disable` as values; where `always` will always
refresh this value, and `disable` will never refresh the profile.

### password_reset

#### disable

{{< confkey type="boolean" default="false" required="no" >}}

This setting controls if users can reset their password from the web frontend or not.

#### custom_url

{{< confkey type="string" required="no" >}}

The custom password reset URL. This replaces the inbuilt password reset functionality and disables the endpoints if
this is configured to anything other than nothing or an empty string.

### password_change

#### disable

{{< confkey type="boolean" default="false" required="no" >}}

This setting controls if users can change their password from the web frontend or not.


### file

The [file](file.md) authentication provider.

### ldap

The [LDAP](ldap.md) authentication provider.

[OpenLDAP]: https://www.openldap.org/
[OpenDJ]: https://www.openidentityplatform.org/opendj
[FreeIPA]: https://www.freeipa.org/
[Microsoft Active Directory]: https://docs.microsoft.com/en-us/windows-server/identity/ad-ds/ad-ds-getting-started
[YAML]: https://yaml.org/
