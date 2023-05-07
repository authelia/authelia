---
title: "Time-based OTP Applications"
description: "A Time-based OTP Application integration reference guide"
lead: "This section contains a Time-based OTP Application integration reference guide for Authelia."
date: 2022-11-19T16:47:09+11:00
draft: false
images: []
menu:
  reference:
    parent: "integrations"
weight: 320
toc: true
---

## Settings

Authelia allows for a wide variety of time-based OTP settings. There are several applications which can support these
algorithms and this matrix is a guide on applications that have been tested that work. It should not be assumed if an
application is on this list that the information is correct for the current version of a product and it's likely they
may now support some that were not previously supported, or in rare cases they may support less than they previously
did.


|      Application       |        Algorithm: SHA1         |       Algorithm: SHA256        |       Algorithm: SHA512        |           Digits: 6            |            Digits 8             |
|:----------------------:|:------------------------------:|:------------------------------:|:------------------------------:|:------------------------------:|:-------------------------------:|
| [Google Authenticator] | {{% support support="full" %}} |        {{% support %}}         |        {{% support %}}         | {{% support support="full" %}} |         {{% support %}}         |
|      [Bitwarden]       | {{% support support="full" %}} | {{% support support="full" %}} | {{% support support="full" %}} | {{% support support="full" %}} | {{% support support="full" %}}  |
| [Yubico Authenticator] | {{% support support="full" %}} |        {{% support %}}         |        {{% support %}}         | {{% support support="full" %}} | {{% support support="full" %}}  |
|  [Authenticator Plus]  | {{% support support="full" %}} |        {{% support %}}         |        {{% support %}}         | {{% support support="full" %}} |         {{% support %}}         |
|      [1Password]       | {{% support support="full" %}} | {{% support support="full" %}} |        {{% support %}}         | {{% support support="full" %}} |         {{% support %}}         |
|        [Ravio]         | {{% support support="full" %}} | {{% support support="full" %}} |        {{% support %}}         | {{% support support="full" %}} |         {{% support %}}         |
|        [Authy]         | {{% support support="full" %}} |        {{% support %}}         |        {{% support %}}         |        {{% support %}}         | {{% support  support="full" %}} |
|        [Aegis]         | {{% support support="full" %}} |        {{% support %}}         | {{% support support="full" %}} | {{% support support="full" %}} | {{% support  support="full" %}} |

[Google Authenticator]: https://play.google.com/store/apps/details?id=com.google.android.apps.authenticator2&hl=en&gl=US&pli=1
[Bitwarden]: https://bitwarden.com/
[Yubico Authenticator]: https://www.yubico.com/products/yubico-authenticator/
[Authenticator Plus]: https://www.authenticatorplus.com/
[1Password]: https://1password.com/
[Ravio]: https://raivo-otp.com/
[Authy]: https://authy.com/
[Aegis]: https://getaegis.app/

