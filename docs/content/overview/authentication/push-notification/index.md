---
title: "Duo / Mobile Push"
description: "Authelia utilizes Duo Push Notifications as one of it's second factor authentication methods."
summary: "Authelia utilizes Duo Push Notifications as one of it's second factor authentication methods."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 250
toc: true
aliases:
  - /docs/features/2fa/push-notifications.html
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

Mobile Push notifications are a really convenient and trendy method to perform 2FA. When 2FA is required Authelia sends
a notification directly to an application on your mobile phone where you can instantly choose to accept or deny.

Authelia leverages [Duo] third party to provide this feature.

{{< figure src="duo-push-1.jpg" caption="The Duo Mobile Push authorization notification" alt="The Duo Mobile Push authorization notification" sizes="50dvh" >}}

{{< figure src="duo-push-2.png" caption="The Duo Mobile Push authorization consent view" alt="The Duo Mobile Push authorization consent view" sizes="50dvh" >}}

First, sign up on their website, log in, create a user account and attach it a mobile device. Beware that the name of
the user must match the name of the user in Authelia, or must have an alias that matches the user in Authelia.

Then, in Duo interface, click on *Applications* and *Protect an Application*. Select the option *Partner Auth API*. This
will generate an integration key, a secret key and a hostname. You can set the name of the application to __Authelia__
and then you must add the generated information to Authelia [configuration](../../../configuration/second-factor/duo.md).

See the [configuration documentation](../../../configuration/second-factor/duo.md) for more details.

Now that Authelia is configured, pass the first factor and select the Push notification option.

{{< figure src="2FA-PUSH.png" caption="The Mobile Push 2FA view" alt="The Mobile Push 2FA view" sizes="50dvh" >}}

You should now receive a notification on your mobile phone with all the details about the authentication request. In
case you have multiple devices available, you will be asked to select your preferred device.

## Frequently Asked Questions

### Why don't I have access to the *Push Notification* option?

It's likely that you have not configured __Authelia__ correctly. Please read this documentation again and be sure you
had a look at {{< github-link path="config.template.yml" >}} and
[configuration documentation](../../../configuration/second-factor/duo.md).

### I have access to the *Push Notification* option, but there is an error message: *"no compatible device found".*

There is a problem with your **Users** configuration in Duo. There are no users configured in Duo that match your Authelia user. Note that the admin user you create when you sign up is not automatically added as a **User** in Duo.

[Duo]: https://duo.com/
