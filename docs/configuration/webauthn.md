---
layout: default
title: Webauthn
parent: Configuration
nav_order: 16
---

The webauthn section has tunable options for the Webauthn implementation.

## Configuration
```yaml
webauthn:
  display_name: Authelia
  attestation_preference: indirect
  user_verification: preferred
```

## Options

### display_name
<div markdown="1">
type: string
{: .label .label-config .label-purple } 
default: Authelia
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Sets the display name which is displayed to the user during attestation or assertion. 