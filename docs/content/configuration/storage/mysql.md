---
title: "MySQL"
description: "MySQL Configuration"
summary: "The MySQL storage provider which supports both MySQL and MariaDB."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
weight: 107600
toc: true
aliases:
  - /docs/configuration/storage/mariadb.html
  - /docs/configuration/storage/mysql.html
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Version support

See the [MySQL Database Integration](../../reference/integrations/database-integrations.md#mysql) reference
guide for supported version information.

## Configuration

{{< config-alert-example >}}

```yaml {title="configuration.yml"}
storage:
  encryption_key: 'a_very_important_secret'
  mysql:
    address: 'tcp://127.0.0.1:3306'
    database: 'authelia'
    username: 'authelia'
    password: 'mypassword'
    timeout: '5s'
    tls:
      server_name: 'mysql.example.com'
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

### encryption_key

See the [encryption_key docs](introduction.md#encryption_key).

### address

{{< confkey type="string" syntax="address" required="yes" >}}

Configures the address for the MySQL/MariaDB Server. The address itself is a connector and the scheme must either be
the `unix` scheme or one of the `tcp` schemes.

__Examples:__

```yaml
storage:
  mysql:
    address: 'tcp://127.0.0.1:3306'
```

```yaml
storage:
  mysql:
    address: 'tcp://[fd00:1111:2222:3333::1]:3306'
```

```yaml
storage:
  mysql:
    address: 'unix:///var/run/mysqld.sock'
```

### database

{{< confkey type="string" required="yes" >}}

The database name on the database server that the assigned [user](#username) has access to for the purpose of
__Authelia__.

### username

{{< confkey type="string" required="yes" >}}

The username paired with the password used to connect to the database.

### password

{{< confkey type="string" required="yes" >}}

*__Important Note:__ This can also be defined using a [secret](../methods/secrets.md) which is __strongly recommended__
especially for containerized deployments.*

The password paired with the [username](#username) used to connect to the database.

It's __strongly recommended__ this is a
[Random Alphanumeric String](../../reference/guides/generating-secure-values.md#generating-a-random-alphanumeric-string) with 64 or more
characters and the user password is changed to this value.

### timeout

{{< confkey type="string,integer" syntax="duration" default="5 seconds" required="no" >}}

The SQL connection timeout.

### tls

{{< confkey type="structure" structure="tls" required="no" >}}

If defined enables connecting to [MySQL] or [MariaDB] over a TLS socket, and additionally controls the TLS connection
validation parameters.

[MySQL]: https://www.mysql.com/
[MariaDB]: https://mariadb.org/
