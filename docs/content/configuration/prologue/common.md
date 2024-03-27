---
title: "Common"
description: "Common configuration options and notations."
summary: "This section details common configuration elements within the Authelia configuration. This section is mainly used as a reference for other sections as necessary."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
weight: 100200
toc: true
aliases:
  - /c/common
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Syntax

The following represent common syntax used within the configuration which have specific format requirements that are
used in multiple areas. This is intended on assisting in understanding these specific values, and not as a specific
guide on configuring any particular instance.

### Duration

The base type for this syntax is a string, and it also handles integers however this is discouraged.

If you supply an integer, it is considered a representation of seconds. If you supply a string, it parses the string in
blocks of quantities and units (number followed by a unit letter).  For example `5h` indicates a quantity of 5 units
of `h`.

The following is ignored or stripped from the input:
  - all spaces
  - leading zeros
  - the word `and`

While you can use multiple of these blocks in combination, we suggest keeping it simple and use a single value. In
addition it's important to note that the format while somewhat human readable still requires you closely follow the
expected formats.

#### Unit Legend

The following is a legend for the unit formats available in this syntax. The long form units are only available from
v4.38.0 or newer.

|     Unit     | Short Unit |   Human Readable Long Unit    |
|:------------:|:----------:|:-----------------------------:|
|    Years     |    `y`     |        `year`, `years`        |
|    Months    |    `M`     |       `month`, `months`       |
|    Weeks     |    `w`     |        `week`, `weeks`        |
|     Days     |    `d`     |         `day`, `days`         |
|    Hours     |    `h`     |        `hour`, `hours`        |
|   Minutes    |    `m`     |      `minute`, `minutes`      |
|   Seconds    |    `s`     |      `second`, `seconds`      |
| Milliseconds |    `ms`    | `millisecond`, `milliseconds` |

#### Examples

|     Desired Value     |    Configuration Examples (Short)     |     Configuration Examples (Long)      |
|:---------------------:|:-------------------------------------:|:--------------------------------------:|
| 1 hour and 30 minutes | `90m` or `1h30m` or `5400` or `5400s` |        `1 hour and 30 minutes`         |
|         1 day         | `1d` or `24h` or `86400` or `86400s`  |                `1 day`                 |
|       10 hours        | `10h` or `600m` or `9h60m` or `36000` |               `10 hours`               |

### Address

The base type for this syntax is a string.

The address type is a string that indicates how to configure a listener (i.e. listening for connections) or connector
(i.e. opening remote connections), which are the two primary categories of addresses.


#### Format

This section outlines the format for these strings. The formats use a conventional POSIX format to indicate optional and
required elements. The square brackets `[]` surround optional portions, and the angled brackets `<>` surround required
portions. Required portions may exist within optional portions, in which case they are often accompanied with other
format specific text which indicates if the accompanying text exists then it is actually required, otherwise it's
entirely optional.

The square brackets indicate optional sections, and the angled brackets indicate required sections. The following
sections elaborate on this. Sections may only be optional for the purposes of parsing, there may be a configuration
requirement that one of these is provided.

##### Hostname

The following format represents the hostname format. It's valid for both a listener and connector in most instances.
Refer to the individual documentation for an option for clarity. In this format as per the notation the scheme and port
are optional. The default for these when not provided varies.

```text
[<scheme>://]<hostname>[:<port>][/<path>]
```

##### Port

The following format represents the port format. It's valid only for a listener in most instances.
Refer to the individual documentation for an option for clarity. In this format as per the notation the scheme and
hostname are optional. The default for the scheme when not provided varies, and the default for the hostname is all
available addresses when not provided.

```text
[<scheme>://][hostname]:<port>[/<path>]
```

##### Unix Domain Socket

The following format represents the unix domain socket format. It's valid for both a listener and connector in most
instances. Refer to the individual documentation for an option for clarity. In this format as per the notation there
are no optional portions.

The Unix Domain Socket format also accepts a query string. The following query parameters control certain behavior of
this address type.

| Parameter | Listeners | Connectors |                                                                                                           Purpose                                                                                                            |
|:---------:|:---------:|:----------:|:----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------:|
|  `umask`  |    Yes    |     No     |                                             Sets the umask prior to creating the socket and restores it after creating it. The value must be an octal number with 3 or 4 digits.                                             |
|  `path`   |    Yes    |     No     | Sets the path variable to configure the subpath, specifically for a unix socket but technically works for TCP as well. Note that this should just be the alphanumeric portion it should not be prefixed with a forward slash |


```text
unix://<path>
```

```text
unix://<path>?umask=0022
```

```text
unix://<path>?path=auth
```

```text
unix://<path>?umask=0022&path=auth
```

##### Examples

Various examples for these formats.

