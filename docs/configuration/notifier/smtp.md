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
notifier:
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
    # This secret can also be set using the env variables AUTHELIA_NOTIFIER_SMTP_PASSWORD
    password: password
    host: 127.0.0.1
    port: 1025
    sender: admin@example.com
    ## disable_require_tls: false
    ## disable_verify_cert: false
    ## trusted_cert: ""
```

## Using Gmail

You need to generate an app password in order to use Gmail SMTP servers. The process is
described [here](https://support.google.com/accounts/answer/185833?hl=en)

```yaml
notifier:
    smtp:
        username: myaccount@gmail.com
        # This secret can also be set using the env variables AUTHELIA_NOTIFIER_SMTP_PASSWORD
        password: yourapppassword
        sender: admin@example.com
        host: smtp.gmail.com
        port: 587
```