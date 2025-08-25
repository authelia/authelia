---
title: "Forwarded Headers"
description: "An introduction into the importance of forwarded headers coming from trusted sources"
summary: "An introduction into the importance of forwarded headers coming from trusted sources."
date: 2024-03-17T20:37:08+11:00
draft: false
images: []
weight: 312
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

The`X-Forwarded-*` headers presented to __Authelia__ must be from trusted sources. As such you must ensure that the
reverse proxies and load balancers utilized with __Authelia__ are configured to remove and replace specific headers when
they come directly from clients and not from proxies in your trusted environment.

Some proxies require users explicitly configure the proxy to trust another proxy, however some implicitly trust all
headers regardless of the source so you will have to manually configure them.

## Network Rules

In particular this is important for [Access Control Rules](../../../configuration/security/access-control.md#rules) as
the [network criteria](../../../configuration/security/access-control.md#networks) relies on the [X-Forwarded-For]
header. This header is expected to have a true representation of the client's actual IP address.

If this is not removed from non-trusted proxies a user could theoretically hijack any rule that contains this criteria
to potentially skip an authentication criteria depending on how it is configured.

## Cloud Proxies

In addition to configuring your own proxies to remove this header from untrusted sources, when using a cloud proxy like
[Cloudflare](#cloudflare) you must ensure they do this or you configure a rule to do it. We aim to have documentation
in this section for cloud proxies that do this, but you should test this yourself and check the documentation for the
cloud proxy.

In addition to this it's important if you wish to preserve the clients actual IP address that you trust the IP addresses
of the cloud proxy in your on-premise proxies. If you don't do this most if not all proxies configured as per our guides
will remove the header and everyone external will appear to come from a proxies source IP address rather than their real
IP address in both logging and access control.

These same rules apply to any off-site hosted proxy or load balancing solution that alters the source IP address.

### Cloudflare

[Cloudflare] adds the [X-Forwarded-For] header if it does not exist, and if it does exist it will just append another IP
to it. This means a client can forge their remote IP address with the most widely accepted remote IP header out of the
box.

It is therefore important you configure [Cloudflare] to remove this IP address. *__Please Note:__ This is by no means an
exhaustive guide on using [Cloudflare] transform rules, however it's enough to configure a couple rules which should
achieve a secure result. Please see the [Cloudflare] documentation on
[transform rules](https://developers.cloudflare.com/rules/transform/) for more information._

#### Steps

1. On the left sidebar, click `Rules`.
2. Click `Transform Rules`.
3. On the right, click `Modify Request Header` tab.
4. Click `Create rule`.
5. Set the `Rule name` to something appropriate like `Remove X-Forwarded-For Header`.
6. Set the `Field` option in the `When incoming requests match` section to an appropriate value (see criteria table
   below).
7. Set the `Operator` option in the `When incoming requests match` section to an appropriate value (see criteria table
   below).
8. Set the `Value` option in the `When incoming requests match` section to an appropriate value (see criteria table
   below).
9. Set the `Then` section dropdown to `Remove`.
10. Set the `Then` section `Header name` to `X-Forwarded-For`.
11. Click `Deploy`.

{{< figure
src="cloudflare_1.png"
alt="Image of Cloudflare dashboard with steps 1 to 4 labeled"
width="736"
caption="Steps 1 - 4: Create New Request Header Transform Rule"
title="Steps 1 - 4: Create New Request Header Transform Rule" >}}

{{< figure
src="cloudflare_2.png"
alt="Image of Cloudflare dashboard with steps 5 to 11 labeled for Option 1"
width="736"
caption="Steps 5 - 11 Option 1: Always Remove Headers"
title="Steps 5 - 11 Option 1: Always Remove Headers" >}}

{{< figure
src="cloudflare_3.png"
alt="Image of Cloudflare dashboard with steps 5 to 11 labeled for Option 2"
width="736"
caption="Steps 5 - 11 Option 2: Remove When Not From Trusted Source"
title="Steps 5 - 11 Option 2: Remove When Not From Trusted Source" >}}

#### Criteria

This table describes the criteria needed to achieve a desired result. Only one of these options should be chosen. You
Should look at the desired result column and apply the appropriate field, operator, and value to [step](#steps) 8.
Generally speaking the `Always Remove` option is the correct option.

|           Desired Result            |       Field       |    Operator    |               Value                |
|:-----------------------------------:|:-----------------:|:--------------:|:----------------------------------:|
|            Always Remove            |  X-Forwarded-For  | does not equal |              *blank*               |
| Remove When Not From Trusted Source | IP Source Address |   is not in    | *list of trusted source addresses* |

Cloudflare publishes its IP address ranges publicly at the easy to remember address
[https://www.cloudflare.com/ips/](https://www.cloudflare.com/ips/). You should use this with the trusted proxies section
of your relevant proxy to ensure it's trusted if you intend to use Cloudflare.

[X-Forwarded-For]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-For
[Cloudflare]: https://www.cloudflare.com
