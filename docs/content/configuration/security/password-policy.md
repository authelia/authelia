---
title: "Password Policy"
description: "Password Policy Configuration"
summary: "Configuring the Password Policy."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 104400
toc: true
aliases:
  - /docs/configuration/password_policy.html
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

*Authelia* allows administrators to configure an enforced password policy.

## Configuration

{{< config-alert-example >}}

```yaml {title="configuration.yml"}
password_policy:
  standard:
    enabled: false
    min_length: 8
    max_length: 0
    require_uppercase: false
    require_lowercase: false
    require_number: false
    require_special: false
  zxcvbn:
    enabled: false
    min_score: 3
```

## Options

This section describes the individual configuration options.

### standard

This section allows you to enable standard security policies.

#### enabled

{{< confkey type="boolean" default="false" required="no" >}}

Enables standard password policy.

#### min_length

{{< confkey type="integer" default="8" required="no" >}}

Determines the minimum allowed password length.

#### max_length

{{< confkey type="integer" default="0" required="no" >}}

Determines the maximum allowed password length.

#### require_uppercase

{{< confkey type="boolean" default="false" required="no" >}}

Indicates that at least one UPPERCASE letter must be provided as part of the password.

#### require_lowercase

{{< confkey type="boolean" default="false" required="no" >}}

Indicates that at least one lowercase letter must be provided as part of the password.

#### require_number

{{< confkey type="boolean" default="false" required="no" >}}

Indicates that at least one number must be provided as part of the password.

#### require_special

{{< confkey type="boolean" default="false" required="no" >}}

Indicates that at least one special character must be provided as part of the password.

### zxcvbn

This password policy enables advanced password strength metering, using [zxcvbn](https://github.com/dropbox/zxcvbn).

Note that this password policy do not restrict the user's entry it just gives the user feedback as to how strong their
password is.

#### enabled

{{< confkey type="boolean" default="false" required="no" >}}

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
Only one password policy can be applied at a time.
{{< /callout >}}

Enables zxcvbn password policy.

#### min_score

{{< confkey type="integer" default="3" required="no" >}}

Configures the minimum zxcvbn score allowed for new passwords. There are 5 levels in the zxcvbn score system (taken from
[github.com/dropbox/zxcvbn](https://github.com/dropbox/zxcvbn#usage)):

* score 0: too guessable: risky password (guesses < 10^3)
* score 1: very guessable: protection from throttled online attacks (guesses < 10^6)
* score 2: somewhat guessable: protection from unthrottled online attacks. (guesses < 10^8)
* score 3: safely unguessable: moderate protection from offline slow-hash scenario. (guesses < 10^10)
* score 4: very unguessable: strong protection from offline slow-hash scenario. (guesses >= 10^10)

We do not allow score 0, if you set the `min_score` value to 0 instead the default will be used instead.
