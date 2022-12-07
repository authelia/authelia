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
[template_path](../../configuration/notifications/introduction.md#templatepath) notification configuration option.

## Important Notes

1. The templates are not covered by our stability guarantees. While we aim to avoid changes to the templates which
   would cause users to have to manually change them changes may be necessary in order to facilitate bug fixes or
   generally improve the templates.
   1. This is especially important for the [Envelope Template](#envelope-template).
   2. It is your responsibility to ensure your templates are up to date. We make no efforts in facilitating this.
2. We may not be able to offer any direct support in debugging these templates. We only offer support and fixes to
   the official templates.
3. All templates __*MUST*__ be encoded in UTF-8 with CRLF line endings. The line endings __*MUST NOT*__ be a simple LF.

## Template Names

|       Template       |                                    Description                                    |
|:--------------------:|:---------------------------------------------------------------------------------:|
| IdentityVerification | Used to render notifications sent when registering devices or resetting passwords |
|    PasswordReset     |    Used to render notifications sent when password has successfully been reset    |

For example, to modify the `IdentityVerification` HTML template, if your
[template_path](../../configuration/notifications/introduction.md#templatepath) was configured as
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

*__Important Note:__ This template must end with a CRLF newline. Failure to include this newline will result in
malformed emails.*

There is also a special envelope template. This is the email envelope which contains the content of the other templates
when sent via the SMTP notifier. It's *__strongly recommended__* that you do not modify this template unless you know
what you're doing. If you really want to modify it the name of the file must be `Envelope.tmpl`.

This template contains the following placeholders which are automatically injected into the template:

|       Placeholder       |                                 Description                                 |
|:-----------------------:|:---------------------------------------------------------------------------:|
|   `{{ .ProcessID }}`    |                          The Authelia Process ID.                           |
|      `{{ .UUID }}`      | A string representation of a UUID v4 generated specifically for this email. |
|      `{{ .Host }}`      |                           The configured [host].                            |
|   `{{ .ServerName }}`   |                      The configured TLS [server_name].                      |
|  `{{ .SenderDomain }}`  |               The domain portion of the configured [sender].                |
|   `{{ .Identifier }}`   |                        The configured [identifier].                         |
|      `{{ .From }}`      |            The string representation of the configured [sender].            |
|       `{{ .To }}`       |         The string representation of the recipients email address.          |
|    `{{ .Subject }}`     |                             The email subject.                              |
|      `{{ .Date }}`      |             The time.Time of the email envelope being rendered.             |

## Original Templates

The original template content can be found on
[GitHub](https://github.com/authelia/authelia/tree/master/internal/templates/src/notification).

[host]: ../../configuration/notifications/smtp.md#host
[server_name]: ../../configuration/notifications/smtp.md#tls
[sender]: ../../configuration/notifications/smtp.md#sender
[identifier]: ../../configuration/notifications/smtp.md#identifier
