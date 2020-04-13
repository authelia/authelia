---
layout: default
title: Password Reset
parent: Features
nav_order: 4
---

# Password Reset

**Authelia** provides a workflow to let users reset their password when they lose it.
To disable reset password functionality please see the [configuration docs](../configuration/authentication/index.md#disabling-reset-password).

A simple click on `Reset password?` for starting the process. Note that resetting a
password requires a new identity verification using the e-mail of the user.

<p align="center">
  <img src="../images/1FA.png" width="400">
</p>

Give your username and receive an e-mail to verify your identity.

<p align="center">
  <img src="../images/RESET-PASSWORD-STEP1.png" width="400">
</p>

Once your identity has been verified, fill in the form to reset your password.

<p align="center">
  <img src="../images/RESET-PASSWORD-STEP2.png" width="400">
</p>

Now you can authenticate with your new credentials.