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
  debug: false
  display_name: Authelia
  conveyance_preference: indirect
  user_verification: preferred
  timeout: 60000
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

### debug
<div markdown="1">
type: boolean
{: .label .label-config .label-purple } 
default: false
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

This enables some additional debug messaging if set to true.

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

### timeout
<div markdown="1">
type: integer
{: .label .label-config .label-purple } 
default: 60000
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

This adjusts the requested timeout for a Webauthn interaction.

### Can I register multiple FIDO2 Webauthn devices?

At present this is not possible in the frontend. However the backend technically supports it. We plan to add this to the
frontend in the near future. Subscribe to [this issue](https://github.com/authelia/authelia/issues/275) for updates.

### Can I perform a passwordless login?

Not at this time. We will tackle this at a later date.

### Why don't I have access to the *Security Key* option?

The [Webauthn] protocol is a new protocol that is only supported by modern browsers. Please ensure your browser is up to
date, supports [Webauthn], and that the feature is not disabled if the option is not available to you in **Authelia**.

### Can my FIDO U2F device operate with Authelia?

At the present time there is no plan to support [FIDO U2F] within Authelia. We do implement a backwards compatible appid
extension within **Authelia** however this only works for devices registered before the upgrade to the [FIDO2]&nbsp;[Webauthn]
protocol.

If there was sufficient interest in supporting registration of old U2F / FIDO devices in **Authelia** we would consider
adding support for this after or at the same time of the multi-device enhancements.