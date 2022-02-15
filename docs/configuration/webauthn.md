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
  conveyance_preference: indirect
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

Sets the display name which is sent to the client to be displayed.

### conveyance_preference
<div markdown="1">
type: string
{: .label .label-config .label-purple } 
default: indirect
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Sets the conveyance preference. Conveyancing allows collection of attestation statements about the authenticator such as
the AAGUID. The AAGUID indicates the model of the device.

Available Options:

|  Value   |                                                                  Description                                                                  |
|:--------:|:---------------------------------------------------------------------------------------------------------------------------------------------:|
|   none   |                                           The client will be instructed not to perform conveyancing                                           |
| indirect | The client will be instructed to perform conveyancing but the client can choose how to do this including using a third party anonymization CA |
|  direct  |               The client will be instructed to perform conveyancing with an attestation statement directly signed by the device               |

### user_verification
<div markdown="1">
type: string
{: .label .label-config .label-purple } 
default: preferred
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Sets the user verification preference. 

Available Options:

|    Value    |                                              Description                                               |
|:-----------:|:------------------------------------------------------------------------------------------------------:|
| discouraged |                    The client will be discouraged from asking for user verification                    |
|  preferred  |          The client if compliant will ask the user for verification if the device supports it          |
|  required   | The client will ask the user for verification or will fail if the device does not support verification |