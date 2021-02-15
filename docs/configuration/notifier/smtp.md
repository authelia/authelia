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
    username: test
    # Password can also be set using a secret: https://www.authelia.com/docs/configuration/secrets.html
    password: password
    host: 127.0.0.1
    port: 1025
    sender: admin@example.com
    # HELO/EHLO Identifier. Some SMTP Servers may reject the default of localhost.
    identifier: localhost
    # Subject configuration of the emails sent.
    # {title} is replaced by the text from the notifier
    subject: "[Authelia] {title}"
    # This address is used during the startup check to verify the email configuration is correct. It's not important what it is except if your email server only allows local delivery.
    startup_check_address: test@authelia.com
    disable_require_tls: false
    disable_html_emails: false

    tls:
      # Server Name for certificate validation (in case you are using the IP or non-FQDN in the host option).
      # server_name: smtp.example.com

      # Skip verifying the server certificate (to allow a self-signed certificate).
      skip_verify: false

      # Minimum TLS version for either StartTLS or SMTPS.
      minimum_version: TLS1.2

  # Sending an email using a Gmail account is as simple as the next section.
  # You need to create an app password by following: https://support.google.com/accounts/answer/185833?hl=en
  ## smtp:
  ##   username: myaccount@gmail.com
  ##   # Password can also be set using a secret: https://www.authelia.com/docs/configuration/secrets.html
  ##   password: yourapppassword
  ##   sender: admin@example.com
  ##   host: smtp.gmail.com
  ##   port: 587
```

## Options

### username

The username sent for authentication with the SMTP server. Paired with the password.

### password

The password sent for authentication with the SMTP server. Paired with the username. Can also be defined using a
[secret](../secrets.md) which is the recommended for containerized deployments.

### host

The hostname of the SMTP server.

If utilising an IPv6 literal address it must be enclosed by square brackets and quoted:

```yaml
host: "[fd00:1111:2222:3333::1]"
```

### port

The port the SMTP service is listening on.

### sender

The address sent in the FROM header for the email. Basically who the email appears to come from. It should be noted
that some SMTP servers require the username provided to have access to send from the specific address listed here.

### identifer

The name to send to the SMTP server as the identifier with the HELO/EHLO command. Some SMTP providers like Google Mail
reject the message if it's localhost.

### subject

This is the subject Authelia will use in the email, it has a single placeholder at present `{title}` which should
be included in all emails as it is the internal descriptor for the contents of the email.

### startup_check_address

**Authelia** checks the SMTP server is valid at startup, one of the checks requires we ask the SMTP server if it can
send an email from us to a specific address, this is that address. No email is actually sent in the process. It is fine
to leave this as is, but you can customize it if you have issues or you desire to.

### disable_require_tls

For security reasons the default settings for Authelia require the SMTP connection is encrypted by TLS. See [security] 
for more information. This option disables this measure (not recommended).

### disable_html_emails

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