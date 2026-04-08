---
title: "Validating Forwarded Authentication"
description: "A reference guide on validating that the Authelia Forwarded Authentication proxy integration is operating correctly after configuration or changes."
summary: "This section contains reference documentation for validating the Forwarded Authentication Integration."
date: 2024-03-14T06:00:14+11:00
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

The [Forwarded Authentication Integration](../../integration/proxies/introduction.md) requires that users validate the
configuration is operational in several scenarios such as:

1. After initial configuration.
2. After making changes to the proxy configuration for Authelia or the relevant integration URL.
3. After making changes to the [server address](../../configuration/miscellaneous/server.md#address) value.
4. After changing the proxy configuration of an app that leverages the integration.
5. After changing Authelia's access control rules.

It's also recommended that users take a moment to validate it when upgrading their proxy as a proxy bug, a change
in how it operates in regards to the integration, or the upgrade deleting or changing the configuration could result in
failures.

## Steps

### Validating General Operation

This validation is important for all users.

1. Identify an application that:
   1. Uses the integration.
   2. Requires authentication to be accessed.
   3. You wish to validate.
2. Ensure you're logged out of Authelia itself.
3. Visit the application you picked in 1 and ensure that:
   1. You are redirected to the Authelia login portal.
   2. That you are required to perform the expected level of authentication.

### Validating Network Access Control Rules

This validation is critical for anyone wishing to use the
[networks](../../configuration/security/access-control.md#networks) option in Access Control Rules. These steps ensure
your proxies do not arbitrarily trust the `X-Forwarded-For` header as described in
[Forwarded Headers](../../integration/proxies/forwarded-headers/index.md).

1. Ensure the checks in [Validating General Operation](#validating-general-operation) are complete.
2. Identify an application that:
   1. Uses the integration.
   2. Requires authentication to be accessed.
   3. You wish to validate.
3. Assuming the application you identified in 2 is on the `app.example.com` domain, add the rule outlined below to the
   very top of the Access Control Rules.
4. Perform the following command `curl -i -H 'X-Forwarded-For: 169.254.1.2' https://app.example.com`
5. Ensure the response looks like the example below with `302` on the first line and `302 Found` on the last line.
6. Remove the rule we added in 3.

#### Example Configuration

```yaml
access_control:
  rules:
    - domain: 'app.example.com'
      policy: 'bypass'
      networks:
        - '169.254.1.2'
    # Your normal rules here.
```

#### Example Secure Output

```
HTTP/2 302
alt-svc: h3=":443"; ma=2592000
content-type: text/html; charset=utf-8
date: Sat, 21 Mar 2026 04:35:35 GMT
location: https://auth.example.com/?rd=https%3A%2F%2Fapp.example.com%2F&rm=GET
permissions-policy: accelerometer=(), autoplay=(), camera=(), display-capture=(), geolocation=(), gyroscope=(), keyboard-map=(), magnetometer=(), microphone=(), midi=(), payment=(), picture-in-picture=(), screen-wake-lock=(), sync-xhr=(), xr-spatial-tracking=(), interest-cohort=()
referrer-policy: strict-origin-when-cross-origin
set-cookie: authelia-session=Zdlhz6#ZTKPg5MOul3!TRLWv4sb$RznL; expires=Sat, 21 Mar 2026 05:35:36 GMT; domain=example.com; path=/; HttpOnly; secure; SameSite=Lax
x-content-type-options: nosniff
x-dns-prefetch-control: off
x-frame-options: DENY
content-length: 119

<a href="https://auth.example.com/?rd=https%3A%2F%2Fapp.example.com%2F&amp;rm=GET">302 Found</a>
```
