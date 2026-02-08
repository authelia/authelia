---
title: "Time-based One-Time Password"
description: "Configuring the Time-based One-Time Password Second Factor Method."
summary: "Authelia supports utilizing Time-based One-Time Passwords as a 2FA method."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 103300
toc: true
aliases:
  - /c/totp
  - /docs/configuration/one-time-password.html
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

The OTP method *Authelia* uses is the Time-Based One-Time Password Algorithm (TOTP) [RFC6238] which is an extension of
HMAC-Based One-Time Password Algorithm (HOTP) [RFC4226].

You have the option to tune the settings of the TOTP generation, and you can see a full example of TOTP configuration
below, as well as sections describing them.

Keep in mind the default settings are chosen for compatibility. Many applications do not support digits other than 6,
and many only support SHA1.

## Configuration

{{< config-alert-example >}}

```yaml {title="configuration.yml"}
totp:
  disable: false
  issuer: 'authelia.com'
  algorithm: 'sha1'
  digits: 6
  period: 30
  skew: 1
  secret_size: 32
  allowed_algorithms:
    - 'SHA1'
  allowed_digits:
    - 6
  allowed_periods:
    - 30
  disable_reuse_security_policy: false
```

## Options

This section describes the individual configuration options.

### disable

{{< confkey type="boolean" default="false" required="no" >}}

This disables One-Time Password (TOTP) if set to true.

### issuer

{{< confkey type="string" default="Authelia" required="no" >}}

Applications generating Time-based One-Time Passwords usually display an issuer to
differentiate applications registered by the user.

Authelia allows customisation of the issuer to differentiate the entry created
by Authelia from others.

### algorithm

{{< confkey type="string" default="sha1" required="no" >}}

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
Many TOTP applications do not support this option. It is strongly advised you find out which
applications your users use and test them before changing this option. It is insufficient to test that the application
can add the key, it must also authenticate with Authelia as some applications silently ignore these options. See the
[Reference Guide](../../reference/integrations/time-based-one-time-password-apps.md) for tested applications.
{{< /callout >}}

[Bitwarden]: https://bitwarden.com/

The algorithm used for the TOTP key.

Possible Values (case-insensitive):

* `sha1`
* `sha256`
* `sha512`

Changing this value only affects newly registered TOTP keys. See the [Registration](#registration) section for more
information.

### digits

{{< confkey type="integer" default="6" required="no" >}}

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
Some TOTP applications do not support this option. It is strongly advised you find out which
applications your users use and test them before changing this option. It is insufficient to test that the application
can add the key, it must also authenticate with Authelia as some applications silently ignore these options. See the
[Reference Guide](../../reference/integrations/time-based-one-time-password-apps.md) for tested applications.
{{< /callout >}}

The number of digits a user needs to input to perform authentication. It's generally not recommended for this to be
altered as many TOTP applications do not support anything other than 6. What's worse is some TOTP applications allow
you to add the key, but do not use the correct number of digits specified by the key.

The valid values are `6` or `8`.

Changing this value only affects newly registered TOTP keys. See the [Registration](#registration) section for more
information.

### period

{{< confkey type="integer" default="30" required="no" >}}

The period of time in seconds between key rotations or the time element of TOTP. Please see the
[input validation](#input-validation) section for how this option and the [skew](#skew) option interact with each other.

It is recommended to keep this value set to 30, the minimum is 15.

Changing this value only affects newly registered TOTP keys. See the [Registration](#registration) section for more
information.

### skew

{{< confkey type="integer" default="1" required="no" >}}

The number of Time-based One-Time Passwords either side of the current valid Time-based One-Time Password that should
also be considered valid. The default of 1 results in 3 Time-based One-Time Passwords valid. A setting of 2 would result
in 5. With the default period of 30 this would result in 90 and 150 seconds of valid Time-based One-Time Passwords
respectively. Please see the [input validation](#input-validation) section for how this option and the [period](#period)
option interact with each other.

Changing this value affects all TOTP validations, not just newly registered ones.

### secret_size

{{< confkey type="integer" default="32" required="no" >}}

The length in bytes of generated shared secrets. The minimum is 20 (or 160 bits), and the default is 32 (or 256 bits).
In most use cases 32 is sufficient. Though some authenticators may have issues with more than the minimum. Our minimum
is the recommended value in [RFC4226], though technically according to the specification 16 bytes (or 128 bits) is the
minimum.

### allowed_algorithms

{{< confkey type="list(integer)" default="SHA1" required="no" >}}

Similar to [algorithm](#algorithm) with the same restrictions except this option allows users to pick from this list.
This list will always contain the value configured in the [algorithm](#algorithm) option.

### allowed_digits

{{< confkey type="list(string)" default="6" required="no" >}}

Similar to [digits](#digits) with the same restrictions except this option allows users to pick from this list. This
list will always contain the value configured in the [digits](#digits) option.

### allowed_periods

{{< confkey type="list(integer)" default="30" required="no" >}}

Similar to [period](#period) with the same restrictions except this option allows users to pick from this list. This
list will always contain the value configured in the [period](#period) option.

### disable_reuse_security_policy

{{< confkey type="boolean" default="false" required="no" >}}

Disables the policy which prevents reuse of a Time-based One-Time Password codes. This is an additional security measure
which prevents codes from being replayed. This should only affect codes which are used within the validity period more
than once.

## Registration

When users register their TOTP device for the first time, the current [issuer](#issuer), [algorithm](#algorithm), and
[period](#period) are used to generate the TOTP link and QR code. These values are saved to the database for future
validations.

This means if the configuration options are changed, users will not need to regenerate their keys. This functionality
takes effect from 4.33.0 onwards, previously the effect was the keys would just fail to validate. If you'd like to force
users to register a new device, you can delete the old device for a particular user by using the
`authelia storage user totp delete <username>` command regardless of if you change the settings or not.

## Input Validation

The period and skew configuration parameters affect each other. The default values are a period of 30 and a skew of 1.
It is highly recommended you do not change these unless you wish to set skew to 0.

These options affect security by changing the length of time a Time-based One-Time Password is valid for. The formula to
calculate the effective validity period is `period + (period * skew * 2)`. For example period 30 and skew 1 would result
in 90 seconds of validity, and period 30 and skew 2 would result in 150 seconds of validity.

## System time accuracy

It's important to note that if the system time is not accurate enough then clients will seemingly not generate valid
passwords for TOTP. Conversely this is the same when the client time is not accurate enough. This is due to the
Time-based One-Time Passwords being time-based.

Authelia by default checks the system time against an [NTP server](../miscellaneous/ntp.md) on startup. This helps to
prevent a time synchronization issue on the server being an issue. There is however no effective and reliable way to
check the clients.

## Encryption

The TOTP secret is [encrypted](../storage/introduction.md#encryption_key) in the database in version 4.33.0 and above.
This is so a user having access to only the database cannot easily compromise your two-factor authentication method.

This may be inconvenient for some users who wish to export TOTP keys from Authelia to other services. As such there is
a command specifically for exporting TOTP configurations from the database. These commands require the configuration or
at least a minimal configuration that has the storage backend connection details and the encryption key.

See the [CLI Documentation](../../reference/cli/authelia/authelia_storage_user_totp_export.md) for methods to perform
exports.

[RFC4226]: https://datatracker.ietf.org/doc/html/rfc4226
[RFC6238]: https://datatracker.ietf.org/doc/html/rfc6238
