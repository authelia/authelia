---
layout: default
title: One-Time Password
parent: Configuration
nav_order: 6
---

# One-Time Password

Authelia uses time based one-time passwords as the OTP method. You have 
the option to tune the settings of the TOTP generation and you can see a
full example of TOTP configuration below, as well as sections describing them.

    totp:
        issuer: authelia.com
        period: 30
        skew: 1
        
## Issuer

Applications generating one-time passwords usually display an issuer to
differentiate applications registered by the user.

Authelia allows customisation of the issuer to differentiate the entry created
by Authelia from others.

## Period and Skew

The period and skew configuration parameters affect each other. The default values are
a period of 30 and a skew of 1. It is highly recommended you do not change these unless
you wish to set skew to 0.

The way you configure these affects security by changing the length of time a one-time
password is valid for. The formula to calculate the effective validity period is 
`period + (period * skew * 2)`. For example period 30 and skew 1 would result in 90 
seconds of validity, and period 30 and skew 2 would result in 150 seconds of validity.


### Period

Configures the period of time in seconds a one-time password is current for. It is important
to note that changing this value will require your users to register their application again.

It is recommended to keep this value set to 30, the minimum is 1.
  
### Skew

Configures the number of one-time passwords either side of the current one that are
considered valid, each time you increase this it makes two more one-time passwords valid. 
For example the default of 1 has a total of 3 keys valid. A value of 2 has 5 one-time passwords 
valid.

It is recommended to keep this value set to 0 or 1, the minimum is 0.