---
title: "WebAuthn"
description: "Configuring the WebAuthn Second Factor Method."
lead: "WebAuthn is the modern browser security key specification that Authelia supports. This section describes configuring it."
date: 2022-03-03T22:20:43+11:00
draft: false
images: []
menu:
  configuration:
    parent: "second-factor"
weight: 103400
toc: true
aliases:
  - /docs/configuration/webauthn.html
---

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

{{< confkey type="boolean" default="false" required="no" >}}

This disables WebAuthn if set to true.

### display_name

{{< confkey type="string" default="Authelia" required="no" >}}

Sets the display name which is sent to the client to be displayed. It's up to individual browsers and potentially
individual operating systems if and how they display this information.

See the [W3C WebAuthn Documentation](https://www.w3.org/TR/webauthn-2/#dom-publickeycredentialentity-name) for more
information.

### attestation_conveyance_preference

{{< confkey type="string" default="indirect" required="no" >}}

Sets the conveyance preference. Conveyancing allows collection of attestation statements about the authenticator such as
the AAGUID. The AAGUID indicates the model of the device.

See the [W3C WebAuthn Documentation](https://www.w3.org/TR/webauthn-2/#enum-attestation-convey) for more information.

Available Options:

|  Value   |                                                                  Description                                                                  |
|:--------:|:---------------------------------------------------------------------------------------------------------------------------------------------:|
|   none   |                                           The client will be instructed not to perform conveyancing                                           |
| indirect | The client will be instructed to perform conveyancing but the client can choose how to do this including using a third party anonymization CA |
|  direct  |               The client will be instructed to perform conveyancing with an attestation statement directly signed by the device               |

### user_verification

{{< confkey type="string" default="preferred" required="no" >}}

Sets the user verification preference.

See the [W3C WebAuthn Documentation](https://www.w3.org/TR/webauthn-2/#enum-userVerificationRequirement) for more information.

Available Options:

|    Value    |                                              Description                                               |
|:-----------:|:------------------------------------------------------------------------------------------------------:|
| discouraged |                    The client will be discouraged from asking for user verification                    |
|  preferred  |          The client if compliant will ask the user for verification if the device supports it          |
|  required   | The client will ask the user for verification or will fail if the device does not support verification |

### timeout

{{< confkey type="duration" default="60s" required="no" >}}

*__Note:__ This setting uses the [duration notation format](../prologue/common.md#duration-notation-format). Please see
the [common options](../prologue/common.md#duration-notation-format) documentation for information on this format.*

This adjusts the requested timeout for a WebAuthn interaction.

## FAQ

See the [Security Key FAQ](../../overview/authentication/security-key/index.md#faq) for the FAQ.
