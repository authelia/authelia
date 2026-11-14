---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "App Authenticator"
description: "An overview of Authelia's App Authenticator second factor method which lets users approve sign in requests in the Authelia mobile app by matching a number."
summary: "Authelia provides a mobile app second factor method where the user approves a sign in by entering the number shown in the browser."
date: 2026-11-12T06:00:14+11:00
draft: false
images: []
weight: 255
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

The App Authenticator is a second factor method similar to [Duo / Mobile Push](../push-notification/index.md) which is
provided by Authelia itself and does not rely on a third party service. The user approves a sign in request in the
Authelia mobile app by entering the number which is shown in the browser.

The method is enabled by default. See the [configuration documentation](../../../configuration/second-factor/app-authenticator.md)
for the options, including the attestation settings which are required before devices can be registered.

## Registration

1. In the settings, open the two-factor authentication section and add a device. This requires an
   [elevated session](../../../configuration/identity-validation/elevated-session.md).
2. The portal shows a QR code. The app scans it, which tells the app which server to talk to.
3. Once the app has started the registration the portal shows a number and the name and platform of the device.
4. Enter the number in the app and unlock it with biometrics.
5. The device appears in the list. The QR code is single use and expires after a few minutes.

The app creates two keys for the device. An approval key which requires biometrics every time it is used, and a
transport key which is bound to the hardware where possible and signs the requests the app makes. Where the platform
supports it, the app also attests to the server that the keys were made by the genuine app on a genuine device.

## Signing In

1. Enter the username and password and choose the app method.
2. The portal shows a number.
3. Open the app. It checks for a pending request when opened, when brought to the foreground, and when refreshed. There
   are no push notifications.
4. The app shows the details of the request: the site, the client IP address, and the browser. Enter the number shown in
   the browser and approve with biometrics, or deny the request.
5. The browser notices the approval and continues to the original destination.

## Number Matching

The number has between 2 and 4 digits depending on the [nonce_length](../../../configuration/second-factor/app-authenticator.md#nonce_length)
option. This prevents a user approving a request they did not start, as the number is only visible to the browser which
made the request.

Each challenge allows exactly one attempt. Entering the wrong number, a bad signature, or any other failure voids the
challenge and the user starts again from the browser. The same applies to the number at registration. A guess therefore
succeeds with probability 1 in 10^n for an n digit number, which is why the length is configurable.

Only the browser session which created a challenge can use its approval, and an approval can only be used once.

## Authentication Method References

Signing in with the app reports the `pop`, `mfa`, and `mca` values, plus `hwk` for an attested device or `swk` for one
which is not, plus `user` when the device is attested. See
[Authentication Method References](../../../reference/guides/authentication-method-references.md).

## Device Management

The settings list the registered devices with the name, model, platform, when it was registered, and when it was last
used. Each device shows its attestation status: the verified app identity, or that the device is not attested (for
example when the [mode](../../../configuration/second-factor/app-authenticator.md#mode) is `preferred` or `none`).

An elevated session is required to rename or remove a device. A user can also remove a device from within the app.
Removing a device immediately stops it from being able to approve any request.

## Frequently Asked Questions

### Why can't I add a device?

With the default configuration the attestation mode is `required` and no trusted app identities are configured, so
registration is not available and the settings page says so. An administrator must configure the
[apple](../../../configuration/second-factor/app-authenticator.md#apple) or
[android](../../../configuration/second-factor/app-authenticator.md#android) identities, or change the
[mode](../../../configuration/second-factor/app-authenticator.md#mode).

### Why doesn't the app show my request?

The app only checks when it is opened, brought to the foreground, or refreshed. Open the app, and pull to refresh if the
request is not shown. A request expires after the [challenge_lifespan](../../../configuration/second-factor/app-authenticator.md#challenge_lifespan).
