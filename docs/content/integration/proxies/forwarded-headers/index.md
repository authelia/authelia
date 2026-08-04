---
title: "Forwarded Headers"
description: "An introduction into the importance of forwarded headers coming from trusted sources and configuring reverse proxies to ensure header integrity and security."
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

#### Remove Client IPs from X-Forwarded-For Header

[Cloudflare] has managed rules with one of them removing client IPs from the X-Forwarded-For header. *__Please Note:__ This is by no means an
exhaustive guide on using [Cloudflare] managed transforms, however it's enough to configure this rule which should
achieve a secure result. Please see the [Cloudflare] documentation on
[managed transforms](https://developers.cloudflare.com/rules/transform/managed-transforms/) for more information._

##### Method 1 Steps

`Rules → Overview → Create rule → Request Header Transform Rule`

{{< figure
src="cloudflare.png"
alt="Image of Cloudflare dashboard with steps 1 to 4 labeled"
width="736"
caption="Steps 1 - 4: Image of Cloudflare dashboard with steps 1 to 4 labeled for method 1"
title="Steps 1 - 4: Image of Cloudflare dashboard with steps 1 to 4 labeled for method 1" >}}

#### Allow Trusted IPs to Add Client IPs to X-Forwarded-For Header

The Managed Transforms option removes visitor IP values from the X-Forwarded-For header regardless of if it originates from a trusted source. If you wish to allow certain IPs to be included in this header, you will need to create a Transform Rule under Overview. *Please Note: This is by no means an exhaustive guide on using Cloudflare transform rules, however it's enough to configure this rule which should achieve a secure result. Please refer to the Cloudflare documentation on [transform rules](https://developers.cloudflare.com/rules/transform/) for more information._

##### Method 2 Steps

1. On the left sidebar, click `Rules`.
2. Click `Overview`.
3. Scroll down to `Request Header Transform Rules` and click `Create rule`.
4. Set the `Rule name` to something appropriate like `Remove X-Forwarded-For Header`.
5. Set the `Field` option in the `When incoming requests match` section to `IP Source Address`.
6. Set the `Operator` option in the `When incoming requests match` section to `does not equal`.
7. Set the `Value` option in the `When incoming requests match` section to any of the IP addresses you trust.
8. Set the `Then` section dropdown to `Remove`.
9. Set the `Then` section Header name to `X-Forwarded-For`.
10. Click `Deploy`.

{{< figure
src="cloudflare2.png"
alt="Image of Cloudflare dashboard with steps 1 to 3 labeled"
width="736"
caption="Steps 1 - 3: Image of Cloudflare dashboard with steps 1 to 3 labeled for method 2"
title="Steps 1 - 3: Image of Cloudflare dashboard with steps 1 to 3 labeled for method 2" >}}

{{< figure
src="cloudflare3.png"
alt="Image of Cloudflare dashboard with steps 4 to 10 labeled"
width="736"
caption="Steps 4 - 10: Image of Cloudflare dashboard with steps 4 to 10 labeled for method 2"
title="Steps 4 - 10: Image of Cloudflare dashboard with steps 4 to 10 labeled for method 2" >}}


Cloudflare publishes its IP address ranges publicly at the easy to remember address
[https://www.cloudflare.com/ips/](https://www.cloudflare.com/ips/). You should use this with the trusted proxies section
of your relevant proxy to ensure it's trusted if you intend to use Cloudflare.

[X-Forwarded-For]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-For
[Cloudflare]: https://www.cloudflare.com
