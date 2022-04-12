---
layout: default
title: Password Policy
parent: Configuration
nav_order: 17
---

# Password Policy

_Authelia_ allows administrators to configure an enforced password policy.

## Configuration

```yaml
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

### standard
<div markdown="1">
type: list
{: .label .label-config .label-purple }
required: no
{: .label .label-config .label-green }
</div>

This section allows you to enable standard security policies.

#### enabled
<div markdown="1">
type: boolean
{: .label .label-config .label-purple }
default: false
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Enables standard password policy.

#### min_length
<div markdown="1">
type: integer
{: .label .label-config .label-purple }
default: 8
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Determines the minimum allowed password length.

#### max_length
<div markdown="1">
type: integer
{: .label .label-config .label-purple }
default: 0
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Determines the maximum allowed password length.

#### require_uppercase
<div markdown="1">
type: boolean
{: .label .label-config .label-purple }
default: false
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Indicates that at least one UPPERCASE letter must be provided as part of the password.

#### require_lowercase
<div markdown="1">
type: boolean
{: .label .label-config .label-purple }
default: false
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Indicates that at least one lowercase letter must be provided as part of the password.

#### require_number
<div markdown="1">
type: boolean
{: .label .label-config .label-purple }
default: false
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Indicates that at least one number must be provided as part of the password.

#### require_special
<div markdown="1">
type: boolean
{: .label .label-config .label-purple }
default: false
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Indicates that at least one special character must be provided as part of the password.

### zxcvbn

This password policy enables advanced password strength metering, using [zxcvbn](https://github.com/dropbox/zxcvbn).

#### enabled
<div markdown="1">
type: boolean
{: .label .label-config .label-purple }
default: false
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

_**Important Note:** only one password policy can be applied at a time._

Enables zxcvbn password policy.

#### min_score
<div markdown="1">
type: integer
{: .label .label-config .label-purple }
default: 3
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Configures the minimum zxcvbn score allowed for new passwords. There are 5 levels in the zxcvbn score system (taken from [github.com/dropbox/zxcvbn](https://github.com/dropbox/zxcvbn#usage)):

- score 0: too guessable: risky password (guesses < 10^3)
- score 1: very guessable: protection from throttled online attacks (guesses < 10^6)
- score 2: somewhat guessable: protection from unthrottled online attacks. (guesses < 10^8)
- score 3: safely unguessable: moderate protection from offline slow-hash scenario. (guesses < 10^10)
- score 4: very unguessable: strong protection from offline slow-hash scenario. (guesses >= 10^10)

We do not allow score 0, if you set the `min_score` value to 0 instead the default will be chosen.