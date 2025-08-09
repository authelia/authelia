---
title: "WebAuthn"
description: "Configuring the WebAuthn Second Factor Method."
summary: "WebAuthn is the modern browser security key specification that Authelia supports. This section describes configuring it."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 103400
toc: true
aliases:
  - '/docs/configuration/webauthn.html'
  - '/configuration/webauthn.html'
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Configuration

{{< config-alert-example >}}

```yaml {title="configuration.yml"}
webauthn:
  disable: false
  enable_passkey_login: false
  display_name: 'Authelia'
  attestation_conveyance_preference: 'indirect'
  timeout: '60 seconds'
  filtering:
    permitted_aaguids: []
    prohibited_aaguids: []
    prohibit_backup_eligibility: false
  selection_criteria:
    attachment: ''
    discoverability: 'preferred'
    user_verification: 'preferred'
  metadata:
    enabled: false
    validate_trust_anchor: true
    validate_entry: true
    validate_entry_permit_zero_aaguid: false
    validate_status: true
    validate_status_permitted: []
    validate_status_prohibited:
      - 'REVOKED'
      - 'USER_KEY_PHYSICAL_COMPROMISE'
      - 'USER_KEY_REMOTE_COMPROMISE'
      - 'USER_VERIFICATION_BYPASS'
      - 'ATTESTATION_KEY_COMPROMISE'
```

## Options

This section describes the individual configuration options.

### disable

{{< confkey type="boolean" default="false" required="no" >}}

This disables WebAuthn if set to true.

### enable_passkey_login

{{< confkey type="boolean" default="false" required="no" >}}

Enables login via a Passkey instead of a username and password. This login only counts as a single factor. The user will
be prompted for their password by default if the request requires multi-factor authentication.

### experimental_enable_passkey_uv_two_factors

{{< confkey type="boolean" default="false" required="no" >}}

