---
title: "Security"
description: "Report security issues."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
aliases:
  - /security
  - /security.html
---

The __Authelia__ team takes security very seriously. Because __Authelia__ is intended as a security product a lot of
decisions are made with security being the priority.

## Coordinated vulnerability disclosure

__Authelia__ follows the
[coordinated vulnerability disclosure](https://en.wikipedia.org/wiki/Coordinated_vulnerability_disclosure) model when
dealing with security vulnerabilities. This was previously known as responsible disclosure. We strongly urge anyone
reporting vulnerabilities to __Authelia__ or any other project to follow this model as it is considered as a best
practice by many in the security industry.

If you believe you have identified a security related bug with Authelia please do not open an issue, do not notify us in
public, and do not disclose this issue to third parties. Please use one of the [contact options](#contact-options)
below.

## Contact Options

### Email

Please utilize the [security@authelia.com](mailto:team@authelia.com) email address for security issues discovered. This
email address is only accessible by key members of the team for the purpose of disclosing security issues within the
__Authelia__ code base.

This is the preferred method of reporting.

### Chat

If you wish to chat directly instead of sending an email please use one of the [chat options](contact.md#chat) but it
is vital that when you do that you only do so privately with one of the maintainers. In order to start a private
discussion you should ask to have a private discussion with a team member without mentioning the reason why you wish to
have a private discussion so that provided the bug is confirmed we can coordinate the release of fixes and information
responsibly.

## Credit

Users who report bugs will optionally be credited for the discovery in the
[security advisory](https://github.com/authelia/authelia/security/advisories) and/or in our
[all contributors](https://github.com/authelia/authelia/blob/master/README.md#contribute) configuration/documentation.

## Process

1. User privately reports a potential vulnerability.
2. The core team reviews the report and ascertain if additional information is required.
3. The core team reproduces the bug.
4. The bug is patched, and if possible the user reporting te bug is given access to a fixed version or git patch.
5. The fix is confirmed to resolve the vulnerability.
6. The fix is released.
7. The security advisory is published sometime after users have had a chance to update.

## Help wanted

We are actively looking for sponsorship to obtain security audits to comprehensively ensure the security of Authelia.
As security is imperative to us we see this as one of the main financial priorities.

We believe that we should obtain the following categories of security audits:

* Code Security Audit / Analysis
* Penetration Testing

If you know of a company which either performs these kinds of audits and would be willing to sponsor the audit in some
way such as doing it pro bono or at a discounted rate, or wants to help improve Authelia in a meaningful way and is
willing to make a financial contribution towards this then please feel free to contact us.
