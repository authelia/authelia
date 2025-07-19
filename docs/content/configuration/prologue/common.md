---
title: "Common"
description: "Common configuration options and notations."
summary: "This section details common configuration elements within the Authelia configuration. This section is mainly used as a reference for other sections as necessary."
date: 2024-03-14T06:00:14+11:00
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

### Dictionary Reference

The dictionary reference syntax is a syntax where often the key can arbitrarily be set by an administrator and the key
can be used elsewhere to reference this configuration.

For instance, when considering the below example if the key named `policies` was noted as a dictionary within the
documentation then the `aribtrary_name` could be used elsewhere to communicate the policy to be applied, like in the
`usage_example` section where it's used as the `policy`.

```yaml
policies:
  arbitrary_name:
    enable: true

usage_example:
  - name: 'example'
    policy: 'arbitrary_name'
```

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

#### Query Parameters

Some schemes support parameters, this table describes them.

| Parameter | Listeners | Connectors |                                                                                                           Purpose                                                                                                            |
|:---------:|:---------:|:----------:|:----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------:|
|  `umask`  |    Yes    |     No     |                                             Sets the umask prior to creating the socket and restores it after creating it. The value must be an octal number with 3 or 4 digits.                                             |
|  `path`   |    Yes    |     No     | Sets the path variable to configure the subpath, specifically for a unix socket but technically works for TCP as well. Note that this should just be the alphanumeric portion it should not be prefixed with a forward slash |


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

##### File Descriptors

The following format represents the file descriptor format. It's valid only for a listener. Refer to the individual
documentation for an option for clarity. In this format as per the notation there are no optional portions.

The File Descriptor format also accepts a query string. The [Query Parameters](#query-parameters) described above
control certain behavior of this address type.

```text
fd://<file descriptor number>
```

```text
fd://<file descriptor number>?umask=0022
```

```text
fd://<file descriptor number>?path=auth
```

```text
fd://<file descriptor number>?umask=0022&path=auth
```

##### Unix Domain Socket

The following format represents the unix domain socket format. It's valid for both a listener and connector in most
instances. Refer to the individual documentation for an option for clarity. In this format as per the notation there
are no optional portions.

The Unix Domain Socket format also accepts a query string. The [Query Parameters](#query-parameters) described above
control certain behavior of this address type.

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
tcp://0.0.0.0:{{< sitevar name="port" nojs="9091" >}}
tcp://0.0.0.0:{{< sitevar name="port" nojs="9091" >}}/subpath
tcp://:{{< sitevar name="port" nojs="9091" >}}
tcp://:{{< sitevar name="port" nojs="9091" >}}/subpath
0.0.0.0:{{< sitevar name="port" nojs="9091" >}}

udp://0.0.0.0:123
udp://:123

unix:///var/lib/authelia.sock
```

#### scheme

The entire scheme is optional, but if the scheme host delimiter `://` is in the string, the scheme must be present. The
scheme must be one of the following (the listeners and connectors columns indicate support for the scheme on the
respective address type):

|    Scheme     | Listeners | Connectors | Default Port |                                     Notes                                      |
|:-------------:|:---------:|:----------:|:------------:|:------------------------------------------------------------------------------:|
|     `tcp`     |    Yes    |    Yes     |     N/A      |          Standard TCP Socket which allows IPv4 and/or IPv6 addresses           |
|    `tcp4`     |    Yes    |    Yes     |     N/A      |              Standard TCP Socket which allows only IPv4 addresses              |
|    `tcp6`     |    Yes    |    Yes     |     N/A      |              Standard TCP Socket which allows only IPv6 addresses              |
|     `udp`     |    Yes    |    Yes     |     N/A      |          Standard UDP Socket which allows IPv4 and/or IPv6 addresses           |
|    `udp4`     |    Yes    |    Yes     |     N/A      |              Standard UDP Socket which allows only IPv4 addresses              |
|    `udp6`     |    Yes    |    Yes     |     N/A      |              Standard UDP Socket which allows only IPv6 addresses              |
|    `unix`     |    Yes    |    Yes     |     N/A      |          Standard Unix Domain Socket which allows only absolute paths          |
|    `ldap`     |    No     |    Yes     |     389      |      Remote LDAP connection via a TCP socket using StartTLS if available       |
|    `ldaps`    |    No     |    Yes     |     636      |                    Remote LDAP connection via a TLS socket                     |
|    `ldapi`    |    No     |    Yes     |     N/A      |                     LDAP connection via Unix Domain Socket                     |
|    `smtp`     |    No     |    Yes     |      25      |      Remote SMTP connection via a TCP socket using StartTLS if available       |
| `submission`  |    No     |    Yes     |     587      | Remote SMTP Submission connection via a TCP socket using StartTLS if available |
| `submissions` |    No     |    Yes     |     465      |               Remote SMTP Submission connection via a TLS socket               |

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

```yaml {title="configuration.yml"}
domain_regex: '^(admin|secure)\.example\.com$'
```

Bad Example:

```yaml {title="configuration.yml"}
domain_regex: "^(admin|secure)\.example\.com$"
```

### Network

We support a network syntax which unmarshalls strings into a network range. The string format uses the standard CIDR
notation and assumes a single host (adapted as /32 for IPv4 and /128 for IPv6) if the CIDR suffix is absent.

|                  Example                  |                    CIDR                    |                                       Range                                       |
|:-----------------------------------------:|:------------------------------------------:|:---------------------------------------------------------------------------------:|
|                192.168.0.1                |               192.168.0.1/32               |                                    192.168.0.1                                    |
|              192.168.1.0/24               |               192.168.1.0/24               |                            192.168.1.0 - 192.168.1.255                            |
|              192.168.2.1/24               |               192.168.2.0/24               |                            192.168.2.0 - 192.168.2.255                            |
|  2001:db8:3333:4444:5555:6666:7777:8888   | 2001:db8:3333:4444:5555:6666:7777:8888/128 |                      2001:db8:3333:4444:5555:6666:7777:8888                       |
|          2001:db8:3333:4400::/56          |          2001:db8:3333:4400::/56           | 2001:0db8:3333:4400:0000:0000:0000:0000 - 2001:0db8:3333:44ff:ffff:ffff:ffff:ffff |
| 2001:db8:3333:4444:5555:6666:7777:8888/56 |          2001:db8:3333:4400::/56           | 2001:0db8:3333:4400:0000:0000:0000:0000 - 2001:0db8:3333:44ff:ffff:ffff:ffff:ffff |

## Structures

The following represent common data structures used within the configuration which have specific requirements that are
used in multiple areas. This is intended on assisting in understanding these specific structures, and not as a specific
guide on configuring any particular instance.

### TLS

Various sections of the configuration use a uniform configuration section called `tls` which configure TLS socket and
TLS verification parameters. This section documents the common parts of this structure. By default Authelia uses the
system certificate trust for TLS certificate verification but you can augment this with the global
[certificates_directory](../miscellaneous/introduction.md#certificates_directory) option, and can disable TLS
certificate verification entirely with the [skip_verify](#skip_verify) option.

```yaml {title="configuration.yml"}
tls:
  server_name: '{{< sitevar name="domain" nojs="example.com" >}}'
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

{{< confkey type="string" required="no" secret="yes" >}}

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

```yaml {title="configuration.yml"}
buffers:
  read: 4096
  write: 4096
```

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

```yaml {title="configuration.yml"}
timeouts:
  read: '6s'
  write: '6s'
  idle: '30s'
```

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

