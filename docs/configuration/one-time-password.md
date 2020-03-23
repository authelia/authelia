---
layout: default
title: One-Time Password
parent: Configuration
nav_order: 6
---

# One-Time Password

Authelia uses time based one-time passwords as the OTP method. You have 
the option to tune the settings of the TOTP generation and you can see an
an example full TOTP configuration below as well as sections describing them.

    totp:
        issuer: authelia.com
        period: 30
        skew: 1
        
## issuer

Applications generating one-time passwords usually display an issuer to
differentiate the various applications registered by the user.

Authelia allows to customize the issuer to differentiate the entry created
by Authelia from others.

## period and skew

The period and skew configuration parameters affect each other. The default values are
a period of 30 and a skew of 1. It is highly recommended you do not change these unless
you wish to set skew to 0.

The way you configure these affects security by changing the length of time a one-time
password is valid for. The formula to calculate the effective validity period is 
`period + (period * skew * 2)`. For example period 30 and skew 1 would result in 90 
seconds of validity, and period 30 and skew 2 would result in 150 seconds of validity.


### period

Configures the period of time in seconds a one-time password is current for.

It is recommended to keep this value set to 30, but the minimum is 1.
  
### skew

Configures the number of one-time passwords either side of the current one that are
considered valid, each time you increase this it makes two more one-time passwords valid. 
For example the default of 1 has a total of 3 keys valid. A value of 2 has 5 one-time passwords 
valid.

It is recommended to keep this value set to 0 or 1, but the minimum is 0.