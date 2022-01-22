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

This option allows the administrator to set custom templates for notifications
the templates folder should contain the following files

|File                    |Description                                        |
|------------------------|---------------------------------------------------|
|PasswordResetStep1.html |HTML Template for Step 1 of password reset process |
|PasswordResetStep1.txt  |Text Template for Step 1 of password reset process |
|PasswordResetStep2.html |HTML Template for Step 2 of password reset process |
|PasswordResetStep2.txt  |Text Template for Step 2 of password reset process |

Note:
* if you don't define some of these files, a default template is used for that notification


In template files, you can use the following variables:

|File                    |Description                                        |
|------------------------|---------------------------------------------------|
|`{{.title}}`| A predefined title for the email. <br> It will be `"Reset your password"` or `"Password changed successfully"`, depending on the current step |
|`{{.url}}`  | The url that allows to reset the user password |
|`{{.displayName}}` |The name of the user, i.e. `John Doe` |
|`{{.button}}` |The content for the password reset button, it's hardcoded to `Reset` |

#### Example

```html
<body>
  <h1>{{.title}}</h1>
  Hi {{.displayName}} <br/>
  This email has been sent to you in order to validate your identity
  Click <a href="{{.url}}" >here</a> to change your password
</body>
```


### filesystem

The [filesystem](filesystem.md) provider.

### smtp

The [smtp](smtp.md) provider.
