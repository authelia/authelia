---
layout: default
title: SMTP
parent: Notifier
grand_parent: Configuration
nav_order: 2
---

# SMTP
**Authelia** can send emails to users through an SMTP server.
It can be configured as described below.

## Configuration

```yaml
notifier:
  disable_startup_check: false
  smtp:
    host: 127.0.0.1
    port: 1025
    timeout: 5s
    username: test
    password: password
    sender: "Authelia <admin@example.com>"
    identifier: localhost
    subject: "[Authelia] {title}"
    startup_check_address: test@authelia.com
    disable_require_tls: false
    disable_html_emails: false
    tls:
      server_name: smtp.example.com
      skip_verify: false
      minimum_version: TLS1.2
```

## Options

### host
<div markdown="1">
type: integer
{: .label .label-config .label-purple } 
required: yes
{: .label .label-config .label-red }
</div>

The hostname of the SMTP server.

If utilising an IPv6 literal address it must be enclosed by square brackets and quoted:

```yaml
host: "[fd00:1111:2222:3333::1]"
```

### port

<div markdown="1">
type: integer
{: .label .label-config .label-purple } 
required: yes
{: .label .label-config .label-red }
</div>

The port the SMTP service is listening on.

### timeout
<div markdown="1">
type: duration
{: .label .label-config .label-purple } 
default: 5s
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

The SMTP connection timeout.

### username
<div markdown="1">
type: string
{: .label .label-config .label-purple }
required: no
{: .label .label-config .label-green }
</div>

The username sent for authentication with the SMTP server. Paired with the password.

### password
<div markdown="1">
type: string
{: .label .label-config .label-purple } 
required: no
{: .label .label-config .label-green }
</div>

The password sent for authentication with the SMTP server. Paired with the username. Can also be defined using a
[secret](../secrets.md) which is the recommended for containerized deployments.

### sender
<div markdown="1">
type: string
{: .label .label-config .label-purple } 
required: yes
{: .label .label-config .label-red }
</div>

The sender is used to construct both the SMTP command `MAIL FROM` and to add the `FROM` header. This address must be
in [RFC5322](https://datatracker.ietf.org/doc/html/rfc5322#section-3.4) format. This means it must one of two formats:
- jsmith@domain.com
- John Smith <jsmith@domain.com>

The `MAIL FROM` command sent to SMTP servers will not include the name portion, this is only set in the `FROM` as per
specifications.

### identifier
<div markdown="1">
type: string
{: .label .label-config .label-purple } 
default: localhost
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

The name to send to the SMTP server as the identifier with the HELO/EHLO command. Some SMTP providers like Google Mail
reject the message if it's localhost.

### subject
<div markdown="1">
type: string
{: .label .label-config .label-purple } 
default: [Authelia] {title}
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

This is the subject Authelia will use in the email, it has a single placeholder at present `{title}` which should
be included in all emails as it is the internal descriptor for the contents of the email.

### startup_check_address
<div markdown="1">
type: string
{: .label .label-config .label-purple } 
default: test@authelia.com
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

**Authelia** checks the SMTP server is valid at startup, one of the checks requires we ask the SMTP server if it can
send an email from us to a specific address, this is that address. No email is actually sent in the process. It is fine
to leave this as is, but you can customize it if you have issues or you desire to.

### disable_require_tls
<div markdown="1">
type: boolean
{: .label .label-config .label-purple } 
default: false
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

For security reasons the default settings for Authelia require the SMTP connection is encrypted by TLS. See [security]
for more information. This option disables this measure (not recommended).

### disable_html_emails
<div markdown="1">
type: boolean
{: .label .label-config .label-purple } 
default: false
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

This setting completely disables HTML formatting of emails and only sends text emails. **Authelia** by default sends
mixed emails which contain both HTML and text so this option is rarely necessary.

### tls

Controls the TLS connection validation process. You can see how to configure the tls section
[here](../index.md#tls-configuration).


## Using Gmail
You need to generate an app password in order to use Gmail SMTP servers. The process is
described [here](https://support.google.com/accounts/answer/185833?hl=en)

```yaml
notifier:
  smtp:
    username: myaccount@gmail.com
    # Password can also be set using a secret: https://www.authelia.com/docs/configuration/secrets.html
    password: yourapppassword
    sender: admin@example.com
    host: smtp.gmail.com
    port: 587
```
