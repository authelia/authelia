---
layout: default
title: Time-based One-Time Password
parent: Configuration
nav_order: 16
---

# Time-based One-Time Password

Authelia uses time-based one-time passwords as the OTP method. You have
the option to tune the settings of the TOTP generation, and you can see a
full example of TOTP configuration below, as well as sections describing them.

## Configuration
```yaml
totp:
  disable: false
  issuer: authelia.com
  algorithm: sha1
  digits: 6
  period: 30
  skew: 1
```

## Options

### disable
<div markdown="1">
type: boolean
{: .label .label-config .label-purple } 
default: false
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

This disables One-Time Password (TOTP) if set to true.

### issuer
<div markdown="1">
type: string
{: .label .label-config .label-purple } 
default: Authelia
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Applications generating one-time passwords usually display an issuer to
differentiate applications registered by the user.

Authelia allows customisation of the issuer to differentiate the entry created
by Authelia from others.

### algorithm
<div markdown="1">
type: string
{: .label .label-config .label-purple } 
default: sha1
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

_**Important Note:** Many TOTP applications do not support this option. It is strongly advised you find out which
applications your users use and test them before changing this option. It is insufficient to test that the application
can add the key, it must also authenticate with Authelia as some applications silently ignore these options. Bitwarden 
is the only one that has been tested at this time. If you'd like to contribute to documenting support for this option 
please see [Issue 2650](https://github.com/authelia/authelia/issues/2650)._

The algorithm used for the TOTP key.

Possible Values (case-insensitive):
- `sha1`
- `sha256`
- `sha512`

Changing this value only affects newly registered TOTP keys. See the [Registration](#registration) section for more
information.

### digits
<div markdown="1">
type: integer
{: .label .label-config .label-purple } 
default: 6
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

_**Important Note:** Some TOTP applications do not support this option. It is strongly advised you find out which
applications your users use and test them before changing this option. It is insufficient to test that the application
can add the key, it must also authenticate with Authelia as some applications silently ignore these options. Bitwarden
is the only one that has been tested at this time. If you'd like to contribute to documenting support for this option
please see [Issue 2650](https://github.com/authelia/authelia/issues/2650)._

The number of digits a user needs to input to perform authentication. It's generally not recommended for this to be 
altered as many TOTP applications do not support anything other than 6. What's worse is some TOTP applications allow
you to add the key, but do not use the correct number of digits specified by the key.

The valid values are `6` or `8`.

Changing this value only affects newly registered TOTP keys. See the [Registration](#registration) section for more
information.

### period
<div markdown="1">
type: integer
{: .label .label-config .label-purple } 
default: 30
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

The period of time in seconds between key rotations or the time element of TOTP. Please see the 
[input validation](#input-validation) section for how this option and the [skew](#skew) option interact with each other.

It is recommended to keep this value set to 30, the minimum is 15.

Changing this value only affects newly registered TOTP keys. See the [Registration](#registration) section for more
information.

### skew
<div markdown="1">
type: integer
{: .label .label-config .label-purple } 
default: 1
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

The number of one time passwords either side of the current valid one time password that should also be considered valid. 
The default of 1 results in 3 one time passwords valid. A setting of 2 would result in 5. With the default period of 30
this would result in 90 and 150 seconds of valid one time passwords respectively. Please see the 
[input validation](#input-validation) section for how this option and the [period](#period) option interact with each
other.

Changing this value affects all TOTP validations, not just newly registered ones.

## Registration
When users register their TOTP device for the first time, the current [issuer](#issuer), [algorithm](#algorithm), and 
[period](#period) are used to generate the TOTP link and QR code. These values are saved to the database for future
validations. 

This means if the configuration options are changed, users will not need to regenerate their keys. This functionality 
takes effect from 4.33.0 onwards, previously the effect was the keys would just fail to validate. If you'd like to force
users to register a new device, you can delete the old device for a particular user by using the 
`authelia storage totp delete <username>` command regardless of if you change the settings or not.

## Input Validation
The period and skew configuration parameters affect each other. The default values are a period of 30 and a skew of 1. 
It is highly recommended you do not change these unless you wish to set skew to 0.

The way you configure these affects security by changing the length of time a one-time
password is valid for. The formula to calculate the effective validity period is
`period + (period * skew * 2)`. For example period 30 and skew 1 would result in 90
seconds of validity, and period 30 and skew 2 would result in 150 seconds of validity.

## System time accuracy
It's important to note that if the system time is not accurate enough then clients will seemingly not generate valid
passwords for TOTP. Conversely this is the same when the client time is not accurate enough. This is due to the Time-based
One Time Passwords being time-based.

Authelia by default checks the system time against an [NTP server](./ntp.md#address) on startup. This helps to prevent
a time synchronization issue on the server being an issue. There is however no effective and reliable way to check the
clients.

## Encryption
The TOTP secret is [encrypted](storage/index.md#encryption_key) in the database in version 4.33.0 and above. This is so
a user having access to only the database cannot easily compromise your two-factor authentication method.

This may be inconvenient for some users who wish to export TOTP keys from Authelia to other services. As such there is
a command specifically for exporting TOTP configurations from the database. These commands require the configuration or
at least a minimal configuration that has the storage backend connection details and the encryption key.

Export in [Key URI Format](https://github.com/google/google-authenticator/wiki/Key-Uri-Format):

```shell
$ authelia storage totp export --format uri
```

Export as CSV:

```shell
$ authelia storage totp export --format csv
```

Help:

```shell
$ authelia storage totp export --help
```
