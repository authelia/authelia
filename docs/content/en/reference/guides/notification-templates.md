---
title: "Notification Templates"
description: "A reference guide on overriding notification templates"
lead: "This section contains reference documentation for Authelia's notification templates."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  reference:
    parent: "guides"
weight: 220
toc: true
---

Authelia uses templates to generate the HTML and plaintext emails sent via the notification service. Each template has
two extensions; `.html` for HTML templates, and `.txt` for plaintext templates.

This guide effectively documents the usage of the
[template_path](../../configuration/notifications/introduction.md#template_path) notification configuration option.

## Template Names

|       Template       |                                    Description                                    |
|:--------------------:|:---------------------------------------------------------------------------------:|
| IdentityVerification | Used to render notifications sent when registering devices or resetting passwords |
|    PasswordReset     |    Used to render notifications sent when password has successfully been reset    |

For example, to modify the `IdentityVerification` HTML template, if your
[template_path](../../configuration/notifications/introduction.md#template_path) was configured as
`/config/email_templates`, you would create the `/config/email_templates/IdentityVerification.html` file to override the
HTML `IdentityVerification` template.

## Placeholder Variables

In template files, you can use the following placeholders which are automatically injected into the templates:

|     Placeholder      |      Templates       |                                                                  Description                                                                   |
|:--------------------:|:--------------------:|:----------------------------------------------------------------------------------------------------------------------------------------------:|
|   `{{ .LinkURL }}`   | IdentityVerification |                                            The URL associated with the notification if applicable.                                             |
|  `{{ .LinkText }}`   | IdentityVerification |                                 The display value for the URL associated with the notification if applicable.                                  |
|    `{{ .Title }}`    |         All          | A predefined title for the email. <br> It will be `"Reset your password"` or `"Password changed successfully"`, depending on the current step. |
| `{{ .DisplayName }}` |         All          |                                                     The name of the user, i.e. `John Doe`                                                      |
|  `{{ .RemoteIP }}`   |         All          |                                      The remote IP address (client) that initiated the request or event.                                       |

## Examples

This is a basic example:

```html
<body>
  <h1>{{ .Title }}</h1>
  Hi {{ .DisplayName }}<br/>
  This email has been sent to you in order to validate your identity.
  Click <a href="{{ .LinkURL }}">here</a> to change your password.
</body>
```

Some Additional examples for specific purposes can be found in the
[examples directory on GitHub](https://github.com/authelia/authelia/tree/master/examples/templates/notifications).

## Envelope Template

There is also a special envelope template. This is the email envelope which contains the content of the other templates.
It's strongly recommended that you do not modify this template unless you know what you're doing. If you really want to
modify it the name of the file must be `Envelope.tmpl`.

This template contains the following placeholders:

In template files, you can use the following placeholders which are automatically injected into the templates:

|       Placeholder       |                                 Description                                 |
|:-----------------------:|:---------------------------------------------------------------------------:|
|      `{{ .UUID }}`      | A string representation of a UUID v4 generated specifically for this email. |
|      `{{ .Host }}`      |                           The configured [host].                            |
|   `{{ .ServerName }}`   |                      The configured TLS [server_name].                      |
|  `{{ .SenderDomain }}`  |               The domain portion of the configured [sender].                |
|   `{{ .Identifier }}`   |                        The configured [identifier].                         |
|      `{{ .From }}`      |            The string representation of the configured [sender].            |
|       `{{ .To }}`       |         The string representation of the recipients email address.          |
|    `{{ .Subject }}`     |                             The email subject.                              |
|      `{{ .Date }}`      |             The time.Time of the email envelope being rendered.             |
|    `{{ .Boundary }}`    |       The random alphanumeric 20 character multi-part email boundary.       |
| `{{ .Body.PlainText }}` |                    The plain text version of the email.                     |
|   `{{ .Body.HTML }}`    |                       The HTML version of the email.                        |

## Original Templates

The original template content can be found on
[GitHub](https://github.com/authelia/authelia/tree/master/internal/templates/src/notification).

[host]: ../../configuration/notifications/smtp.md#host
[server_name]: ../../configuration/notifications/smtp.md#tls
[sender]: ../../configuration/notifications/smtp.md#sender
[identifier]: ../../configuration/notifications/smtp.md#identifier
