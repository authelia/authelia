---
layout: default
title: Default
parent: Configuration
nav_order: 3
---

This section sets default values.

## Configuration

The configuration is as follows:
```yaml
default:
  user_second_factor_method: totp
```

## Options

### user_second_factor_method
<div markdown="1">
type: string
{: .label .label-config .label-purple } 
default: ""
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Sets the default second factor method. If this method is available, new users as well as users who have a disabled
method, will default to this method.

Options are:

- totp
- webauthn
- mobile_push