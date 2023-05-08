---
title: "Common"
description: "Common configuration options and notations."
lead: "This section details common configuration elements within the Authelia configuration. This section is mainly used as a reference for other sections as necessary."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  configuration:
    parent: "prologue"
weight: 100200
toc: true
aliases:
  - /c/common
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

The following is ignored:
  - all spaces
  - leading zeros

While you can use multiple of these blocks in combination, we suggest keeping it simple and use a single value.

#### Unit Legend

##### Short Units

These values have been available for a long time.

|  Unit   | Associated Letter |
|:-------:|:-----------------:|
|  Years  |         y         |
| Months  |         M         |
|  Weeks  |         w         |
|  Days   |         d         |
|  Hours  |         h         |
| Minutes |         m         |
| Seconds |         s         |

##### Long Units

These values are more human readable but have only been available since v4.38.0.

|     Unit     |   Human Readable Long Unit    |
|:------------:|:-----------------------------:|
|    Years     |        `year`, `years`        |
|    Months    |       `month`, `months`       |
|    Weeks     |        `week`, `weeks`        |
|     Days     |         `day`, `days`         |
|    Hours     |        `hour`, `hours`        |
|   Minutes    |      `minute`, `minutes`      |
|   Seconds    |      `second`, `seconds`      |
| Milliseconds | `millisecond`, `milliseconds` |

#### Examples

|     Desired Value     |        Configuration Examples         |
|:---------------------:|:-------------------------------------:|
| 1 hour and 30 minutes | `90m` or `1h30m` or `5400` or `5400s` |
|         1 day         | `1d` or `24h` or `86400` or `86400s`  |
|       10 hours        | `10h` or `600m` or `9h60m` or `36000` |

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
[<scheme>://]<hostname>[:<port>]
```

##### Port

The following format represents the port format. It's valid only for a listener in most instances.
Refer to the individual documentation for an option for clarity. In this format as per the notation the scheme and
hostname are optional. The default for the scheme when not provided varies, and the default for the hostname is all
available addresses when not provided.

```text
[<scheme>://][hostname]:<port>
```

##### Unix Domain Socket

The following format represents the unix domain socket format. It's valid for both a listener and connector in most
instances. Refer to the individual documentation for an option for clarity. In this format as per the notation there
are no optional portions.

```text
unix://<path>
```

##### Examples

Various examples for these formats.

```text
0.0.0.0
tcp://0.0.0.0
tcp://0.0.0.0:9091
tcp://:9091
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

### TLS Configuration

Various sections of the configuration use a uniform configuration section called TLS. Notably LDAP and SMTP.
This section documents the usage.

#### server_name

{{< confkey type="string" required="no" >}}

The key `server_name` overrides the name checked against the certificate in the verification process. Useful if you
require an IP address for the host of the backend service but want to verify a specific certificate server name.

#### skip_verify

{{< confkey type="boolean" default="false" required="no" >}}

The key `skip_verify` completely negates validating the certificate of the backend service. This is not recommended,
instead you should tweak the `server_name` option, and the global option
[certificates directory](../miscellaneous/introduction.md#certificatesdirectory).

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

The certificate chain/bundle to be used with the [private_key](#privatekey) to perform mutual TLS authentication with
the server.

The value must be one or more certificates encoded in the DER base64 ([RFC4648]) encoded PEM format.

#### private_key

{{< confkey type="string" required="no" >}}

*__Important Note:__ This can also be defined using a [secret](../methods/secrets.md) which is __strongly recommended__
especially for containerized deployments.*

The private key to be used with the [certificate_chain](#certificatechain) for mutual TLS authentication.

The value must be one private key encoded in the DER base64 ([RFC4648]) encoded PEM format. If more than one certificate
is provided, in top down order, each certificate must be signed by the next certificate if provided.

[RFC4648]: https://datatracker.ietf.org/doc/html/rfc4648

### Server Buffers

#### read

{{< confkey type="integer" default="4096" required="no" >}}

Configures the maximum request size. The default of 4096 is generally sufficient for most use cases.

#### write

{{< confkey type="integer" default="4096" required="no" >}}

Configures the maximum response size. The default of 4096 is generally sufficient for most use cases.

### Server Timeouts

#### read

{{< confkey type="duration" default="6s" required="no" >}}

*__Reference Note:__ This configuration option uses the [duration common syntax](#duration).
Please see the [documentation](../prologue/common.md#duration) on this format for more information.*

Configures the server read timeout.

#### write

{{< confkey type="duration" default="6s" required="no" >}}

*__Reference Note:__ This configuration option uses the [duration common syntax](#duration).
Please see the [documentation](#duration) on this format for more information.*

Configures the server write timeout.

#### idle

{{< confkey type="duration" default="30s" required="no" >}}

*__Reference Note:__ This configuration option uses the [duration common syntax](#duration).
Please see the [documentation](#duration) on this format for more information.*

Configures the server idle timeout.

## Historical References

This contains links to historical anchors.

##### Duration Notation Format

See [duration common syntax](#duration).

