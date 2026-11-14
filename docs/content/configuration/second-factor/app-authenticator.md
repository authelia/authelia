---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "App Authenticator"
description: "Configuring the Authelia App Authenticator second factor method including the display name, lifespans, number length, and device attestation."
summary: ""
date: 2026-11-12T06:00:14+11:00
draft: false
images: []
weight: 103250
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

Authelia supports approving sign in requests with the Authelia mobile app. The user is shown a number in the browser
and enters it in the app after unlocking their device.

See the dedicated [documentation](../../overview/authentication/app-authenticator/index.md) for an overview of how the
method works.

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
This method is enabled by default. With the default configuration the [attestation](#attestation) mode is `required`
and no trusted app identities are configured, which means users can not register a device until the
[apple](#apple) or [android](#android) lists are configured or the [mode](#mode) is set to `preferred` or `none`. Authelia logs a warning at
startup when this is the case and the user settings page tells the user registration is not configured.
{{< /callout >}}

## Configuration

{{< config-alert-example >}}

```yaml {title="configuration.yml"}
app_authenticator:
  disable: false
  display_name: 'Authelia'
  registration_token_lifespan: '5 minutes'
  challenge_lifespan: '2 minutes'
  nonce_length: 2
  attestation:
    mode: 'required'
    apple:
      - team_id: 'ABCDE12345'
        bundle_id: 'com.example.app'
        environment: 'production'
    android:
      - package_name: 'com.example.app'
        signing_certificate_digests:
          - 'ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789'
```

## Options

This section describes the individual configuration options.

### disable

{{< confkey type="boolean" default="false" required="no" >}}

Disables the App Authenticator method. When disabled the method is not offered, registration is hidden, and none of the
associated endpoints are available.

### display_name

{{< confkey type="string" default="Authelia" required="no" >}}

The name the app shows for this server. An empty value uses the default.

### registration_token_lifespan

{{< confkey type="string,integer" syntax="duration" default="5 minutes" required="no" >}}

How long a registration QR code remains valid, and how long a registration remains valid once the app has started it.
Must be positive.

### challenge_lifespan

{{< confkey type="string,integer" syntax="duration" default="2 minutes" required="no" >}}

How long a sign in challenge remains valid. Must be positive. A value over 5 minutes is accepted but logs a warning.

### nonce_length

{{< confkey type="integer" default="2" required="no" >}}

The number of digits in the number the user enters in the app, for both registration and sign in. Must be between 2 and 4.

A person who guesses the number has exactly one attempt per challenge or registration, as any wrong number voids it. The
odds of a blind guess succeeding are 1 in 10^n, so 1 in 100 for the default and 1 in 10000 for 4. A longer number
reduces the chance at the cost of more typing. Changing this option does not affect a challenge or registration that has
already started.

### attestation

Device attestation proves to Authelia that the key was created by a genuine copy of the app on a genuine device. It is
checked once, when the device is registered.

#### mode

{{< confkey type="string" default="required" required="no" >}}

The attestation policy. The following table describes how each mode treats the format the app presents at
registration. A rejected registration is voided and the user must start again.

| Mode        | `none` | `apple-appattest` / `android-key` | Any other format |
| ----------- | ------ | --------------------------------- | ---------------- |
| `required`  | Reject | Verify, reject on failure         | Reject           |
| `preferred` | Accept | Verify, reject on failure         | Reject           |
| `none`      | Accept | Reject                            | Reject           |

Notes:

- With `preferred` a device which sends no attestation, for example a simulator or unsupported hardware, is accepted and
  recorded as not attested. If it does present an attestation which fails to verify it is rejected.
- With `none` attestation is not verified at all and Authelia logs a warning at startup. This is intended for testing
  and CI. Devices registered this way are shown as not attested in the user settings.
- With `required` at least one identity must be present in [apple](#apple) or [android](#android) for registration to be
  available. Until the official app identities are published this is the default state: validation does not fail, but
  Authelia logs a warning at startup, registration requests are refused, and the user settings page displays that
  registration of the app is not configured on the server. Configure an identity or change the mode to resolve this.

#### apple

{{< confkey type="list(object)" required="no" >}}

The trusted iOS app identities verified with Apple App Attest. A device must present an attestation from an app that
matches one of the entries. Attestations are verified against the pinned Apple App Attestation root certificate.

##### team_id

{{< confkey type="string" required="yes" >}}

The 10 character Apple developer team identifier. Must consist of the characters `A-Z` and `0-9`.

##### bundle_id

{{< confkey type="string" required="yes" >}}

The bundle identifier of the app.

##### environment

{{< confkey type="string" default="production" required="no" >}}

The App Attest environment. Valid values are `production` and `development`.

#### android

{{< confkey type="list(object)" required="no" >}}

The trusted Android app identities verified with Android Key Attestation. Certificate chains must lead to one of the
pinned Google attestation roots, which are trusted by public key, and must not be listed in the Google attestation
revocation list. Authelia caches the revocation list for 24 hours. If it can not be refreshed the last good copy is used
for up to 7 days after which attestation fails closed, i.e. Android devices can not be registered until the list can be
retrieved.

##### package_name

{{< confkey type="string" required="yes" >}}

The Android package name of the app.

##### signing_certificate_digests

{{< confkey type="list(string)" required="yes" >}}

The SHA-256 digests of the certificates the app is signed with. At least one is required. With Play App Signing this is
the digest of the Google app signing certificate. Each digest is accepted as 64 hexadecimal characters or as 32
colon separated pairs (the formats printed by the Play Console and `apksigner`) in either case, and is normalized to
lowercase without separators.

## Examples

### Testing and CI

This configuration disables attestation so a software device can be registered. It should not be used in production.

```yaml {title="configuration.yml"}
app_authenticator:
  nonce_length: 2
  attestation:
    mode: 'none'
```

A suite can then drive the registration with the fake device from the app repository. The registration QR text and the
numbers are read from the portal, and every command requires `--insecure-software-keys`:

```bash
pnpm fake-device register-begin "$QR" --state dev.json --insecure-software-keys
pnpm fake-device register-confirm --state dev.json --nonce "$REGISTRATION_NUMBER" --insecure-software-keys
pnpm fake-device check --state dev.json --insecure-software-keys
pnpm fake-device approve --state dev.json --nonce "$LOGIN_NUMBER" --insecure-software-keys
pnpm fake-device check --state dev.json --insecure-software-keys
pnpm fake-device deny --state dev.json --insecure-software-keys
pnpm fake-device remove --state dev.json --insecure-software-keys
```

Against a listener without TLS add `--allow-http` to `register-begin` and use the `authelia-app+http://` form of the QR
text. For a private certificate authority set `NODE_EXTRA_CA_CERTS` to the CA bundle.

### Own signed build

```yaml {title="configuration.yml"}
app_authenticator:
  attestation:
    mode: 'required'
    apple:
      - team_id: 'ABCDE12345'
        bundle_id: 'com.example.app'
        environment: 'production'
    android:
      - package_name: 'com.example.app'
        signing_certificate_digests:
          - 'AB:CD:EF:01:23:45:67:89:AB:CD:EF:01:23:45:67:89:AB:CD:EF:01:23:45:67:89:AB:CD:EF:01:23:45:67:89'
```

## Rate Limits

The endpoints used by the app are rate limited per client IP address. See
[Server Endpoint Rate Limits](../miscellaneous/server-endpoint-rate-limits.md) for the
[app_authenticator_registration](../miscellaneous/server-endpoint-rate-limits.md#app_authenticator_registration),
[app_authenticator_check](../miscellaneous/server-endpoint-rate-limits.md#app_authenticator_check),
[app_authenticator_confirmation](../miscellaneous/server-endpoint-rate-limits.md#app_authenticator_confirmation), and
[second_factor_app](../miscellaneous/server-endpoint-rate-limits.md#second_factor_app) limits.