{{< callout context="danger" title="Stability and Security Notice" icon="outline/alert-octagon" >}}
This option is not considered supported. It is completely experimental and will be replaced by
[custom policies](../../roadmap/active/granular-authorization.md#custom-policy-flows) that can be defined in the access
control section and allow deterministic results for authentication. This is in an effort to properly support
Authentication Method Reference Values for authentication flows. The likely versions this will release in will be
detailed in the [roadmap item](../../roadmap/active/granular-authorization.md#custom-policy-flows) (specifically at the
time we add the flow to replace this) that this option will cause a startup failure.
{{< /callout >}}

This option allows for authenticators that enforce user verification (PIN entry, biometric proof, etc) and reports they
have performed user verification, to satisfy the `two_factor` policy for access control rules.

### experimental_enable_passkey_upgrade

{{< confkey type="boolean" default="false" required="no" >}}

{{< callout context="danger" title="Stability Notice" icon="outline/alert-octagon" >}}
This option is not considered supported. It is completely experimental and will be removed in the future. It will either
become a deliberate flow within the UI, be automatic, or be entirely removed.
{{< /callout >}}

Some authenticators or browsers have issues with reporting if a WebAuthn credential is actually a Passkey or not. The
relevant specifications instruct implementers to assume they are not a Passkey in this instance. However this presents
some user frustration. This option attempts to alleviate some of that frustration.

When true this will attempt to automatically upgrade a WebAuthn credential to a Passkey when the browser or
authenticator does not report a newly registered WebAuthn credential to be a Passkey.

### display_name

{{< confkey type="string" default="Authelia" required="no" >}}

Sets the display name which is sent to the client to be displayed. It's up to individual browsers and potentially
individual operating systems if and how they display this information.

See the [W3C WebAuthn Documentation](https://www.w3.org/TR/webauthn-2/#dom-publickeycredentialentity-name) for more
information.

### attestation_conveyance_preference

{{< confkey type="string" default="indirect" required="no" >}}

Sets the conveyance preference. Conveyancing allows collection of attestation statements about the authenticator such as
the AAGUID. The AAGUID indicates the model of the authenticator.

See the [W3C WebAuthn Documentation](https://www.w3.org/TR/webauthn-2/#enum-attestation-convey) for more information.

Available Options:

|  Value   |                                                                  Description                                                                  |
|:--------:|:---------------------------------------------------------------------------------------------------------------------------------------------:|
|   none   |                                           The client will be instructed not to perform conveyancing                                           |
| indirect | The client will be instructed to perform conveyancing but the client can choose how to do this including using a third party anonymization CA |
|  direct  |           The client will be instructed to perform conveyancing with an attestation statement directly signed by the authenticator            |

### timeout

{{< confkey type="string,integer" syntax="duration" default="60 seconds" required="no" >}}

This adjusts the requested timeout for a WebAuthn interaction.

### filtering

This section configures various filtering options during registration.

#### permitted_aaguids

{{< confkey type="list(string)" syntax="uuid" required="no" >}}

A list of Authenticator Attestation GUID's that are the only ones allowed to be registered. Useful if you have a company
policy that requires certain authenticators. Mutually exclusive with [prohibited_aaguids](#prohibited_aaguids).

#### prohibited_aaguids

{{< confkey type="list(string)" syntax="uuid" required="no" >}}

A list of Authenticator Attestation GUID's that users will not be able to register. Useful if company policy prevents
certain authenticators. Mutually exclusive with [permitted_aaguids](#permitted_aaguids).

#### prohibit_backup_eligibility

{{< confkey type="boolean" default="false" required="no" >}}

Setting this value to true will ensure Authenticators which can export credentials will not be able to register. This
will likely prevent synchronized credentials from being registered.

### selection_criteria

The selection criteria options set preferences for selecting a suitable authenticator.

#### attachment

{{< confkey type="string" default="" required="no" >}}

Sets the attachment preference for newly created credentials.

Available Options:

|     Value      |                                           Description                                           |
|:--------------:|:-----------------------------------------------------------------------------------------------:|
|    _empty_     | The Authenticators that are available will be shown and the user can pick the specific criteria |
| cross-platform |     Authenticators that can move from one system to another such as physical security keys      |
|    platform    |        Authenticators that are part of the platform such as Windows Hello, AppleID, etc         |

#### discoverability

{{< confkey type="string" default="preferred" required="no" >}}

Sets the discoverability preference. May affect the creation of Passkeys.

|    Value    |                             Description                             |
|:-----------:|:-------------------------------------------------------------------:|
| discouraged |                     Prefers no discoverability                      |
|  preferred  | Prefers discoverability and will not error if it's not discoverable |
|  required   |   Requires discoverability and may error if it's not discoverable   |

#### user_verification

{{< confkey type="string" default="preferred" required="no" >}}

Sets the user verification preference.

See the [W3C WebAuthn Documentation](https://www.w3.org/TR/webauthn-2/#enum-userVerificationRequirement) for more information.

Available Options:

|    Value    |                                                  Description                                                  |
|:-----------:|:-------------------------------------------------------------------------------------------------------------:|
| discouraged |                       The client will be discouraged from asking for user verification                        |
|  preferred  |          The client if compliant will ask the user for verification if the authenticator supports it          |
|  required   | The client will ask the user for verification or will fail if the authenticator does not support verification |

### metadata

Configures the metadata service which is used to check the authenticity of authenticators. Useful if company policy
requires only conformant authenticators.

See the [reference guide](../../reference/guides/webauthn.md#recommended-configurations) for the recommended
configuration.

#### enabled

{{< confkey type="boolean" default="false" required="no" >}}

Enables metadata service validation of authenticators and credentials. This requires the download of the metadata
service blob which will utilize about 5MB of data in your configured [storage](../storage/introduction.md) backend.

By default to prevent breaking changes this value is false. It's recommended however users take the time to configure
it now that it's available.

#### cache_policy

{{< confkey type="string" default="strict" required="no" >}}

Changes the mode by which the metadata service cache operates. The options are `strict` and `relaxed`. The `strict`
policy will try to download a fresh copy at startup and will error if this is not possible. The `relaxed` policy
will still try to download a fresh copy at startup, however provided there is a cached version of the metadata service
data and it's not expired, it will just log an error if the metadata service returns a `429 Too Many Requests` status
code.

#### validate_trust_anchor

{{< confkey type="boolean" default="true" required="no" >}}

Enables validation of the attestation certificate against the Certificate Authority certificate in the validated MDS3
blob. It's recommended this value is always the default value.

#### validate_entry

{{< confkey type="boolean" default="true" required="no" >}}

Enables validation that an entry exists for the authenticator in the MDS3 blob. It's recommended that this option is
the default value, however this may exclude some authenticators which **_DO NOT_** have FIDO compliance certification or
have otherwise not registered with the MDS3. The recommendation is based on the fact that the authenticity of a
particular authenticator cannot be validated without this.

#### validate_entry_permit_zero_aaguid

{{< confkey type="boolean" default="false" required="no" >}}

Allows authenticators which have provided an empty Authenticator Attestation GUID. This may be required for certain
authenticators which **_DO NOT_** have FIDO compliance certification.

#### validate_status

{{< confkey type="boolean" default="true" required="no" >}}

Enables validation of the attestation entry statuses. There is generally never a reason to disable this as the
authenticators excluded by default are likely compromised.

#### validate_status_permitted

{{< confkey type="list(string)" required="no" >}}

A list of exclusively required statuses for an authenticator to pass validation. See the
[reference guide](../../reference/guides/webauthn.md#metadata-status) for information on valid values.

#### validate_status_prohibited

{{< confkey type="list(string)" required="no" >}}

A list of authenticator statuses which for an authenticator that are prohibited from being registered. See the
[reference guide](../../reference/guides/webauthn.md#metadata-status) for information on valid values. It's strongly
recommended not changing the default value.

The default configuration for this option is as per the [Configuration](#configuration) example above.

## Frequently Asked Questions

See the [Security Key FAQ](../../overview/authentication/security-key/index.md#frequently-asked-questions) for the FAQ.
