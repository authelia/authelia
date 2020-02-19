---
layout: default
title: One-Time Password
parent: Configuration
nav_order: 6
---

# One-Time Password

Applications generating one-time passwords usually displays an issuer to
differentiate the various applications registered by the user.

Authelia allows to customize the issuer to differentiate the entry created
by Authelia from others.

    totp:
        issuer: authelia.com