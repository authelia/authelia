---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Events"
description: "A reference of every notification Authelia sends."
summary: "Authelia sends a fixed set of notifications. This section lists each one, when it is sent, and which template renders it."
date: 2026-09-10T10:00:00+10:00
draft: false
images: []
weight: 108400
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

Authelia sends a fixed set of notifications. Every notification is addressed to the email address recorded for the user
in the [authentication backend](../first-factor/introduction.md), is rendered from one of three
[templates](../../reference/guides/notification-templates.md), and is delivered by whichever
[notifier](introduction.md) you have configured.

Authelia does not send marketing, digest, or administrative messages, and there is no per-event configuration option to
enable or disable an individual notification.

## Events

|            Event            |              Subject              |         Template          |
| :-------------------------: | :-------------------------------: | :-----------------------: |
|  Password reset requested   |       `Reset your password`       | `IdentityVerificationJWT` |
| Session elevation requested |      `Confirm your identity`      | `IdentityVerificationOTC` |
|  Password reset completed   |  `Password changed successfully`  |          `Event`          |
| Password reset unsuccessful | `Password reset was unsuccessful` |          `Event`          |
|      Password changed       |  `Password changed successfully`  |          `Event`          |
|  Second factor registered   |   `Second Factor Method Added`    |          `Event`          |
|    Second factor removed    |  `Second Factor Method Removed`   |          `Event`          |

The subject line is also used as the `{{ .Title }}` placeholder within the template.

### Password reset requested

Sent when a user requests a password reset from the sign-in portal. The notification contains a single-use link which
proves the recipient controls the mailbox, and a second link which revokes the request.

This notification is sent whether or not the requested user exists, from the perspective of the requester: the portal
always responds identically in order to prevent user enumeration. No notification is sent when the user does not exist.

### Session elevation requested

Sent when a user attempts an operation which requires an elevated session, such as managing their credentials, and
[identity validation](../identity-validation/introduction.md) requires them to confirm their identity. The notification
contains a One-Time Code rather than a link, and a link which revokes the elevation request.

### Password reset completed

Sent to the user after their password has been successfully changed via the password reset flow.

### Password reset unsuccessful

Sent when a password reset fails for a reason the user can act on. The body names the cause: the new password did not
meet the password policy, did not meet the requirements of the authentication backend, matched the password already set
on the account, or was changed too recently to be changed again.

Nothing is sent for any other cause. A failure the user cannot resolve, such as the authentication backend being
unreachable, is logged and no notification is sent, so this never discloses the state of the backend to whoever holds
the mailbox. The response shown in the browser stays generic either way.

### Password changed

Sent to the user after they have successfully changed their own password from the settings area.

### Second factor registered

Sent when a user registers a new second factor method. The body names which method was registered, either a
One-Time Password or a WebAuthn Credential, and in the case of a WebAuthn Credential the description the user gave it.

### Second factor removed

Sent when a user removes a registered second factor method. As with registration, the body names the method and, for a
WebAuthn Credential, its description.

## Delivery

What a recipient actually receives depends on the configured notifier.

|                  Notifier                  |                                          Delivery                                           |
| :----------------------------------------: | :-----------------------------------------------------------------------------------------: |
|              [SMTP](smtp.md)               |           A multipart message containing both the plaintext and HTML renderings.            |
| [SMTP](smtp.md) with `disable_html_emails` |                                The plaintext rendering only.                                |
|           [Filesystem](file.md)            | The plaintext rendering only, preceded by a header naming the date, recipient, and subject. |

The [filesystem](file.md) notifier is intended for testing and development. It writes every notification to a single
file rather than delivering it to the user, so a deployment using it sends the user nothing.

## Customizing

The content of each notification is controlled by its template. To override one, set
[template_path](introduction.md#template_path) and provide a file named after the template with either a `.html` or
`.txt` extension.

See the [Notification Templates](../../reference/guides/notification-templates.md) reference guide for the available
templates, the placeholder variables each one accepts, and the functions available within them.
