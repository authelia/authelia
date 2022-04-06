---
layout: default
title: Notifier
parent: Configuration
nav_order: 8
has_children: true
---

# Notifier

**Authelia** sometimes needs to send messages to users in order to
verify their identity.

## Configuration

```yaml
notifier:
  disable_startup_check: false
  template_path: /path/to/templates/folder
  filesystem: {}
  smtp: {}
```

## Options

### disable_startup_check
<div markdown="1">
type: boolean
{: .label .label-config .label-purple }
default: false
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

The notifier has a startup check which validates the specified provider
configuration is correct and will be able to send emails. This can be
disabled with the `disable_startup_check` option:

### template_path
<div markdown="1">
type: string
{: .label .label-config .label-purple }
default: ""
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

This option allows the administrator to set a path where custom templates for notifications can be found. Each template
has two extensions; `.html` for HTML templates, and `.txt` for plaintext templates.

|       Template       |                          Description                          |
|:--------------------:|:-------------------------------------------------------------:|
|        Basic         |      Template used to send basic notifications to users       |
| IdentityVerification | Template used when registering devices or resetting passwords |

For example, to modify the `IdentityVerification` HTML template, if your `template_path` was `/config/email_templates`,
you would create the `/config/email_templates/IdentityVerification.html` file.

_**Note:** you may configure this directory and add only add the templates you wish to override, any templates not
overriden will utilize the default templates._ 


In template files, you can use the following variables:

|    Placeholder     |      Templates       |                                                                  Description                                                                   |
|:------------------:|:--------------------:|:----------------------------------------------------------------------------------------------------------------------------------------------:|
|     `{{.url}}`     | IdentityVerification |                                          The URL of the used with the IdentityVerification template.                                           |
|    `{{.title}}`    |         All          | A predefined title for the email. <br> It will be `"Reset your password"` or `"Password changed successfully"`, depending on the current step  |
| `{{.displayName}}` |         All          |                                                     The name of the user, i.e. `John Doe`                                                      |
|   `{{.button}}`    |         All          |                                      The content for the password reset button, it's hardcoded to `Reset`                                      |
|  `{{.remoteIP}}`   |         All          |                                           The remote IP address that initiated the request or event                                            |

#### Examples

This is a basic example:

```html
<body>
  <h1>{{.title}}</h1>
  Hi {{.displayName}} <br/>
  This email has been sent to you in order to validate your identity
  Click <a href="{{.url}}" >here</a> to change your password
</body>
```

Some Additional examples for specific purposes can be found in the 
[examples directory on GitHub](https://github.com/authelia/authelia/tree/master/examples/templates/notifications).

### filesystem

The [filesystem](filesystem.md) provider.

### smtp

The [smtp](smtp.md) provider.
