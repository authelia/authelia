---
title: "SMTP"
description: "Configuring the SMTP Notifications Settings."
summary: "Authelia can send emails to users through an SMTP server. This section describes how to configure this."
date: 2020-02-29T01:43:59+01:00
draft: false
images: []
weight: 108200
toc: true
aliases:
  - /docs/configuration/notifier/smtp.html
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---


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
    sender: "Authelia <admin@example.com>"
    identifier: 'localhost'
    subject: "[Authelia] {title}"
    startup_check_address: 'test@authelia.com'
    disable_require_tls: false
    disable_starttls: false
    disable_html_emails: false
    tls:
      server_name: 'smtp.example.com'
      skip_verify: false
      minimum_version: 'TLS1.2'
      maximum_version: 'TLS1.3'
      certificate_chain: |
        -----BEGIN CERTIFICATE-----
        MIIC5jCCAc6gAwIBAgIRAK4Sj7FiN6PXo/urPfO4E7owDQYJKoZIhvcNAQELBQAw
        EzERMA8GA1UEChMIQXV0aGVsaWEwHhcNNzAwMTAxMDAwMDAwWhcNNzEwMTAxMDAw
        MDAwWjATMREwDwYDVQQKEwhBdXRoZWxpYTCCASIwDQYJKoZIhvcNAQEBBQADggEP
        ADCCAQoCggEBAPKv3pSyP4ozGEiVLJ14dIWFCEGEgq7WUMI0SZZqQA2ID0L59U/Q
        /Usyy7uC9gfMUzODTpANtkOjFQcQAsxlR1FOjVBrX5QgjSvXwbQn3DtwMA7XWSl6
        LuYx2rBYSlMSN5UZQm/RxMtXfLK2b51WgEEYDFi+nECSqKzR4R54eOPkBEWRfvuY
        91AMjlhpivg8e4JWkq4LVQUKbmiFYwIdK8XQiN4blY9WwXwJFYs5sQ/UYMwBFi0H
        kWOh7GEjfxgoUOPauIueZSMSlQp7zqAH39N0ZSYb6cS0Npj57QoWZSY3ak87ebcR
        Nf4rCvZLby7LoN7qYCKxmCaDD3x2+NYpWH8CAwEAAaM1MDMwDgYDVR0PAQH/BAQD
        AgWgMBMGA1UdJQQMMAoGCCsGAQUFBwMBMAwGA1UdEwEB/wQCMAAwDQYJKoZIhvcN
        AQELBQADggEBAHSITqIQSNzonFl3DzxHPEzr2hp6peo45buAAtu8FZHoA+U7Icfh
        /ZXjPg7Xz+hgFwM/DTNGXkMWacQA/PaNWvZspgRJf2AXvNbMSs2UQODr7Tbv+Fb4
        lyblmMUNYFMCFVAMU0eIxXAFq2qcwv8UMcQFT0Z/35s6PVOakYnAGGQjTfp5Ljuq
        wsdc/xWmM0cHWube6sdRRUD7SY20KU/kWzl8iFO0VbSSrDf1AlEhnLEkp1SPaxXg
        OdBnl98MeoramNiJ7NT6Jnyb3zZ578fjaWfThiBpagItI8GZmG4s4Ovh2JbheN8i
        ZsjNr9jqHTjhyLVbDRlmJzcqoj4JhbKs6/I^invalid DO NOT USE=
        -----END CERTIFICATE-----
        -----BEGIN CERTIFICATE-----
        MIIDBDCCAeygAwIBAgIRALJsPg21kA0zY4F1wUCIuoMwDQYJKoZIhvcNAQELBQAw
        EzERMA8GA1UEChMIQXV0aGVsaWEwHhcNNzAwMTAxMDAwMDAwWhcNNzEwMTAxMDAw
        MDAwWjATMREwDwYDVQQKEwhBdXRoZWxpYTCCASIwDQYJKoZIhvcNAQEBBQADggEP
        ADCCAQoCggEBAMXHBvVxUzYk0u34/DINMSF+uiOekKOAjOrC6Mi9Ww8ytPVO7t2S
        zfTvM+XnEJqkFQFgimERfG/eGhjF9XIEY6LtnXe8ATvOK4nTwdufzBaoeQu3Gd50
        5VXr6OHRo//ErrGvFXwP3g8xLePABsi/fkH3oDN+ztewOBMDzpd+KgTrk8ysv2ou
        kNRMKFZZqASvCgv0LD5KWvUCnL6wgf1oTXG7aztduA4oSkUP321GpOmBC5+5ElU7
        ysoRzvD12o9QJ/IfEaulIX06w9yVMo60C/h6A3U6GdkT1SiyTIqR7v7KU/IWd/Qi
        Lfftcj91VhCmJ73Meff2e2S2PrpjdXbG5FMCAwEAAaNTMFEwDgYDVR0PAQH/BAQD
        AgKkMA8GA1UdJQQIMAYGBFUdJQAwDwYDVR0TAQH/BAUwAwEB/zAdBgNVHQ4EFgQU
        Z7AtA3mzFc0InSBA5fiMfeLXA3owDQYJKoZIhvcNAQELBQADggEBAEE5hm1mtlk/
        kviCoHH4evbpw7rxPxDftIQlqYTtvMM4eWY/6icFoSZ4fUHEWYyps8SsPu/8f2tf
        71LGgZn0FdHi1QU2H8m0HHK7TFw+5Q6RLrLdSyk0PItJ71s9en7r8pX820nAFEHZ
        HkOSfJZ7B5hFgUDkMtVM6bardXAhoqcMk4YCU96e9d4PB4eI+xGc+mNuYvov3RbB
        D0s8ICyojeyPVLerz4wHjZu68Z5frAzhZ68YbzNs8j2fIBKKHkHyLG1iQyF+LJVj
        2PjCP+auJsj6fQQpMGoyGtpLcSDh+ptcTngUD8JsWipzTCjmaNqdPHAOYmcgtf4b
        qocikt3WAdU^invalid DO NOT USE=
        -----END CERTIFICATE-----
      private_key: |
        -----BEGIN RSA PRIVATE KEY-----
        MIIEpAIBAAKCAQEA8q/elLI/ijMYSJUsnXh0hYUIQYSCrtZQwjRJlmpADYgPQvn1
        T9D9SzLLu4L2B8xTM4NOkA22Q6MVBxACzGVHUU6NUGtflCCNK9fBtCfcO3AwDtdZ
        KXou5jHasFhKUxI3lRlCb9HEy1d8srZvnVaAQRgMWL6cQJKorNHhHnh44+QERZF+
        +5j3UAyOWGmK+Dx7glaSrgtVBQpuaIVjAh0rxdCI3huVj1bBfAkVizmxD9RgzAEW
        LQeRY6HsYSN/GChQ49q4i55lIxKVCnvOoAff03RlJhvpxLQ2mPntChZlJjdqTzt5
        txE1/isK9ktvLsug3upgIrGYJoMPfHb41ilYfwIDAQABAoIBAQDTOdFf2JjHH1um
        aPgRAvNf9v7Nj5jytaRKs5nM6iNf46ls4QPreXnMhqSeSwj6lpNgBYxOgzC9Q+cc
        Y4ob/paJJPaIJTxmP8K/gyWcOQlNToL1l+eJ20eQoZm23NGr5fIsunSBwLEpTrdB
        ENqqtcwhW937K8Pxy/Q1nuLyU2bc6Tn/ivLozc8n27dpQWWKh8537VY7ancIaACr
        LJJLYxKqhQpjtBWAyCDvZQirnAOm9KnvIHaGXIswCZ4Xbsu0Y9NL+woARPyRVQvG
        jfxy4EmO9s1s6y7OObSukwKDSNihAKHx/VIbvVWx8g2Lv5fGOa+J2Y7o9Qurs8t5
        BQwMTt0BAoGBAPUw5Z32EszNepAeV3E2mPFUc5CLiqAxagZJuNDO2pKtyN29ETTR
        Ma4O1cWtGb6RqcNNN/Iukfkdk27Q5nC9VJSUUPYelOLc1WYOoUf6oKRzE72dkMQV
        R4bf6TkjD+OVR17fAfkswkGahZ5XA7j48KIQ+YC4jbnYKSxZTYyKPjH/AoGBAP1i
        tqXt36OVlP+y84wWqZSjMelBIVa9phDVGJmmhz3i1cMni8eLpJzWecA3pfnG6Tm9
        ze5M4whASleEt+M00gEvNaU9ND+z0wBfi+/DwJYIbv8PQdGrBiZFrPhTPjGQUldR
        lXccV2meeLZv7TagVxSi3DO6dSJfSEHyemd5j9mBAoGAX8Hv+0gOQZQCSOTAq8Nx
        6dZcp9gHlNaXnMsP9eTDckOSzh636JPGvj6m+GPJSSbkURUIQ3oyokMNwFqvlNos
        fTaLhAOfjBZI9WnDTTQxpugWjphJ4HqbC67JC/qIiw5S6FdaEvGLEEoD4zoChywZ
        9oGAn+fz2d/0/JAH/FpFPgsCgYEAp/ipZgPzziiZ9ov1wbdAQcWRj7RaWnssPFpX
        jXwEiXT3CgEMO4MJ4+KWIWOChrti3qFBg6i6lDyyS6Qyls7sLFbUdC7HlTcrOEMe
        rBoTcCI1GqZNlqWOVQ65ZIEiaI7o1vPBZo2GMQEZuq8mDKFsOMThvvTrM5cAep84
        n6HJR4ECgYABWcbsSnr0MKvVth/inxjbKapbZnp2HUCuw87Ie5zK2Of/tbC20wwk
        yKw3vrGoE3O1t1g2m2tn8UGGASeZ842jZWjIODdSi5+icysQGuULKt86h/woz2SQ
        27GoE2i5mh6Yez6VAYbUuns3FcwIsMyWLq043Tu2DNkx9ijOOAuQzw^invalid..
        DO NOT USE==
        -----END RSA PRIVATE KEY-----
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

