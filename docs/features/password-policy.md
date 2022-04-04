---
layout: default
title: Password Policy
parent: Features
nav_order: 8
---

# Password Policy

Password policy enforces the security by requiring the users to use strong passwords.

Currently, two methods are supported:

## classic

This mode of operation allows administrators to set the rules that user passwords must comply with when changing their
password.

The available options are: 
  * Minimum password length
  * Require Uppercase
  * Require Lowercase
  * Require Numbers
  * Require Special characters


<p align="center">
  <img src="../images/password-policy-classic-1.png" width="400">
</p>


## zxcvbn

This mode uses [zxcvbn](https://github.com/dropbox/zxcvbn) for password strength checking. In this mode of operation, 
the user is not forced to follow any rules. The user is notified if their passwords is weak or strong.

<p align="center">
  <img src="../images/password-policy-zxcvbn-1.png" width="400">
</p>




