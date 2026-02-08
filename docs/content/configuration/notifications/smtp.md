---
title: "SMTP"
description: "Configuring the SMTP Notifications Settings."
summary: "Authelia can send emails to users through an SMTP server. This section describes how to configure this."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 108200
toc: true
aliases:
  - '/docs/configuration/notifier/smtp.html'
  - '/configuration/authentication/smtp/'
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Variables

Some of the values within this page can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

{{< config-alert-example >}}

```yaml {title="configuration.yml"}
notifier:
  disable_startup_check: false
  smtp:
    address: 'smtp://127.0.0.1:25'
    timeout: '5s'
    username: 'test'
    password: 'password'
    sender: "Authelia <admin@{{< sitevar name="domain" nojs="example.com" >}}>"
    identifier: 'localhost'
    subject: "[Authelia] {title}"
    startup_check_address: 'test@{{< sitevar name="domain" nojs="example.com" >}}'
    disable_require_tls: false
    disable_starttls: false
    disable_html_emails: false
    tls:
      server_name: 'smtp.{{< sitevar name="domain" nojs="example.com" >}}'
      skip_verify: false
      minimum_version: 'TLS1.2'
      maximum_version: 'TLS1.3'
      certificate_chain: |
        -----BEGIN CERTIFICATE-----
        ...
        -----END CERTIFICATE-----
        -----BEGIN CERTIFICATE-----
        ...
        -----END CERTIFICATE-----
      private_key: |
        -----BEGIN PRIVATE KEY-----
        ...
        -----END PRIVATE KEY-----
```

## Options

This section describes the individual configuration options.

### address

{{< confkey type="string" syntax="address" required="yes" >}}

Configures the address for the SMTP Server. The address itself is a connector and the scheme must be `smtp`,
`submission`, or `submissions`. The only difference between these schemes are the default ports and `submissions`
requires a TLS transport per [SMTP Ports Security Measures][docs-security-smtp-port], whereas `submission` and `smtp`
use a standard TCP transport and typically enforce StartTLS.

[docs-security-smtp-port]: ../../overview/security/measures.md#smtp-ports

__Examples:__

```yaml {title="configuration.yml"}
notifier:
  smtp:
    address: 'smtp://127.0.0.1:25'
```

```yaml {title="configuration.yml"}
notifier:
  smtp:
    address: 'submissions://[fd00:1111:2222:3333::1]:465'
```

### timeout

{{< confkey type="string,integer" syntax="duration" default="5 seconds" required="no" >}}

The SMTP connection timeout.

### username

{{< confkey type="string" required="no" >}}

The username sent for authentication with the SMTP server. Paired with the password.

### password

{{< confkey type="string" required="no" secret="yes" >}}

The password paired with the [username](#username) sent for authentication with the SMTP server.

It's __strongly recommended__ this is a
[Random Alphanumeric String](../../reference/guides/generating-secure-values.md#generating-a-random-alphanumeric-string) with 64 or more
characters and the user password is changed to this value.

### sender

{{< confkey type="string" required="yes" >}}

The sender is used to construct both the SMTP command `MAIL FROM` and to add the `FROM` header. This address must be
in [RFC5322](https://datatracker.ietf.org/doc/html/rfc5322#section-3.4) format. This means it must one of two formats:

* `jsmith@domain.com`
* `John Smith <jsmith@domain.com>`

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

{{< callout context="danger" title="Security Note" icon="outline/alert-octagon" >}}
Enabling this value will result in all emails being sent in the clear and will leak values which have a critical
security impact. This is almost certainly an indication attackers would be able to easily perform a phishing attack
using your SMTP server and Authelia SMTP credentials.
{{< /callout >}}

For security reasons the default settings for Authelia require the SMTP connection is encrypted by TLS. See [security]
for more information. This option disables this measure but this is highly discouraged and not a supported
configuration.

### disable_starttls

{{< confkey type="boolean" default="false" required="no" >}}

Some SMTP servers ignore SMTP specifications and claim to support STARTTLS when they in fact do not.
For security reasons Authelia refuses to send messages to these servers.
This option disables this measure and is enabled  *__AT YOUR OWN RISK__*. It's *__strongly recommended__*
that instead of enabling this option you either fix the issue with the SMTP server's configuration or
have the administrators of the server fix it. If the issue can't be fixed via the SMTP server configuration we recommend
lodging an issue with the authors of the SMTP server.

See [security] for more information.

### disable_html_emails

{{< confkey type="boolean" default="false" required="no" >}}

This setting completely disables HTML formatting of emails and only sends text emails. __Authelia__ by default sends
mixed emails which contain both HTML and text so this option is rarely necessary.

### tls

{{< confkey type="structure" structure="tls" required="no" >}}

If defined this option controls the TLS connection verification parameters for the SMTP server.

By default Authelia uses the system certificate trust for TLS certificate verification of TLS connections and the
[certificates_directory](../miscellaneous/introduction.md#certificates_directory) global option can be used to augment
this.

## Using Gmail

You need to generate an app password in order to use Gmail SMTP servers. The process is described
[here](https://support.google.com/accounts/answer/185833?hl=en).

```yaml {title="configuration.yml"}
notifier:
  smtp:
    address: 'submission://smtp.gmail.com:587'
    username: 'myaccount@gmail.com'
    # Password can also be set using a secret: https://www.authelia.com/configuration/methods/secrets/
    password: 'yourapppassword'
    sender: 'admin@{{< sitevar name="domain" nojs="example.com" >}}'
```

[security]: ../../overview/security/measures.md#notifier-security-measures-smtp
