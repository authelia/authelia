---
title: "SMTP"
description: "Configuring the SMTP Notifications Settings."
lead: "Authelia can send emails to users through an SMTP server. This section describes how to configure this."
date: 2020-02-29T01:43:59+01:00
draft: false
images: []
menu:
  configuration:
    parent: "notifications"
weight: 107200
toc: true
aliases:
  - /docs/configuration/notifier/smtp.html
---


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

{{< confkey type="integer" required="yes" >}}

The hostname of the SMTP server.

If utilising an IPv6 literal address it must be enclosed by square brackets and quoted:

```yaml
host: "[fd00:1111:2222:3333::1]"
```

### port

{{< confkey type="integer" required="yes" >}}

The port the SMTP service is listening on.

### timeout

{{< confkey type="duration" default="5s" required="no" >}}

The SMTP connection timeout.

### username

{{< confkey type="string" required="no" >}}

The username sent for authentication with the SMTP server. Paired with the password.

### password

{{< confkey type="string" required="no" >}}

*__Important Note:__ This can also be defined using a [secret](../methods/secrets.md) which is __strongly recommended__
especially for containerized deployments.*

The password paired with the [username](#username) sent for authentication with the SMTP server.

It's __strongly recommended__ this is a
[Random Alphanumeric String](../miscellaneous/guides.md#generating-a-random-alphanumeric-string) with 64 or more
characters and the user password is changed to this value.

### sender

{{< confkey type="string" required="yes" >}}

The sender is used to construct both the SMTP command `MAIL FROM` and to add the `FROM` header. This address must be
in [RFC5322](https://www.rfc-editor.org/rfc/rfc5322.html#section-3.4) format. This means it must one of two formats:

* jsmith@domain.com
* John Smith <jsmith@domain.com>

The `MAIL FROM` command sent to SMTP servers will not include the name portion, this is only set in the `FROM` as per
specifications.

### identifier

{{< confkey type="string" default="localhost" required="no" >}}

The name to send to the SMTP server as the identifier with the HELO/EHLO command. Some SMTP providers like Google Mail
reject the message if it's localhost.

### subject

{{< confkey type="string" default="[Authelia] {title}" required="no" >}}

This is the subject Authelia will use in the email, it has a single placeholder at present `{title}` which should
be included in all emails as it is the internal descriptor for the contents of the email.

### startup_check_address

{{< confkey type="string" default="test@authelia.com" required="no" >}}

__Authelia__ checks the SMTP server is valid at startup, one of the checks requires we ask the SMTP server if it can
send an email from us to a specific address, this is that address. No email is actually sent in the process. It is fine
to leave this as is, but you can customize it if you have issues or you desire to.

### disable_require_tls

{{< confkey type="boolean" default="false" required="no" >}}

For security reasons the default settings for Authelia require the SMTP connection is encrypted by TLS. See [security]
for more information. This option disables this measure (not recommended).

### disable_html_emails

{{< confkey type="boolean" default="false" required="no" >}}

This setting completely disables HTML formatting of emails and only sends text emails. __Authelia__ by default sends
mixed emails which contain both HTML and text so this option is rarely necessary.

### tls

Controls the TLS connection validation process. You can see how to configure the tls section
[here](../prologue/common.md#tls-configuration).

## Using Gmail

You need to generate an app password in order to use Gmail SMTP servers. The process is described
[here](https://support.google.com/accounts/answer/185833?hl=en).

```yaml
notifier:
  smtp:
    username: myaccount@gmail.com
    # Password can also be set using a secret: https://www.authelia.com/configuration/methods/secrets/
    password: yourapppassword
    sender: admin@example.com
    host: smtp.gmail.com
    port: 587
```
