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
  mode: classic
  min_length: 8
  max_length: 12
  require_uppercase: true
  require_lowercase: true
  require_number: true
  require_special: true
```

## Options

### mode
<div markdown="1">
type: string
{: .label .label-config .label-purple } 
default: none
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Determines the password policy mode:
#### none
* to password policy is applied
#### classic
* enables classic password policy
* allows to determine basic rules
#### zxcvbn
* enables advanced password strengh metering

### min_length
<div markdown="1">
type: integer
{: .label .label-config .label-purple } 
default: 8
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Determines the minimun password length for `mode=classic`

### max_length
<div markdown="1">
type: integer
{: .label .label-config .label-purple } 
default: 0
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Determines the maximum password length for `mode=classic`

### require_uppercase
<div markdown="1">
type: integer
{: .label .label-config .label-purple } 
default: false
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Determines that the password must contain at least one UPPERCASE letter.
Applies to  `mode=classic`


### require_lowercase
<div markdown="1">
type: integer
{: .label .label-config .label-purple } 
default: false
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Determines that the password must contain at least one lowercase letter.
Applies to  `mode=classic`

### require_number
<div markdown="1">
type: integer
{: .label .label-config .label-purple } 
default: false
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Determines that the password must contain at least one digit (0-9).
Applies to  `mode=classic`

### require_special
<div markdown="1">
type: integer
{: .label .label-config .label-purple } 
default: false
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Determines that the password must contain at least one special character.
Applies to  `mode=classic`