```text
0.0.0.0
tcp://0.0.0.0
tcp://0.0.0.0/subpath
tcp://0.0.0.0:9091
tcp://0.0.0.0:9091/subpath
tcp://:9091
tcp://:9091/subpath
0.0.0.0:9091

udp://0.0.0.0:123
udp://:123

unix:///var/lib/authelia.sock
```

#### scheme

The entire scheme is optional, but if the scheme host delimiter `://` is in the string, the scheme must be present. The
scheme must be one of the following (the listeners and connectors columns indicate support for the scheme on the
respective address type):

|    Scheme     | Listeners | Connectors | Default Port |                                   Notes                                   |
|:-------------:|:---------:|:----------:|:------------:|:-------------------------------------------------------------------------:|
|     `tcp`     |    Yes    |    Yes     |     N/A      |        Standard TCP Socket which allows IPv4 and/or IPv6 addresses        |
|    `tcp4`     |    Yes    |    Yes     |     N/A      |           Standard TCP Socket which allows only IPv4 addresses            |
|    `tcp6`     |    Yes    |    Yes     |     N/A      |           Standard TCP Socket which allows only IPv6 addresses            |
|     `udp`     |    Yes    |    Yes     |     N/A      |        Standard UDP Socket which allows IPv4 and/or IPv6 addresses        |
|    `udp4`     |    Yes    |    Yes     |     N/A      |           Standard UDP Socket which allows only IPv4 addresses            |
|    `udp6`     |    Yes    |    Yes     |     N/A      |           Standard UDP Socket which allows only IPv6 addresses            |
|    `unix`     |    Yes    |    Yes     |     N/A      |       Standard Unix Domain Socket which allows only absolute paths        |
|    `ldap`     |    No     |    Yes     |     389      |       Remote LDAP connection via TCP with implicit TLS via StartTLS       |
|    `ldaps`    |    No     |    Yes     |     636      |             Remote LDAP connection via TCP with explicit TLS              |
|    `ldapi`    |    No     |    Yes     |     N/A      |                  LDAP connection via Unix Domain Socket                   |
|    `smtp`     |    No     |    Yes     |      25      |      Remote SMTP connection via TCP using implicit TLS via StartTLS       |
| `submission`  |    No     |    Yes     |     587      | Remote SMTP Submission connection via TCP using implicit TLS via StartTLS |
| `submissions` |    No     |    Yes     |     465      |       Remote SMTP Submission connection via TCP using explicit TLS        |

If the scheme is absent, the default scheme is assumed. If the address has a `/` prefix it's assumed to be `unix`,
otherwise it's assumed to be`tcp`. If the scheme is `unix` it must be suffixed with an absolute path i.e.
`/var/local/authelia.sock` would be represented as `unix:///var/run/authelia.sock`.

#### hostname

