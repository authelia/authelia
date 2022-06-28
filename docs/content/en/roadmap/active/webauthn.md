---
title: "WebAuthn"
description: "Authelia WebAuthn Implementation"
lead: "An introduction into the Authelia roadmap."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  roadmap:
    parent: "active"
weight: 220
toc: true
aliases:
  - /r/webauthn
---

[WebAuthn] requires urgent implementation as it is being deprecated by Chrome. It is a modern evolution of the
[FIDO U2F] protocol and is very similar in many ways. It even includes a backwards compatability extension called
the [FIDO AppID Extension] which allows a previously registered [FIDO U2F] device to be used with the protocol to
authenticate.

## Stages

This section represents the stages involved in implementation of this feature. The stages are either in order of
implementation due to there being an underlying requirement to implement them in this order, or in their likely order
due to how important or difficult to implement they are.

### Initial Implementation

{{< roadmap-status stage="complete" version="v4.34.0" >}}

Implement [WebAuthn] as a replacement for [FIDO U2F] with backwards compatibility.

|                       Setting                        |     Value      |                                                                Effect                                                                |
|:----------------------------------------------------:|:--------------:|:------------------------------------------------------------------------------------------------------------------------------------:|
|              [Conveyancing Preference]               |    indirect    | Configurable: ask users to permit collection of the AAGUID, this is like a model number, this GUID will be stored in the SQL storage |
|           [User Verification Requirement]            |   preferred    |                           Configurable: ask the browser to prompt for the users PIN or other verification                            |
|              [Resident Key Requirement]              |  discouraged   |                                       See the [passwordless login stage](#passwordless-login)                                        |
|              [Authenticator Attachment]              | cross-platform |                                   See the [platform authenticator stage](#platform-authenticator)                                    |
|              Multi-Device Registration               |  unavailable   |                                see the [multi device registration stage](#multi-device-registration)                                 |

### Multi Device Registration

{{< roadmap-status >}}

Implement multi device registration as part of the user interface. This is technically implemented for the most part in
the backend, it's just the public facing interface elements remaining.

### Platform Authenticator

{{< roadmap-status >}}

Implement [WebAuthn] Platform Authenticators so that people can use things like [Windows Hello], [TouchID], [FaceID],
or [Android Security Key]. This would also allow configuration of the [Authenticator Attachment] setting most likely,
or at least allow admins to configure which ones are available for registration.

### Passwordless Login

{{< roadmap-status >}}

Implement the [WebAuthn] flow for [Passwordless Login]. This would also allow configuration of the
[Resident Key Requirement] setting most likely, or at least allow admins to configure which ones are available for
registration.

[FIDO U2F]: https://fidoalliance.org/specs/u2f-specs-master/fido-u2f-overview.html
[WebAuthn]: https://www.w3.org/TR/webauthn-2/
[Passwordless Login]: https://www.w3.org/TR/webauthn-2/#client-side-discoverable-public-key-credential-source
[Conveyancing Preference]: https://www.w3.org/TR/webauthn-2/#enum-attestation-convey
[User Verification Requirement]: https://www.w3.org/TR/webauthn-2/#enum-userVerificationRequirement
[Resident Key Requirement]: https://www.w3.org/TR/webauthn-2/#enum-residentKeyRequirement
[Authenticator Attachment]: https://www.w3.org/TR/webauthn-2/#enum-attachment
[FIDO AppID Extension]: https://www.w3.org/TR/webauthn-2/#sctn-appid-extension

[Windows Hello]: https://support.microsoft.com/en-us/windows/learn-about-windows-hello-and-set-it-up-dae28983-8242-bb2a-d3d1-87c9d265a5f0
[TouchID]: https://support.apple.com/en-us/HT201371
[FaceID]: https://support.apple.com/en-au/HT208109
[Android Security Key]: https://support.google.com/accounts/answer/9289445?hl=en&co=GENIE.Platform%3DAndroid
