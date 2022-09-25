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

## Duration Notation Format

We have implemented a string/integer based notation for configuration options that take a duration of time. This section
describes the implementation of this. You can use this implementation in various areas of configuration such as:

* session:
  * expiration
  * inactivity
  * remember_me_duration
* regulation:
  * ban_time
  * find_time
* ntp:
  * max_desync
* webauthn:
  * timeout

The way this format works is you can either configure an integer or a string in the specific configuration areas. If you
supply an integer, it is considered a representation of seconds. If you supply a string, it parses the string in blocks
of quantities and units (number followed by a unit letter).  For example `5h` indicates a quantity of 5 units of `h`.

While you can use multiple of these blocks in combination, we suggest keeping it simple and use a single value.

### Unit Legend

|  Unit   | Associated Letter |
|:-------:|:-----------------:|
|  Years  |         y         |
| Months  |         M         |
|  Weeks  |         w         |
|  Days   |         d         |
|  Hours  |         h         |
| Minutes |         m         |
| Seconds |         s         |

### Examples

|     Desired Value     |        Configuration Examples         |
|:---------------------:|:-------------------------------------:|
| 1 hour and 30 minutes | `90m` or `1h30m` or `5400` or `5400s` |
|         1 day         | `1d` or `24h` or `86400` or `86400s`  |
|       10 hours        | `10h` or `600m` or `9h60m` or `36000` |

## Address

The address type is a string that takes the following format:

```text
[<scheme>://]<ip>[:<port>]
```

The square brackets indicate optional sections, and the angled brackets indicate required sections. The following
sections elaborate on this. Sections may only be optional for the purposes of parsing, there may be a configuration
requirement that one of these is provided.

### scheme

The entire scheme is optional, but if the scheme host delimiter `://` is in the string, the scheme must be present. The
scheme must be one of `tcp://`, or `udp://`. The default scheme is `tcp://`.

### ip

The IP is required. If specifying an IPv6 it should be wrapped in square brackets. For example for the IPv6 address
`::1` with the `tcp://` scheme and port `80`: `tcp://[::1]:80`.

### port

The entire port is optional, but if the host port delimiter `:` exists it must also include a numeric port.

## Regular Expressions