The hostname is required if the scheme is one of the `tcp` or `udp` schemes and there is no [port](#port) specified. It
can be any IP that is locally addressable or a hostname which resolves to a locally addressable IP.

If specifying an IPv6 it should be wrapped in square brackets. For example for the IPv6 address `::1` with the `tcp`
scheme and port `80` the correct address would be `tcp://[::1]:80`.

#### port

The hostname is required if the scheme is one of the `tcp` or `udp` schemes and there is no [hostname](#hostname)
specified.

### Regular Expressions

We have several sections of configuration that utilize regular expressions. We use the Google RE2 regular expression
engine which is the full Go regular expression syntax engine, the syntax of which is described
[here](https://github.com/google/re2/wiki/Syntax) by the authors. It's very similar to regular expression engines like
PCRE, Perl, and Python; with the major exceptions being that it doesn't have backtracking.

It's recommended to validate your regular expressions manually either via tools like [Regex 101](https://regex101.com/)
(ensure you pick the `Golang` option) or some other means.

It's important when attempting to utilize a backslash that it's utilized correctly. The YAML parser is likely to parse
this as you trying to use YAML escape syntax instead of regex escape syntax. To avoid this use single quotes instead of
no quotes or double quotes.

Good Example:

```yaml
domain_regex: '^(admin|secure)\.example\.com$'
```

Bad Example:

```yaml
domain_regex: "^(admin|secure)\.example\.com$"
```

## Structures

The following represent common data structures used within the configuration which have specific requirements that are
used in multiple areas. This is intended on assisting in understanding these specific structures, and not as a specific
guide on configuring any particular instance.

### TLS

Various sections of the configuration use a uniform configuration section called TLS. Notably LDAP and SMTP.
This section documents the usage.

Various sections of the configuration use a uniform configuration section called `tls` which configure TLS socket and
TLS verification parameters. Notably the [LDAP](../first-factor/ldap.md#tls), [SMTP](../notifications/smtp.md#tls),
[PostgreSQL](../storage/postgres.md#tls), [MySQL](../storage/mysql.md#tls), and [Redis](../session/redis.md#tls)
sections. This section documents the common parts of this structure.

{{< details "Example: TLS" >}}
```yaml
tls:
  server_name: 'example.com'
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
{{< /details >}}

#### server_name

{{< confkey type="string" required="no" >}}

The key `server_name` overrides the name checked against the certificate in the verification process. Useful if you
require an IP address for the host of the backend service but want to verify a specific certificate server name.

#### skip_verify

{{< confkey type="boolean" default="false" required="no" >}}

The key `skip_verify` completely negates validating the certificate of the backend service. This is not recommended,
instead you should tweak the `server_name` option, and the global option
[certificates directory](../miscellaneous/introduction.md#certificates_directory).

#### minimum_version

{{< confkey type="string" default="TLS1.2" required="no" >}}

Controls the minimum TLS version Authelia will use when performing TLS handshakes.
The possible values are `TLS1.3`, `TLS1.2`, `TLS1.1`, `TLS1.0`, `SSL3.0`. Anything other than `TLS1.3` or `TLS1.2`
are very old and deprecated. You should avoid using these and upgrade your backend service instead of decreasing
this value. At the time of this writing `SSL3.0` will always produce errors.

#### maximum_version

{{< confkey type="string" default="TLS1.3" required="no" >}}

Controls the maximum TLS version Authelia will use when performing TLS handshakes.
The possible values are `TLS1.3`, `TLS1.2`, `TLS1.1`, `TLS1.0`, `SSL3.0`. Anything other than `TLS1.3` or `TLS1.2`
are very old and deprecated. You should avoid using these and upgrade your backend service instead of decreasing
this value. At the time of this writing `SSL3.0` will always produce errors.

#### certificate_chain

{{< confkey type="string" required="no" >}}

The certificate chain/bundle to be used with the [private_key](#private_key) to perform mutual TLS authentication with
the server.

The value must be one or more certificates encoded in the DER base64 ([RFC4648]) encoded PEM format. If more than one
certificate is provided, in top down order, each certificate must be signed by the next certificate if provided.

#### private_key

{{< confkey type="string" required="no" >}}

*__Important Note:__ This can also be defined using a [secret](../methods/secrets.md) which is __strongly recommended__
especially for containerized deployments.*

The private key to be used with the [certificate_chain](#certificate_chain) for mutual TLS authentication. The public key
material of the private key must match the private key of the first certificate in the
[certificate_chain](#certificate_chain).

The value must be one private key encoded in the DER base64 ([RFC4648]) encoded PEM format and must be encoded per the
[PKCS#8], [PKCS#1], or [SECG1] specifications.

[PKCS#8]: https://datatracker.ietf.org/doc/html/rfc5208
[PKCS#1]: https://datatracker.ietf.org/doc/html/rfc8017
[SECG1]: https://datatracker.ietf.org/doc/html/rfc5915
[RFC4648]: https://datatracker.ietf.org/doc/html/rfc4648

### Server Buffers

Various sections of the configuration use a uniform configuration section called `buffers` which configure HTTP server
buffers. Notably the [server](../miscellaneous/server.md#buffers) and
[metrics telemetry](../telemetry/metrics.md#buffers) sections. This section documents the common parts of this
structure.

{{< details "Example: Server Buffers" >}}
```yaml
buffers:
  read: 4096
  write: 4096
```
{{< /details >}}

#### read

{{< confkey type="integer" default="4096" required="no" >}}

Configures the maximum request size. The default of 4096 is generally sufficient for most use cases.

#### write

{{< confkey type="integer" default="4096" required="no" >}}

Configures the maximum response size. The default of 4096 is generally sufficient for most use cases.

### Server Timeouts

Various sections of the configuration use a uniform configuration section called `timeouts` which configure HTTP server
timeouts. Notably the [server](../miscellaneous/server.md#timeouts) and
[metrics telemetry](../telemetry/metrics.md#timeouts) sections. This section documents the common parts of this
structure.

{{< details "Example: Server Timeouts" >}}
```yaml
timeouts:
  read: '6s'
  write: '6s'
  idle: '30s'
```
{{< /details >}}

#### read

{{< confkey type="string,integer" syntax="duration" default="6 seconds" required="no" >}}

Configures the server read timeout.

#### write

{{< confkey type="string,integer" syntax="duration" default="6 seconds" required="no" >}}

Configures the server write timeout.

#### idle

{{< confkey type="string,integer" syntax="duration" default="30 seconds" required="no" >}}

Configures the server idle timeout.

## Historical References

This contains links to historical anchors.

##### Duration Notation Format

See [duration common syntax](#duration).