```yaml
notifier:
  smtp:
    address: 'smtp://127.0.0.1:25'
```

```yaml
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

{{< confkey type="string" required="no" >}}

*__Important Note:__ This can also be defined using a [secret](../methods/secrets.md) which is __strongly recommended__
especially for containerized deployments.*

The password paired with the [username](#username) sent for authentication with the SMTP server.

It's __strongly recommended__ this is a
[Random Alphanumeric String](../../reference/guides/generating-secure-values.md#generating-a-random-alphanumeric-string) with 64 or more
characters and the user password is changed to this value.

### sender

{{< confkey type="string" required="yes" >}}

The sender is used to construct both the SMTP command `MAIL FROM` and to add the `FROM` header. This address must be
in [RFC5322](https://datatracker.ietf.org/doc/html/rfc5322#section-3.4) format. This means it must one of two formats:

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

Controls the TLS connection validation parameters for either StartTLS or the TLS socket.

## Using Gmail

You need to generate an app password in order to use Gmail SMTP servers. The process is described
[here](https://support.google.com/accounts/answer/185833?hl=en).

```yaml
notifier:
  smtp:
    username: 'myaccount@gmail.com'
    # Password can also be set using a secret: https://www.authelia.com/configuration/methods/secrets/
    password: 'yourapppassword'
    sender: 'admin@example.com'
    host: 'smtp.gmail.com'
    port: 587
```
