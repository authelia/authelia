---
title: "Security Key"
description: "Authelia utilizes WebAuthn security keys as one of it's second factor authentication methods."
summary: "Authelia utilizes WebAuthn security keys as one of it's second factor authentication methods."
date: 2020-02-29T01:43:59+01:00
draft: false
images: []
weight: 240
toc: true
aliases:
  - /docs/features/2fa/security-key
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

__Authelia__ supports hardware-based second factors leveraging [FIDO2] [WebAuthn] compatible security keys like
[YubiKey]'s.

Security keys are among the most secure second factor. This method is already supported by many major applications and
platforms like Google, Facebook, GitHub, some banks, and much more.

{{< figure src="yubikey.jpg" caption="A YubiKey Security Key" alt="A YubiKey Security Key" width=150 >}}

Normally, the protocol requires your security key to be enrolled on each site before being able to authenticate with it.
Since Authelia provides Single Sign-On, your users will need to enroll their device only once to get access to all your
applications.

{{< figure src="REGISTER-U2F.png" caption="The WebAuthn Registration View" alt="2FA WebAuthn Registration View" width=400 >}}

After having successfully passed the first factor, select *Security Key* method and click on *Register device* link.
This will send you an email to verify your identity.

*NOTE: This e-mail has likely been sent to the mailbox at https://mail.example.com:8080/ if you're testing Authelia.*

Confirm your identity by clicking on __Register__ and you'll be asked to touch the token of your security key to
complete the enrollment.

Upon successful enrollment, you can authenticate using your security key by simply touching the token again when
requested:

{{< figure src="2FA-U2F.png" caption="The WebAuthn Authentication View" alt="2FA WebAuthn Authentication View" width=400 >}}

Easy, right?!

## Frequently Asked Questions

### Can I register multiple FIDO2 WebAuthn credentials?

Yes, as of v4.38.0 and above Authelia supports registering multiple WebAuthn credentials as per the
[roadmap](../../../roadmap/active/webauthn.md#multi-device-registration).

### Can I perform a passwordless login?

Yes, as of v4.39.0 and above Authelia supports passwordless logins via Passkeys as per the
[roadmap](../../../roadmap/active/webauthn.md#passwordless-login).

### Why don't I have access to the *Security Key* option?

The [WebAuthn] protocol is a new protocol that is only supported by modern browsers. Please ensure your browser is up to
date, supports [WebAuthn], and that the feature is not disabled if the option is not available to you in __Authelia__.

### Can my FIDO U2F device operate with Authelia?

At the present time there is no plan to support [FIDO U2F] within Authelia. We do implement a backwards compatible appid
extension within __Authelia__ however this only works for devices registered before the upgrade to the [FIDO2]
[WebAuthn] protocol.

If there was sufficient interest in supporting registration of old U2F / FIDO devices in __Authelia__ we would consider
adding support for this after or at the same time of the multi-device enhancements.

[FIDO U2F]: https://www.yubico.com/authentication-standards/fido-u2f/
[FIDO2]: https://www.yubico.com/authentication-standards/fido2/
[WebAuthn]: https://www.yubico.com/authentication-standards/webauthn/
[YubiKey]: https://www.yubico.com/products/yubikey-5-overview/
