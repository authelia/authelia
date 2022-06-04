---
title: "Password Policy"
description: "Authelia utilizes WebAuthn security keys as one of it's second factor authentication methods."
lead: "Authelia utilizes WebAuthn security keys as one of it's second first authentication methods."
date: 2022-04-12T14:40:22+10:00
lastmod: 2022-06-03T10:43:55+10:00
draft: false
images: []
menu:
  overview:
    parent: "authentication"
weight: 260
toc: true
---

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

{{< figure src="password-policy-classic-1.png" caption="Classic Password Policy" alt="Classic Password Policy" width=400 >}}

## zxcvbn

This mode uses [zxcvbn](https://github.com/dropbox/zxcvbn) for password strength checking. In this mode of operation,
the user is not forced to follow any rules. The user is notified if their passwords is weak or strong.

{{< figure src="password-policy-zxcvbn-1.png" caption="zxcvbn Password Policy" alt="zxcvbn Password Policy" width=400 >}}