We have several sections of configuration that utilize regular expressions. It's recommended to validate your regex
manually either via tools like [Regex 101](https://regex101.com/) (ensure you pick the `Golang` option) or some other
means.

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

## TLS Configuration

Various sections of the configuration use a uniform configuration section called TLS. Notably LDAP and SMTP.
This section documents the usage.

### server_name

{{< confkey type="string" required="no" >}}

The key `server_name` overrides the name checked against the certificate in the verification process. Useful if you
require an IP address for the host of the backend service but want to verify a specific certificate server name.

### skip_verify

{{< confkey type="boolean" default="false" required="no" >}}

The key `skip_verify` completely negates validating the certificate of the backend service. This is not recommended,
instead you should tweak the `server_name` option, and the global option
[certificates directory](../miscellaneous/introduction.md#certificates_directory).

### minimum_version

{{< confkey type="string" default="TLS1.2" required="no" >}}

The key `minimum_version` controls the minimum TLS version Authelia will use when opening TLS connections.
The possible values are `TLS1.3`, `TLS1.2`, `TLS1.1`, `TLS1.0`. Anything other than `TLS1.3` or `TLS1.2`
are very old and deprecated. You should avoid using these and upgrade your backend service instead of decreasing
this value.

### client_auth_keypair

{{< confkey type="string" required="no" >}}

The optional TLS client authentication keypair used if the server requests client authentication in the TLS handshake.
Must conform to the following rules:

1. The value must contain at least one certificate and exactly one private key.
2. The certificates must be Base64 PEM Blocks encoded in ASN.1 DER format ([RFC5280] as per [RFC7468] in [section 5]).
3. The private keys must be a Base64 PEM Block encoded in a relevant ASN.1 DER format ([RFC5208] as per [RFC7468] in [section 10] for PKCS #8)
4. The private key must asymmetrically match the first certificate.

[RFC5208]: https://www.rfc-editor.org/rfc/rfc5208
[RFC5280]: https://www.rfc-editor.org/rfc/rfc5280
[RFC7468]: https://www.rfc-editor.org/rfc/rfc7468
[section 5]: https://www.rfc-editor.org/rfc/rfc7468#section-5
[section 10]: https://www.rfc-editor.org/rfc/rfc7468#section-10

#### Example Client Authentication Keypair

##### Single Certificate

This example only contains the public key for the provided private key. It does not contain the certificate chain.
Generally this all that's required.

```yaml
tls:
  client_auth_keypair: |
    -----BEGIN CERTIFICATE-----
    MIIC5jCCAc6gAwIBAgIRANQn+N/s2XpbyLjHhrhbLMYwDQYJKoZIhvcNAQELBQAw
    EzERMA8GA1UEChMIQXV0aGVsaWEwHhcNMjIwOTIzMDA1NTM5WhcNMjMwOTIzMDA1
    NTM5WjATMREwDwYDVQQKEwhBdXRoZWxpYTCCASIwDQYJKoZIhvcNAQEBBQADggEP
    ADCCAQoCggEBAJrdtDhT0LXvenka1pxPKN3ay82gGMuUHQPEwer+vhOjQmTuiNzz
    lYOQ5HRKkRippfCNwxf1QKpvllLwWxWh55MeFwgoIdy1ro2Q/sCIHsRdfIIgFxXu
    CBcyAPnfxUrujCIKhbvE72GC1MJWGCwtCJjPWSlqGUZPMFEy8n4WTtzgBDzx7tPU
    roH6tMyERO+LzLex6udNaY3L43TwIIdXfiK7X6tFtIcgySGNMoEoA3TzXvpr8N+5
    oPtFMx7nf8sBjV85AZ134ZAzsSVv/pPF1JyAasBT9n/b4zhr8t/km8BKxrwU4OZ6
    rx7ftZGtqx1k7czy3u0gCtFUcShItfL/vSECAwEAAaM1MDMwDgYDVR0PAQH/BAQD
    AgWgMBMGA1UdJQQMMAoGCCsGAQUFBwMBMAwGA1UdEwEB/wQCMAAwDQYJKoZIhvcN
    AQELBQADggEBAFZ5MPqio6v/4YysGXD+ljgXYTRdqb11FA1iMbEYTT80RGdD5Q/D
    f3MKgYfcq7tQTTYD05u4DEvxIv0VtFK9uyG3W13n/Bt+2Wv4bKuTIJEwpdnbFq8G
    nq2dmRtZL4K+oesdWOUXWcXouCN/M+b12Ik+9NlBbXkIBbz9/ni3i2FgaeN+cfGE
    ik4MjWBclSTMWQCB4jPhkunybzgdpTW+zhFBoZFHdbM3LlMTXJ5LXvWPGCcHy3c+
    XXgc6RG3GfuKWBOUfKJ/ejt6lKSI3vGkKgHjCAoHVsgHFz5CuGK3YISeX54sXA2D
    WXAcqD7v1ddNQKmE2eWZU4+2boBdXKMPtUQ=
    -----END CERTIFICATE-----
    -----BEGIN RSA PRIVATE KEY-----
    MIIEowIBAAKCAQEAmt20OFPQte96eRrWnE8o3drLzaAYy5QdA8TB6v6+E6NCZO6I
    3POVg5DkdEqRGKml8I3DF/VAqm+WUvBbFaHnkx4XCCgh3LWujZD+wIgexF18giAX
    Fe4IFzIA+d/FSu6MIgqFu8TvYYLUwlYYLC0ImM9ZKWoZRk8wUTLyfhZO3OAEPPHu
    09Sugfq0zIRE74vMt7Hq501pjcvjdPAgh1d+Irtfq0W0hyDJIY0ygSgDdPNe+mvw
    37mg+0UzHud/ywGNXzkBnXfhkDOxJW/+k8XUnIBqwFP2f9vjOGvy3+SbwErGvBTg
    5nqvHt+1ka2rHWTtzPLe7SAK0VRxKEi18v+9IQIDAQABAoIBAHS8dBIVk/jgmOhb
    A7T1sq9xMzk/2hDzB+AEW8yA09THts+QQxiSgHyZJqxGXRNDJjOrGImhtGoFDUJd
    rbsjvQTXpLLgVY4iYX6S8oU81jxc3/LSr7Q3JmAdsECqnfR61qT+W4qLy4osbaZD
    8ZqzI4zUl7gxIvYt0RUUG1hSBoZVJo4RjY2JG+YQbD1XYA7RYV9E/OKCH7yX/7T4
    iaIC0vN79yVHGPW6mvYE2INmCOCwabvSRKKqngVS/usm3cgo7LqL4OSqFVpHqf9F
    b+iP5fmzhexwQ/iZeXrv0Y9mGBjX0XpCNZRxiyDLuCXWvZeX/13Ae4/xEOipDw28
    mWBzyvUCgYEAx4PW0gEpVDMX2u3GmlK4WhCV3ArRDnYL0mU6xobQP6MuNYoyHh4L
    9yAt4YCmkTXx2Mc6kN6BJDG+6d0PwKnaMXqQGXjCo5+kdqzIYLCmHgG2Xv6SSqAU
    0KfaOZPgMaOmghGw2QqrBwWo+LEc92Fj4bmpNCjl1rZ+ZdpdGI7FoVsCgYEAxrXc
    goQAffsPIrNFOPKk5sOpN5cBHpaX62SZmTCEZjzvquBzE5b4jb5mbbw/5uFyeH39
    s74LRS24UrVjhJ5VaFKUNmUrAoUihb8eUl5lb3eUYIaKNz81YhDq369G0Ku4KRtc
    mwlnCFqY0G0ijK400lExlZR4ogIA9V1QHZHtSDMCgYEAjQ+p0tD/Z4i4VRHIWVQj
    A4q2ad078f2EXj00USkAE/5LrY8H4ENeMluOFOHg4spBNAOoZMTsiaqiULb7bDyr
    CFCfkWLQOt+kaEPBaJt817peNsvGovyLuvryT8M9v9r03wGjB9GDGnPmA+81i7JP
    7EhYWYiQ+D4PH/RD3hkTogECgYBPAwMyVmCHt2tWRegxc7IEHCrN8to8Gm8/5xl4
    IyWSLYqy7Fp1oKMmYV4DJkZWfLByns5hSSDcGgjfwkZW9kpJmARc+K84ak3G1q6s
    2+IDh43VL8oHm7eTTdzGosBKuu0YU0voTb3NQZDf13VUcPSJ6EUKECZDbP6KkdcI
    Wvz5pwKBgCffF+1xaMPOY8cVomvbnPfEAGes3EZGDq9TE36jDbQkOomBYZJ7+kPp
    mzhg2cUnDBiZ3eRGlfmBmgxIHdSweZQ5yxAyDBLfWxw9yqIX8bLyP7rBzxSy4efu
    6C3Nu7c0HX6rhcgrDFzZKc2TMuV8ihsMPfp5WSOqOIHXhUTfxPdh
    -----END RSA PRIVATE KEY-----
```

##### Multiple Certificates

This example contains the public key for the provided private key, plus it contains an additional certificate providing
chain information. This generally is not required unless the certificate trusted by the server for client authentication
signed a subordinate certificate authority which was used to sign the actual certificate. This may not be supported by
some servers at all.

```yaml
tls:
  client_auth_keypair: |
    -----BEGIN CERTIFICATE-----
    MIIC5jCCAc6gAwIBAgIRANQn+N/s2XpbyLjHhrhbLMYwDQYJKoZIhvcNAQELBQAw
    EzERMA8GA1UEChMIQXV0aGVsaWEwHhcNMjIwOTIzMDA1NTM5WhcNMjMwOTIzMDA1
    NTM5WjATMREwDwYDVQQKEwhBdXRoZWxpYTCCASIwDQYJKoZIhvcNAQEBBQADggEP
    ADCCAQoCggEBAJrdtDhT0LXvenka1pxPKN3ay82gGMuUHQPEwer+vhOjQmTuiNzz
    lYOQ5HRKkRippfCNwxf1QKpvllLwWxWh55MeFwgoIdy1ro2Q/sCIHsRdfIIgFxXu
    CBcyAPnfxUrujCIKhbvE72GC1MJWGCwtCJjPWSlqGUZPMFEy8n4WTtzgBDzx7tPU
    roH6tMyERO+LzLex6udNaY3L43TwIIdXfiK7X6tFtIcgySGNMoEoA3TzXvpr8N+5
    oPtFMx7nf8sBjV85AZ134ZAzsSVv/pPF1JyAasBT9n/b4zhr8t/km8BKxrwU4OZ6
    rx7ftZGtqx1k7czy3u0gCtFUcShItfL/vSECAwEAAaM1MDMwDgYDVR0PAQH/BAQD
    AgWgMBMGA1UdJQQMMAoGCCsGAQUFBwMBMAwGA1UdEwEB/wQCMAAwDQYJKoZIhvcN
    AQELBQADggEBAFZ5MPqio6v/4YysGXD+ljgXYTRdqb11FA1iMbEYTT80RGdD5Q/D
    f3MKgYfcq7tQTTYD05u4DEvxIv0VtFK9uyG3W13n/Bt+2Wv4bKuTIJEwpdnbFq8G
    nq2dmRtZL4K+oesdWOUXWcXouCN/M+b12Ik+9NlBbXkIBbz9/ni3i2FgaeN+cfGE
    ik4MjWBclSTMWQCB4jPhkunybzgdpTW+zhFBoZFHdbM3LlMTXJ5LXvWPGCcHy3c+
    XXgc6RG3GfuKWBOUfKJ/ejt6lKSI3vGkKgHjCAoHVsgHFz5CuGK3YISeX54sXA2D
    WXAcqD7v1ddNQKmE2eWZU4+2boBdXKMPtUQ=
    -----END CERTIFICATE-----
    -----BEGIN CERTIFICATE-----
    MIIC5jCCAc6gAwIBAgIRANQn+N/s2XpbyLjHhrhbLMYwDQYJKoZIhvcNAQELBQAw
    EzERMA8GA1UEChMIQXV0aGVsaWEwHhcNMjIwOTIzMDA1NTM5WhcNMjMwOTIzMDA1
    NTM5WjATMREwDwYDVQQKEwhBdXRoZWxpYTCCASIwDQYJKoZIhvcNAQEBBQADggEP
    ADCCAQoCggEBAJrdtDhT0LXvenka1pxPKN3ay82gGMuUHQPEwer+vhOjQmTuiNzz
    lYOQ5HRKkRippfCNwxf1QKpvllLwWxWh55MeFwgoIdy1ro2Q/sCIHsRdfIIgFxXu
    CBcyAPnfxUrujCIKhbvE72GC1MJWGCwtCJjPWSlqGUZPMFEy8n4WTtzgBDzx7tPU
    roH6tMyERO+LzLex6udNaY3L43TwIIdXfiK7X6tFtIcgySGNMoEoA3TzXvpr8N+5
    oPtFMx7nf8sBjV85AZ134ZAzsSVv/pPF1JyAasBT9n/b4zhr8t/km8BKxrwU4OZ6
    rx7ftZGtqx1k7czy3u0gCtFUcShItfL/vSECAwEAAaM1MDMwDgYDVR0PAQH/BAQD
    AgWgMBMGA1UdJQQMMAoGCCsGAQUFBwMBMAwGA1UdEwEB/wQCMAAwDQYJKoZIhvcN
    AQELBQADggEBAFZ5MPqio6v/4YysGXD+ljgXYTRdqb11FA1iMbEYTT80RGdD5Q/D
    f3MKgYfcq7tQTTYD05u4DEvxIv0VtFK9uyG3W13n/Bt+2Wv4bKuTIJEwpdnbFq8G
    nq2dmRtZL4K+oesdWOUXWcXouCN/M+b12Ik+9NlBbXkIBbz9/ni3i2FgaeN+cfGE
    ik4MjWBclSTMWQCB4jPhkunybzgdpTW+zhFBoZFHdbM3LlMTXJ5LXvWPGCcHy3c+
    XXgc6RG3GfuKWBOUfKJ/ejt6lKSI3vGkKgHjCAoHVsgHFz5CuGK3YISeX54sXA2D
    WXAcqD7v1ddNQKmE2eWZU4+2boBdXKMPtUQ=
    -----END CERTIFICATE-----
    -----BEGIN RSA PRIVATE KEY-----
    MIIEowIBAAKCAQEAmt20OFPQte96eRrWnE8o3drLzaAYy5QdA8TB6v6+E6NCZO6I
    3POVg5DkdEqRGKml8I3DF/VAqm+WUvBbFaHnkx4XCCgh3LWujZD+wIgexF18giAX
    Fe4IFzIA+d/FSu6MIgqFu8TvYYLUwlYYLC0ImM9ZKWoZRk8wUTLyfhZO3OAEPPHu
    09Sugfq0zIRE74vMt7Hq501pjcvjdPAgh1d+Irtfq0W0hyDJIY0ygSgDdPNe+mvw
    37mg+0UzHud/ywGNXzkBnXfhkDOxJW/+k8XUnIBqwFP2f9vjOGvy3+SbwErGvBTg
    5nqvHt+1ka2rHWTtzPLe7SAK0VRxKEi18v+9IQIDAQABAoIBAHS8dBIVk/jgmOhb
    A7T1sq9xMzk/2hDzB+AEW8yA09THts+QQxiSgHyZJqxGXRNDJjOrGImhtGoFDUJd
    rbsjvQTXpLLgVY4iYX6S8oU81jxc3/LSr7Q3JmAdsECqnfR61qT+W4qLy4osbaZD
    8ZqzI4zUl7gxIvYt0RUUG1hSBoZVJo4RjY2JG+YQbD1XYA7RYV9E/OKCH7yX/7T4
    iaIC0vN79yVHGPW6mvYE2INmCOCwabvSRKKqngVS/usm3cgo7LqL4OSqFVpHqf9F
    b+iP5fmzhexwQ/iZeXrv0Y9mGBjX0XpCNZRxiyDLuCXWvZeX/13Ae4/xEOipDw28
    mWBzyvUCgYEAx4PW0gEpVDMX2u3GmlK4WhCV3ArRDnYL0mU6xobQP6MuNYoyHh4L
    9yAt4YCmkTXx2Mc6kN6BJDG+6d0PwKnaMXqQGXjCo5+kdqzIYLCmHgG2Xv6SSqAU
    0KfaOZPgMaOmghGw2QqrBwWo+LEc92Fj4bmpNCjl1rZ+ZdpdGI7FoVsCgYEAxrXc
    goQAffsPIrNFOPKk5sOpN5cBHpaX62SZmTCEZjzvquBzE5b4jb5mbbw/5uFyeH39
    s74LRS24UrVjhJ5VaFKUNmUrAoUihb8eUl5lb3eUYIaKNz81YhDq369G0Ku4KRtc
    mwlnCFqY0G0ijK400lExlZR4ogIA9V1QHZHtSDMCgYEAjQ+p0tD/Z4i4VRHIWVQj
    A4q2ad078f2EXj00USkAE/5LrY8H4ENeMluOFOHg4spBNAOoZMTsiaqiULb7bDyr
    CFCfkWLQOt+kaEPBaJt817peNsvGovyLuvryT8M9v9r03wGjB9GDGnPmA+81i7JP
    7EhYWYiQ+D4PH/RD3hkTogECgYBPAwMyVmCHt2tWRegxc7IEHCrN8to8Gm8/5xl4
    IyWSLYqy7Fp1oKMmYV4DJkZWfLByns5hSSDcGgjfwkZW9kpJmARc+K84ak3G1q6s
    2+IDh43VL8oHm7eTTdzGosBKuu0YU0voTb3NQZDf13VUcPSJ6EUKECZDbP6KkdcI
    Wvz5pwKBgCffF+1xaMPOY8cVomvbnPfEAGes3EZGDq9TE36jDbQkOomBYZJ7+kPp
    mzhg2cUnDBiZ3eRGlfmBmgxIHdSweZQ5yxAyDBLfWxw9yqIX8bLyP7rBzxSy4efu
    6C3Nu7c0HX6rhcgrDFzZKc2TMuV8ihsMPfp5WSOqOIHXhUTfxPdh
    -----END RSA PRIVATE KEY-----
```

### certificate_chain

{{< confkey type="string" required="no" >}}

The certificate chain / bundle to be used for mutual TLS with the server.

#### Standard Certificate Chain Rules

X.509 Certificate Chains used with Authelia must conform to the following rules:

1. The certificates must be encoded in DER base64 PEM format ([RFC5280] as per [RFC7468] in [section 5]).
2. If more than one certificate is specified, the nth certificate must be signed by the nth+1 certificate where the
   order of certificates is top down.
3. Each certificate must have a valid date i.e. the current date must be after the not before property and before the
   not after property.
4. If the chain is intended to be used for mutual TLS the first certificate must have the extended key usage `clientAuth`.

#### Example Certificate Chains

##### Single

```yaml
tls:
  certificate_chain: |
    -----BEGIN CERTIFICATE-----
    MIIC5jCCAc6gAwIBAgIRANQn+N/s2XpbyLjHhrhbLMYwDQYJKoZIhvcNAQELBQAw
    EzERMA8GA1UEChMIQXV0aGVsaWEwHhcNMjIwOTIzMDA1NTM5WhcNMjMwOTIzMDA1
    NTM5WjATMREwDwYDVQQKEwhBdXRoZWxpYTCCASIwDQYJKoZIhvcNAQEBBQADggEP
    ADCCAQoCggEBAJrdtDhT0LXvenka1pxPKN3ay82gGMuUHQPEwer+vhOjQmTuiNzz
    lYOQ5HRKkRippfCNwxf1QKpvllLwWxWh55MeFwgoIdy1ro2Q/sCIHsRdfIIgFxXu
    CBcyAPnfxUrujCIKhbvE72GC1MJWGCwtCJjPWSlqGUZPMFEy8n4WTtzgBDzx7tPU
    roH6tMyERO+LzLex6udNaY3L43TwIIdXfiK7X6tFtIcgySGNMoEoA3TzXvpr8N+5
    oPtFMx7nf8sBjV85AZ134ZAzsSVv/pPF1JyAasBT9n/b4zhr8t/km8BKxrwU4OZ6
    rx7ftZGtqx1k7czy3u0gCtFUcShItfL/vSECAwEAAaM1MDMwDgYDVR0PAQH/BAQD
    AgWgMBMGA1UdJQQMMAoGCCsGAQUFBwMBMAwGA1UdEwEB/wQCMAAwDQYJKoZIhvcN
    AQELBQADggEBAFZ5MPqio6v/4YysGXD+ljgXYTRdqb11FA1iMbEYTT80RGdD5Q/D
    f3MKgYfcq7tQTTYD05u4DEvxIv0VtFK9uyG3W13n/Bt+2Wv4bKuTIJEwpdnbFq8G
    nq2dmRtZL4K+oesdWOUXWcXouCN/M+b12Ik+9NlBbXkIBbz9/ni3i2FgaeN+cfGE
    ik4MjWBclSTMWQCB4jPhkunybzgdpTW+zhFBoZFHdbM3LlMTXJ5LXvWPGCcHy3c+
    XXgc6RG3GfuKWBOUfKJ/ejt6lKSI3vGkKgHjCAoHVsgHFz5CuGK3YISeX54sXA2D
    WXAcqD7v1ddNQKmE2eWZU4+2boBdXKMPtUQ=
    -----END CERTIFICATE-----
```
##### Multiple

```yaml
tls:
  certificate_chain: |
    -----BEGIN CERTIFICATE-----
    MIIC5jCCAc6gAwIBAgIRANQn+N/s2XpbyLjHhrhbLMYwDQYJKoZIhvcNAQELBQAw
    EzERMA8GA1UEChMIQXV0aGVsaWEwHhcNMjIwOTIzMDA1NTM5WhcNMjMwOTIzMDA1
    NTM5WjATMREwDwYDVQQKEwhBdXRoZWxpYTCCASIwDQYJKoZIhvcNAQEBBQADggEP
    ADCCAQoCggEBAJrdtDhT0LXvenka1pxPKN3ay82gGMuUHQPEwer+vhOjQmTuiNzz
    lYOQ5HRKkRippfCNwxf1QKpvllLwWxWh55MeFwgoIdy1ro2Q/sCIHsRdfIIgFxXu
    CBcyAPnfxUrujCIKhbvE72GC1MJWGCwtCJjPWSlqGUZPMFEy8n4WTtzgBDzx7tPU
    roH6tMyERO+LzLex6udNaY3L43TwIIdXfiK7X6tFtIcgySGNMoEoA3TzXvpr8N+5
    oPtFMx7nf8sBjV85AZ134ZAzsSVv/pPF1JyAasBT9n/b4zhr8t/km8BKxrwU4OZ6
    rx7ftZGtqx1k7czy3u0gCtFUcShItfL/vSECAwEAAaM1MDMwDgYDVR0PAQH/BAQD
    AgWgMBMGA1UdJQQMMAoGCCsGAQUFBwMBMAwGA1UdEwEB/wQCMAAwDQYJKoZIhvcN
    AQELBQADggEBAFZ5MPqio6v/4YysGXD+ljgXYTRdqb11FA1iMbEYTT80RGdD5Q/D
    f3MKgYfcq7tQTTYD05u4DEvxIv0VtFK9uyG3W13n/Bt+2Wv4bKuTIJEwpdnbFq8G
    nq2dmRtZL4K+oesdWOUXWcXouCN/M+b12Ik+9NlBbXkIBbz9/ni3i2FgaeN+cfGE
    ik4MjWBclSTMWQCB4jPhkunybzgdpTW+zhFBoZFHdbM3LlMTXJ5LXvWPGCcHy3c+
    XXgc6RG3GfuKWBOUfKJ/ejt6lKSI3vGkKgHjCAoHVsgHFz5CuGK3YISeX54sXA2D
    WXAcqD7v1ddNQKmE2eWZU4+2boBdXKMPtUQ=
    -----END CERTIFICATE-----
    -----BEGIN CERTIFICATE-----
    MIIC5jCCAc6gAwIBAgIRANQn+N/s2XpbyLjHhrhbLMYwDQYJKoZIhvcNAQELBQAw
    EzERMA8GA1UEChMIQXV0aGVsaWEwHhcNMjIwOTIzMDA1NTM5WhcNMjMwOTIzMDA1
    NTM5WjATMREwDwYDVQQKEwhBdXRoZWxpYTCCASIwDQYJKoZIhvcNAQEBBQADggEP
    ADCCAQoCggEBAJrdtDhT0LXvenka1pxPKN3ay82gGMuUHQPEwer+vhOjQmTuiNzz
    lYOQ5HRKkRippfCNwxf1QKpvllLwWxWh55MeFwgoIdy1ro2Q/sCIHsRdfIIgFxXu
    CBcyAPnfxUrujCIKhbvE72GC1MJWGCwtCJjPWSlqGUZPMFEy8n4WTtzgBDzx7tPU
    roH6tMyERO+LzLex6udNaY3L43TwIIdXfiK7X6tFtIcgySGNMoEoA3TzXvpr8N+5
    oPtFMx7nf8sBjV85AZ134ZAzsSVv/pPF1JyAasBT9n/b4zhr8t/km8BKxrwU4OZ6
    rx7ftZGtqx1k7czy3u0gCtFUcShItfL/vSECAwEAAaM1MDMwDgYDVR0PAQH/BAQD
    AgWgMBMGA1UdJQQMMAoGCCsGAQUFBwMBMAwGA1UdEwEB/wQCMAAwDQYJKoZIhvcN
    AQELBQADggEBAFZ5MPqio6v/4YysGXD+ljgXYTRdqb11FA1iMbEYTT80RGdD5Q/D
    f3MKgYfcq7tQTTYD05u4DEvxIv0VtFK9uyG3W13n/Bt+2Wv4bKuTIJEwpdnbFq8G
    nq2dmRtZL4K+oesdWOUXWcXouCN/M+b12Ik+9NlBbXkIBbz9/ni3i2FgaeN+cfGE
    ik4MjWBclSTMWQCB4jPhkunybzgdpTW+zhFBoZFHdbM3LlMTXJ5LXvWPGCcHy3c+
    XXgc6RG3GfuKWBOUfKJ/ejt6lKSI3vGkKgHjCAoHVsgHFz5CuGK3YISeX54sXA2D
    WXAcqD7v1ddNQKmE2eWZU4+2boBdXKMPtUQ=
    -----END CERTIFICATE-----
```

## Server Buffers

### read

{{< confkey type="integer" default="4096" required="no" >}}

Configures the maximum request size. The default of 4096 is generally sufficient for most use cases.

### write

{{< confkey type="integer" default="4096" required="no" >}}

Configures the maximum response size. The default of 4096 is generally sufficient for most use cases.

## Server Timeouts

### read

{{< confkey type="duration" default="2s" required="no" >}}

*__Note:__ This setting uses the [duration notation format](#duration-notation-format). Please see the
[common options](#duration-notation-format) documentation for information on this format.*

Configures the server read timeout.

### write

{{< confkey type="duration" default="2s" required="no" >}}

*__Note:__ This setting uses the [duration notation format](#duration-notation-format). Please see the
[common options](#duration-notation-format) documentation for information on this format.*

Configures the server write timeout.

### idle

{{< confkey type="duration" default="30s" required="no" >}}

*__Note:__ This setting uses the [duration notation format](#duration-notation-format). Please see the
[common options](#duration-notation-format) documentation for information on this format.*

Configures the server write timeout.
