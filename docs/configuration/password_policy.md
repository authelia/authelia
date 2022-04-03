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
    require_uppercase: true
    require_lowercase: true
    require_number: true
    require_special: true
  zxcvbn:
    enabled: false
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
type: bool
{: .label .label-config .label-purple } 
required: no
{: .label .label-config .label-green }
</div>
Enables standard password policy

#### min_length 
<div markdown="1">
type: integer
{: .label .label-config .label-purple } 
required: no
{: .label .label-config .label-green }
</div>
Determines the minimum allowed password length

#### max_length 
<div markdown="1">
type: integer
{: .label .label-config .label-purple } 
required: no
{: .label .label-config .label-green }
</div>
Determines the maximum allowed password length

#### require_uppercase 
<div markdown="1">
type: bool
{: .label .label-config .label-purple } 
required: no
{: .label .label-config .label-green }
</div>
Indicates that at least one UPPERCASE letter must be provided as part of the password

#### require_lowercase 
<div markdown="1">
type: bool
{: .label .label-config .label-purple } 
required: no
{: .label .label-config .label-green }
</div>
Indicates that at least one lowercase letter must be provided as part of the password

#### require_number 
<div markdown="1">
type: bool
{: .label .label-config .label-purple } 
required: no
{: .label .label-config .label-green }
</div>
Indicates that at least one number must be provided as part of the password

#### require_special 
<div markdown="1">
type: bool
{: .label .label-config .label-purple } 
required: no
{: .label .label-config .label-green }
</div>
Indicates that at least one special character must be provided as part of the password


### zxcvbn
This password policy enables advanced password strengh metering, using [Dropbox zxcvbn package](https://github.com/dropbox/zxcvbn).

Note that this password policy do not restrict the user's entry, just warns the user that if their password is too weak


#### enabled 
<div markdown="1">
type: bool
{: .label .label-config .label-purple } 
required: no
{: .label .label-config .label-green }
</div>
Enables standard password policy

Note:
* only one password policy can be applied at a time
