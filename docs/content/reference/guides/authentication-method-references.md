---
title: "Authentication Method Reference Values"
description: "This guide shows a list of Authentication Method Reference Values based on RFC8176 and how they are implemented within Authelia"
summary: "This guide shows a list of other frequently asked question documents as well as some general ones."
date: 2024-10-04T21:06:51+10:00
draft: false
images: []
weight: 220
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

Authelia leverages [RFC8176] in a few areas to communicate and control authorizatiaon.

Below is a list of the Authentication Method Reference Values and how we currently support them:

| Value  |                            Description                            |   Factor   | Channel  |
|:------:|:-----------------------------------------------------------------:|:----------:|:--------:|
|  mfa   |      User used multiple factors to login (see factor column)      |    N/A     |   N/A    |
|  mca   |     User used multiple channels to login (see channel column)     |    N/A     |   N/A    |
|  user  |  User confirmed they were present when using their hardware key   |    N/A     |   N/A    |
|  pin   | User confirmed they are the owner of the hardware key with a pin  |    N/A     |   N/A    |
|  kba   |         User used a knowledge based authentication factor         | Knowledge  |   N/A    |
|  pwd   |            User used a username and password to login             | Knowledge  | Browser  |
|  otp   |                      User used TOTP to login                      | Possession | Browser  |
|  pop   | User used a software or hardware proof-of-possession key to login | Possession | Browser  |
|  hwk   |       User used a hardware proof-of-possession key to login       | Possession | Browser  |
|  swk   |       User used a software proof-of-possession key to login       | Possession | Browser  |
|  sms   |                      User used Duo to login                       | Possession | External |
|  face  |                           _Unsupported_                           |    N/A     |   N/A    |
|  fpt   |                           _Unsupported_                           |    N/A     |   N/A    |
|  geo   |                           _Unsupported_                           |    N/A     |   N/A    |
|  iris  |                           _Unsupported_                           |    N/A     |   N/A    |
|  rba   |                           _Unsupported_                           |    N/A     |   N/A    |
| retina |                           _Unsupported_                           |    N/A     |   N/A    |
|   sc   |                           _Unsupported_                           |    N/A     |   N/A    |
|  tel   |                           _Unsupported_                           |    N/A     |   N/A    |
|  vbm   |                           _Unsupported_                           |    N/A     |   N/A    |
|  wia   |                           _Unsupported_                           |    N/A     |   N/A    |


[RFC8176]: https://datatracker.ietf.org/doc/html/rfc8176
