---
layout: default
title: Webauthn
parent: Configuration
nav_order: 16
---

The Webauthn section has tunable options for the Webauthn implementation.

## Configuration
```yaml
webauthn:
  disable: false
  display_name: Authelia
  attestation_conveyance_preference: indirect
  user_verification: preferred
  timeout: 60s
```

## Options

### disable
<div markdown="1">
type: boolean
{: .label .label-config .label-purple } 
default: false
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

This disables Webauthn if set to true.

### display_name
<div markdown="1">
type: string
{: .label .label-config .label-purple } 
default: Authelia
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Sets the display name which is sent to the client to be displayed. It's up to individual browsers and potentially
individual operating systems if and how they display this information.

See the [W3C Webauthn Documentation](https://www.w3.org/TR/webauthn-2/#dom-publickeycredentialentity-name) for more information.

### attestation_conveyance_preference
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

See the [W3C Webauthn Documentation](https://www.w3.org/TR/webauthn-2/#enum-attestation-convey) for more information.

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

See the [W3C Webauthn Documentation](https://www.w3.org/TR/webauthn-2/#enum-userVerificationRequirement) for more information.

Available Options:

|    Value    |                                              Description                                               |
|:-----------:|:------------------------------------------------------------------------------------------------------:|
| discouraged |                    The client will be discouraged from asking for user verification                    |
|  preferred  |          The client if compliant will ask the user for verification if the device supports it          |
|  required   | The client will ask the user for verification or will fail if the device does not support verification |

### timeout
<div markdown="1">
type: string (duration) 
{: .label .label-config .label-purple } 
default: 60s
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

This adjusts the requested timeout for a Webauthn interaction. The period of time is in
[duration notation format](index.md#duration-notation-format).

## FAQ

See the [Security Key FAQ](../features/2fa/security-key.md#faq) for the FAQ.
