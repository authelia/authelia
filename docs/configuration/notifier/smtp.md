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

```yaml
# Configuration of the notification system.
#
# Notifications are sent to users when they require a password reset, a u2f
# registration or a TOTP registration.
# Use only an available configuration: filesystem, smtp.
notifier:
  # You can disable the notifier startup check by setting this to true.
  disable_startup_check: false

  # For testing purpose, notifications can be sent in a file
  ## filesystem:
  ##   filename: /config/notification.txt

  # Use a SMTP server for sending notifications. Authelia uses PLAIN or LOGIN method to authenticate.
  # [Security] By default Authelia will:
  #   - force all SMTP connections over TLS including unauthenticated connections
  #      - use the disable_require_tls boolean value to disable this requirement (only works for unauthenticated connections)
  #   - validate the SMTP server x509 certificate during the TLS handshake against the hosts trusted certificates
  #     - trusted_cert option:
  #       - this is a string value, that may specify the path of a PEM format cert, it is completely optional
  #       - if it is not set, a blank string, or an invalid path; will still trust the host machine/containers cert store
  #     - defaults to the host machine (or docker container's) trusted certificate chain for validation
  #     - use the trusted_cert string value to specify the path of a PEM format public cert to trust in addition to the hosts trusted certificates
  #     - use the disable_verify_cert boolean value to disable the validation (prefer the trusted_cert option as it's more secure)
  smtp:
    username: test
    # Password can also be set using a secret: https://docs.authelia.com/configuration/secrets.html
    password: password
    host: 127.0.0.1
    port: 1025
    sender: admin@example.com
    # Subject configuration of the emails sent.
    # {title} is replaced by the text from the notifier
    subject: "[Authelia] {title}"
    # This address is used during the startup check to verify the email configuration is correct. It's not important what it is except if your email server only allows local delivery.
    startup_check_address: test@authelia.com
    trusted_cert: ""
    disable_require_tls: false
    disable_verify_cert: false
    disable_html_emails: false
```

## Configuration options

Most configuration options are self-explanatory, however here is an explanation of the ones that may not
be as obvious.

### subject
This is the subject Authelia will use in the email, it has a single placeholder at present `{title}` which should
be included in all emails as it is the internal descriptor for the contents of the email.

### disable_require_tls
For security reasons the default settings for Authelia require the SMTP connection is encrypted by TLS. See [security] for
more information. This option disables this measure (not recommended).

###  disable_verify_cert
For security reasons Authelia only trusts certificates valid according to the OS's PKI chain. See [security] for more information.
This option disables this measure (not recommended).

### disable_html_emails
This option forces Authelia to only send plain text email via the notifier. This is the default for the file based 
notifier, but some users may wish to use plain text for security reasons.

### trusted_cert
This option allows you to specify the file path to a public key portion of a X509 certificate in order to trust it, or 
certificates signed with the private key portion of the X509 certificate. This is an alternative to `disable_verify_cert`
that is much more secure. This is not required if your certificate is trusted by the operating system PKI. 

## Using Gmail

You need to generate an app password in order to use Gmail SMTP servers. The process is
described [here](https://support.google.com/accounts/answer/185833?hl=en)

```yaml
notifier:
  smtp:
    username: myaccount@gmail.com
    # Password can also be set using a secret: https://docs.authelia.com/configuration/secrets.html
    password: yourapppassword
    sender: admin@example.com
    host: smtp.gmail.com
    port: 587
```

## Loading a password from a secret instead of inside the configuration

Password can also be defined using a [secret](../secrets.md).

[security]: ../../security/measures.md#notifier-security-measures-smtp