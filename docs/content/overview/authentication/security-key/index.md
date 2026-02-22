---
title: "Security Key and Passkeys"
description: "Authelia utilizes WebAuthn Crefentials as one of it's second factor authentication methods and a passwordless login method via Passkeys."
summary: "Authelia utilizes WebAuthn Crefentials as one of it's second factor authentication methods and a passwordless login method via Passkeys."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 240
toc: true
aliases:
  - '/docs/features/2fa/security-key'
  - '/overview/authentication/webauthn-security-key'
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

__Authelia__ supports hardware-based second factors leveraging [FIDO2] [WebAuthn] compatible security keys like
[YubiKey]'s, or software-based second factors leveraging Passkeys.

Security keys are among the most secure second factor. This method is already supported by many major applications and
platforms like Google, Facebook, GitHub, some banks, and much more.

{{< figure src="yubikey.jpg" caption="A YubiKey Security Key" alt="A YubiKey Security Key" sizes="30dvh" >}}

Normally, the protocol requires your security key to be enrolled on each site before being able to authenticate with it.
Since Authelia provides Single Sign-On, your users will need to enroll their device only once to get access to all your
applications.

{{< figure src="REGISTER-U2F.png" caption="The WebAuthn Registration View" alt="2FA WebAuthn Registration View" sizes="50dvh" >}}

After having successfully passed the first factor, select *Security Key* method and click on *Register device* link.
This will send you an email to verify your identity.

*NOTE: This e-mail has likely been sent to the mailbox at https://mail.example.com:8080/ if you're testing Authelia.*

Confirm your identity by clicking on __Register__ and you'll be asked to touch the token of your security key to
complete the enrollment.

Upon successful enrollment, you can authenticate using your security key by simply touching the token again when
requested:

{{< figure src="2FA-U2F.png" caption="The WebAuthn Authentication View" alt="2FA WebAuthn Authentication View" sizes="50dvh" >}}

Easy, right?!

## Frequently Asked Questions

### Can I register multiple FIDO2 WebAuthn credentials?

Yes, as of v4.38.0 and above Authelia supports registering multiple WebAuthn credentials as per the
[roadmap](../../../roadmap/complete/webauthn.md#multi-device-registration).

### Can I perform a passwordless login?

Yes, as of v4.39.0 and above Authelia supports passwordless logins via Passkeys as per the
[roadmap](../../../roadmap/complete/webauthn.md#passwordless-login).

{{< figure src="passkeys.png" caption="The Passkey Authentication Portal View" alt="The Passkey Authentication Portal View" sizes="50dvh" >}}

### Why does it ask me for my password after using a Passkey to login?

This exists to ensure the `two_factor` policy is enforced. The Passkey itself is a single factor and we do have plans to
offer very [granular control policies and their requirements](../../../roadmap/active/granular-authorization.md). For
example it will likely be possible to create your own custom policy equal to `two_factor` today which also considers a
single Passkey login as satisfactory for a particular access control policy.

{{< figure src="password_2fa.png" caption="The Passkey MFA Password Authentication Portal View" alt="The Passkey MFA Password Authentication Portal View" width=400 sizes="50dvh" >}}

In the meantime the [configuration](../../../configuration/second-factor/webauthn.md) has an experimental option to
allow Passkey authenticators which support user verification, that perform user verification, and that report they
performed user verification; to count as satisfying `two_factor`.

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
